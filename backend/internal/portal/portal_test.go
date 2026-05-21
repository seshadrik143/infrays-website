package portal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/email"
	"github.com/seshadrik143/infrays-website/backend/internal/portal"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// ─── Fixtures ───────────────────────────────────────────────────────

type captureEmail struct {
	mu   sync.Mutex
	sent []email.Message
}

func (c *captureEmail) Send(_ context.Context, msg email.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, msg)
	return nil
}
func (c *captureEmail) Name() string { return "capture" }
func (c *captureEmail) last() email.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.sent) == 0 {
		return email.Message{}
	}
	return c.sent[len(c.sent)-1]
}

type fakeBilling struct {
	url string
	err error
}

func (f *fakeBilling) CreateSession(_ string) (string, error) { return f.url, f.err }

type harness struct {
	t          *testing.T
	srv        *httptest.Server
	store      store.Store
	auditLog   audit.Log
	email      *captureEmail
	billing    *fakeBilling
	clock      time.Time
	clockMutex sync.Mutex
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	st := store.NewMemory()
	auditLog := audit.NewMemory()
	em := &captureEmail{}
	bp := &fakeBilling{url: "https://stripe.example/portal"}
	h := &harness{
		t:        t,
		store:    st,
		auditLog: auditLog,
		email:    em,
		billing:  bp,
		clock:    time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC),
	}
	srv := portal.NewServer(portal.Config{
		Store:         st,
		Audit:         auditLog,
		Email:         em,
		BillingPortal: bp,
		AppURL:        "https://app.infrays.org",
		Secure:        false,
		Now:           h.now,
	})
	h.srv = httptest.NewServer(srv.Routes())
	t.Cleanup(h.srv.Close)
	t.Cleanup(func() {
		_ = st.Close()
		_ = auditLog.Close()
	})
	return h
}

func (h *harness) now() time.Time {
	h.clockMutex.Lock()
	defer h.clockMutex.Unlock()
	return h.clock
}

func (h *harness) advance(d time.Duration) {
	h.clockMutex.Lock()
	defer h.clockMutex.Unlock()
	h.clock = h.clock.Add(d)
}

func (h *harness) client() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

func (h *harness) do(client *http.Client, method, path string, body any) (*http.Response, []byte) {
	h.t.Helper()
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, h.srv.URL+path, rdr)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		h.t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp, out
}

// signupAndVerify is a helper that returns an authenticated client +
// the customer's verified row.
func (h *harness) signupAndVerify(email, password string) (*http.Client, *store.Customer) {
	h.t.Helper()
	c := h.client()
	resp, _ := h.do(c, "POST", "/api/portal/auth/signup", map[string]any{
		"email":    email,
		"password": password,
		"name":     "Test User",
	})
	if resp.StatusCode != 200 {
		h.t.Fatalf("signup: %d", resp.StatusCode)
	}
	cust, err := h.store.GetCustomerByEmail(context.Background(), email)
	if err != nil {
		h.t.Fatalf("lookup: %v", err)
	}
	// Verify-email: we need the plaintext token. The email body
	// contains "?token=<plain>" — extract from the captured email.
	msg := h.email.last()
	token := extractToken(h.t, msg.TextBody)
	resp, _ = h.do(c, "POST", "/api/portal/auth/verify-email", map[string]any{"token": token})
	if resp.StatusCode != 200 {
		h.t.Fatalf("verify: %d", resp.StatusCode)
	}
	// Login.
	resp, _ = h.do(c, "POST", "/api/portal/auth/login", map[string]any{
		"email":    email,
		"password": password,
	})
	if resp.StatusCode != 200 {
		h.t.Fatalf("login: %d", resp.StatusCode)
	}
	cust, _ = h.store.GetCustomerByEmail(context.Background(), email)
	return c, cust
}

func extractToken(t *testing.T, body string) string {
	t.Helper()
	idx := strings.Index(body, "?token=")
	if idx < 0 {
		t.Fatalf("no token in email body: %q", body)
	}
	rest := body[idx+len("?token="):]
	// Trim at first whitespace/newline.
	for i, r := range rest {
		if r == ' ' || r == '\n' || r == '\r' {
			return rest[:i]
		}
	}
	return rest
}

// ─── Tests ──────────────────────────────────────────────────────────

func TestSignupVerifyLoginMeFlow(t *testing.T) {
	h := newHarness(t)
	c, cust := h.signupAndVerify("alice@example.com", "supersecret")

	resp, body := h.do(c, "GET", "/api/portal/auth/me", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("/me: %d (%s)", resp.StatusCode, body)
	}
	var profile portal.CustomerProfile
	_ = json.Unmarshal(body, &profile)
	if profile.Email != "alice@example.com" || !profile.EmailVerified {
		t.Fatalf("profile mismatch: %+v", profile)
	}
	if profile.ID != cust.ID {
		t.Fatalf("id mismatch")
	}
}

func TestUnauthenticatedAccessDenied(t *testing.T) {
	h := newHarness(t)
	c := h.client()
	resp, _ := h.do(c, "GET", "/api/portal/auth/me", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	h := newHarness(t)
	h.signupAndVerify("bob@example.com", "rightpassword")
	c := h.client()
	resp, _ := h.do(c, "POST", "/api/portal/auth/login", map[string]any{
		"email": "bob@example.com", "password": "wrongpassword",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestSignupExistingClaimsAccount(t *testing.T) {
	h := newHarness(t)
	// Pre-create a customer via the store directly (simulates admin/Stripe path).
	_ = h.store.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust_existing", Email: "carol@example.com", Status: "active",
		CreatedAt: h.now(), UpdatedAt: h.now(),
	})
	c := h.client()
	resp, _ := h.do(c, "POST", "/api/portal/auth/signup", map[string]any{
		"email": "carol@example.com", "password": "newpassword",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("signup claim: %d", resp.StatusCode)
	}
	got, _ := h.store.GetCustomerByEmail(context.Background(), "carol@example.com")
	if got.ID != "cust_existing" {
		t.Fatalf("expected existing id, got %s", got.ID)
	}
	if got.PasswordHash == "" {
		t.Fatalf("password hash should be set after claim")
	}
	_ = c
}

func TestSignupExistingWithPasswordRejected(t *testing.T) {
	h := newHarness(t)
	h.signupAndVerify("dave@example.com", "mypass123")
	c := h.client()
	resp, _ := h.do(c, "POST", "/api/portal/auth/signup", map[string]any{
		"email": "dave@example.com", "password": "anotherpass",
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestVerifyEmailWrongToken(t *testing.T) {
	h := newHarness(t)
	c := h.client()
	resp, _ := h.do(c, "POST", "/api/portal/auth/signup", map[string]any{
		"email": "eve@example.com", "password": "password1",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("signup: %d", resp.StatusCode)
	}
	resp, _ = h.do(c, "POST", "/api/portal/auth/verify-email", map[string]any{
		"token": "wrong-token",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestChangePasswordRotatesOtherSessions(t *testing.T) {
	h := newHarness(t)
	clientA, _ := h.signupAndVerify("frank@example.com", "oldpassword")
	// Second login on a fresh client = second session.
	clientB := h.client()
	resp, _ := h.do(clientB, "POST", "/api/portal/auth/login", map[string]any{
		"email": "frank@example.com", "password": "oldpassword",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("login B: %d", resp.StatusCode)
	}
	// Change password on A.
	resp, _ = h.do(clientA, "POST", "/api/portal/auth/change-password", map[string]any{
		"old_password": "oldpassword", "new_password": "newpassword",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("change-password: %d", resp.StatusCode)
	}
	// A still works.
	resp, _ = h.do(clientA, "GET", "/api/portal/auth/me", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("A after change: %d", resp.StatusCode)
	}
	// B is invalidated.
	resp, _ = h.do(clientB, "GET", "/api/portal/auth/me", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("B should be 401, got %d", resp.StatusCode)
	}
}

func TestSessionExpiryRollover(t *testing.T) {
	h := newHarness(t)
	c, _ := h.signupAndVerify("gina@example.com", "password1")
	// Advance past TTL.
	h.advance(portal.SessionTTL + time.Hour)
	resp, _ := h.do(c, "GET", "/api/portal/auth/me", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 after expiry, got %d", resp.StatusCode)
	}
}

func TestCrossCustomerAccessDenied(t *testing.T) {
	h := newHarness(t)
	_, aliceC := h.signupAndVerify("alice@x.com", "password1")
	bobClient, _ := h.signupAndVerify("bob@x.com", "password2")

	// Seed: Alice owns a subscription + enrollment token + deployment.
	now := h.now()
	sub := &store.Subscription{
		ID: "sub_alice", CustomerID: aliceC.ID, Tier: "professional",
		Status: "active", CurrentPeriodStart: now, CurrentPeriodEnd: now.Add(30 * 24 * time.Hour),
		CreatedAt: now, UpdatedAt: now,
	}
	_ = h.store.CreateSubscription(context.Background(), sub)
	_ = h.store.CreateEnrollmentToken(context.Background(), &store.EnrollmentToken{
		ID: "etok_alice", CustomerID: aliceC.ID, SubscriptionID: sub.ID,
		TokenHash: "deadbeef", CreatedAt: now, ExpiresAt: now.Add(24 * time.Hour),
	})
	_ = h.store.UpsertDeployment(context.Background(), &store.Deployment{
		ID: "dep_alice", CustomerID: aliceC.ID, DeploymentID: "alice-prod",
		FirstSeenAt: now, LastSeenAt: now, CreatedAt: now,
	})
	_ = h.store.CreateLicense(context.Background(), &store.License{
		ID: "lic_alice", JTI: "jti1", LicenseID: "lid_alice", CustomerID: aliceC.ID,
		SubscriptionID: sub.ID, DeploymentID: "alice-prod", EntitlementSetID: "free-v1",
		Tier: "professional", Iat: now, NotBefore: now, ExpiresAt: now.Add(365 * 24 * time.Hour),
		GraceUntil: now.Add(456 * 24 * time.Hour), Kid: "kid", PayloadJWS: "alice.jws.payload",
		CreatedAt: now,
	})

	// Bob's listing must not include Alice's rows.
	resp, body := h.do(bobClient, "GET", "/api/portal/subscriptions", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("subs: %d (%s)", resp.StatusCode, body)
	}
	if strings.Contains(string(body), "sub_alice") {
		t.Fatalf("bob saw alice's subscription: %s", body)
	}
	resp, body = h.do(bobClient, "GET", "/api/portal/deployments", nil)
	if strings.Contains(string(body), "alice-prod") {
		t.Fatalf("bob saw alice's deployment: %s", body)
	}
	resp, body = h.do(bobClient, "GET", "/api/portal/enrollment-tokens", nil)
	if strings.Contains(string(body), "etok_alice") {
		t.Fatalf("bob saw alice's token: %s", body)
	}
	resp, body = h.do(bobClient, "GET", "/api/portal/licenses", nil)
	if strings.Contains(string(body), "lid_alice") {
		t.Fatalf("bob saw alice's license: %s", body)
	}
	// Bob can't download Alice's license.
	resp, _ = h.do(bobClient, "GET", "/api/portal/offline-license?license_id=lid_alice", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-customer license, got %d", resp.StatusCode)
	}
	// Bob can't revoke Alice's enrollment token.
	resp, _ = h.do(bobClient, "POST", "/api/portal/enrollment-tokens/revoke", map[string]any{
		"token_id": "etok_alice",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-customer revoke, got %d", resp.StatusCode)
	}
}

func TestCreateEnrollmentTokenOwnershipCheck(t *testing.T) {
	h := newHarness(t)
	_, aliceC := h.signupAndVerify("alice2@x.com", "password1")
	bobClient, _ := h.signupAndVerify("bob2@x.com", "password2")
	now := h.now()
	sub := &store.Subscription{
		ID: "sub_alice2", CustomerID: aliceC.ID, Tier: "professional",
		Status: "active", CreatedAt: now, UpdatedAt: now,
	}
	_ = h.store.CreateSubscription(context.Background(), sub)
	resp, _ := h.do(bobClient, "POST", "/api/portal/enrollment-tokens", map[string]any{
		"subscription_id": "sub_alice2",
		"label":           "stolen",
		"ttl_hours":       1,
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 cross-customer, got %d", resp.StatusCode)
	}
}

func TestCreateEnrollmentTokenSuccess(t *testing.T) {
	h := newHarness(t)
	cli, cust := h.signupAndVerify("hank@x.com", "password1")
	now := h.now()
	sub := &store.Subscription{
		ID: "sub_hank", CustomerID: cust.ID, Tier: "professional",
		Status: "active", CreatedAt: now, UpdatedAt: now,
	}
	_ = h.store.CreateSubscription(context.Background(), sub)
	resp, body := h.do(cli, "POST", "/api/portal/enrollment-tokens", map[string]any{
		"subscription_id": "sub_hank",
		"label":           "prod",
		"ttl_hours":       1,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d (%s)", resp.StatusCode, body)
	}
	var resp1 struct {
		ID        string `json:"id"`
		Plaintext string `json:"plaintext"`
	}
	_ = json.Unmarshal(body, &resp1)
	if !strings.HasPrefix(resp1.Plaintext, "NP-ENROLL-") {
		t.Fatalf("expected NP-ENROLL- prefix, got %q", resp1.Plaintext)
	}

	// List should show it.
	resp, body = h.do(cli, "GET", "/api/portal/enrollment-tokens", nil)
	if !strings.Contains(string(body), resp1.ID) {
		t.Fatalf("token not in list: %s", body)
	}

	// Revoke.
	resp, _ = h.do(cli, "POST", "/api/portal/enrollment-tokens/revoke", map[string]any{
		"token_id": resp1.ID,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("revoke: %d", resp.StatusCode)
	}
}

func TestEnrollmentTokenRequiresEmailVerification(t *testing.T) {
	h := newHarness(t)
	c := h.client()
	// Signup but DON'T verify.
	resp, _ := h.do(c, "POST", "/api/portal/auth/signup", map[string]any{
		"email": "irma@x.com", "password": "password1",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("signup: %d", resp.StatusCode)
	}
	// Login (no verification required to log in — we want to surface
	// the "verify your email" banner inside the app).
	resp, _ = h.do(c, "POST", "/api/portal/auth/login", map[string]any{
		"email": "irma@x.com", "password": "password1",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("login: %d", resp.StatusCode)
	}
	cust, _ := h.store.GetCustomerByEmail(context.Background(), "irma@x.com")
	now := h.now()
	_ = h.store.CreateSubscription(context.Background(), &store.Subscription{
		ID: "sub_irma", CustomerID: cust.ID, Status: "active", Tier: "free",
		CreatedAt: now, UpdatedAt: now,
	})
	resp, body := h.do(c, "POST", "/api/portal/enrollment-tokens", map[string]any{
		"subscription_id": "sub_irma", "label": "x", "ttl_hours": 1,
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 without email verified, got %d (%s)", resp.StatusCode, body)
	}
}

func TestPasswordResetFlow(t *testing.T) {
	h := newHarness(t)
	h.signupAndVerify("jane@x.com", "oldpass1234")
	// Request reset.
	c := h.client()
	resp, _ := h.do(c, "POST", "/api/portal/auth/request-password-reset", map[string]any{
		"email": "jane@x.com",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("request-reset: %d", resp.StatusCode)
	}
	token := extractToken(t, h.email.last().TextBody)

	// Reset.
	resp, _ = h.do(c, "POST", "/api/portal/auth/reset-password", map[string]any{
		"token": token, "new_password": "newpass1234",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("reset: %d", resp.StatusCode)
	}

	// New password works.
	resp, _ = h.do(c, "POST", "/api/portal/auth/login", map[string]any{
		"email": "jane@x.com", "password": "newpass1234",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("login after reset: %d", resp.StatusCode)
	}

	// Old password rejected.
	resp, _ = h.do(c, "POST", "/api/portal/auth/login", map[string]any{
		"email": "jane@x.com", "password": "oldpass1234",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("old password should fail, got %d", resp.StatusCode)
	}
}

func TestPasswordResetUnknownEmailDoesNotLeak(t *testing.T) {
	h := newHarness(t)
	c := h.client()
	resp, body := h.do(c, "POST", "/api/portal/auth/request-password-reset", map[string]any{
		"email": "nobody@x.com",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 (no leak), got %d (%s)", resp.StatusCode, body)
	}
}

func TestBillingPortalNoStripeCustomer(t *testing.T) {
	h := newHarness(t)
	c, _ := h.signupAndVerify("nostripe@x.com", "password1")
	resp, _ := h.do(c, "POST", "/api/portal/billing-portal-url", nil)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestBillingPortalWithStripeCustomer(t *testing.T) {
	h := newHarness(t)
	c, cust := h.signupAndVerify("stripe@x.com", "password1")
	cust.StripeCustomerID = "cus_test_123"
	if err := h.store.UpdateCustomer(context.Background(), cust); err != nil {
		t.Fatalf("update: %v", err)
	}
	resp, body := h.do(c, "POST", "/api/portal/billing-portal-url", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d (%s)", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "stripe.example/portal") {
		t.Fatalf("expected portal URL in body: %s", body)
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	h := newHarness(t)
	c, _ := h.signupAndVerify("kim@x.com", "password1")
	resp, _ := h.do(c, "POST", "/api/portal/auth/logout", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("logout: %d", resp.StatusCode)
	}
	resp, _ = h.do(c, "GET", "/api/portal/auth/me", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", resp.StatusCode)
	}
}

func TestUpdateAccount(t *testing.T) {
	h := newHarness(t)
	c, _ := h.signupAndVerify("luna@x.com", "password1")
	name := "Luna Lovegood"
	company := "Quibbler Inc."
	resp, body := h.do(c, "PATCH", "/api/portal/account", map[string]any{
		"name": name, "company": company,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update: %d (%s)", resp.StatusCode, body)
	}
	got, _ := h.store.GetCustomerByEmail(context.Background(), "luna@x.com")
	if got.Name != name || got.Company != company {
		t.Fatalf("not updated: %+v", got)
	}
}

func TestOfflineLicenseDownload(t *testing.T) {
	h := newHarness(t)
	c, cust := h.signupAndVerify("mark@x.com", "password1")
	now := h.now()
	_ = h.store.CreateLicense(context.Background(), &store.License{
		ID: "lic1", JTI: "j1", LicenseID: "lid_mark", CustomerID: cust.ID,
		SubscriptionID: "s1", DeploymentID: "d1", EntitlementSetID: "free-v1",
		Tier: "free", Iat: now, NotBefore: now, ExpiresAt: now.Add(365 * 24 * time.Hour),
		GraceUntil: now.Add(456 * 24 * time.Hour), Kid: "kid",
		PayloadJWS: "header.payload.sig", CreatedAt: now,
	})
	resp, body := h.do(c, "GET", "/api/portal/offline-license?license_id=lid_mark", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("download: %d", resp.StatusCode)
	}
	if string(body) != "header.payload.sig" {
		t.Fatalf("payload mismatch: %q", string(body))
	}
	if !strings.Contains(resp.Header.Get("Content-Disposition"), "lid_mark") {
		t.Fatalf("attachment header missing")
	}
}

// ─── Sanity: ensure portal sub-package compiles under -race ─────────

var _ error = errors.New("placeholder")
