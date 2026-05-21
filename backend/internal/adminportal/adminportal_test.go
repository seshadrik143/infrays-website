package adminportal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/seshadrik143/infrays-website/backend/internal/adminportal"
	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

type harness struct {
	t        *testing.T
	srv      *httptest.Server
	store    store.Store
	auditLog audit.Log
	clock    time.Time
	clockMu  sync.Mutex
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	st := store.NewMemory()
	auditLog := audit.NewMemory()
	h := &harness{
		t: t, store: st, auditLog: auditLog,
		clock: time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC),
	}
	srv := adminportal.NewServer(adminportal.Config{
		Store: st, Audit: auditLog,
		Secure: false, TOTPIssuer: "infraYS-test",
		Now: h.now,
	})
	h.srv = httptest.NewServer(srv.Routes())
	t.Cleanup(h.srv.Close)
	t.Cleanup(func() { _ = st.Close(); _ = auditLog.Close() })
	return h
}

func (h *harness) now() time.Time {
	h.clockMu.Lock()
	defer h.clockMu.Unlock()
	return h.clock
}

func (h *harness) advance(d time.Duration) {
	h.clockMu.Lock()
	defer h.clockMu.Unlock()
	h.clock = h.clock.Add(d)
}

func (h *harness) client() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

func (h *harness) do(c *http.Client, method, path string, body any) (*http.Response, []byte) {
	h.t.Helper()
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, h.srv.URL+path, rdr)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		h.t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp, out
}

// seedAdmin creates an admin user via store + returns the ID.
// enrollMFA=true also sets MFAEnrolled with a known secret so tests
// can produce valid TOTP codes.
func (h *harness) seedAdmin(email, password string, enrollMFA bool) (id, secret string) {
	h.t.Helper()
	hash, err := adminportal.HashPassword(password)
	if err != nil {
		h.t.Fatal(err)
	}
	id = adminportal.NewAdminID()
	a := &store.AdminUser{
		ID: id, Email: email, PasswordHash: hash, Role: "admin",
		CreatedAt: h.now(),
	}
	if enrollMFA {
		key, _ := totp.Generate(totp.GenerateOpts{Issuer: "test", AccountName: email, SecretSize: 20})
		a.MFASecret = key.Secret()
		a.MFAEnrolled = true
		secret = key.Secret()
	}
	if err := h.store.CreateAdminUser(context.Background(), a); err != nil {
		h.t.Fatal(err)
	}
	return id, secret
}

func mustTOTP(t *testing.T, secret string, at time.Time) string {
	t.Helper()
	code, err := totp.GenerateCode(secret, at)
	if err != nil {
		t.Fatal(err)
	}
	return code
}

// loginAndVerifyMFA returns a fully-authenticated client (stage 2).
func (h *harness) loginAndVerifyMFA(email, password, secret string) *http.Client {
	h.t.Helper()
	c := h.client()
	resp, _ := h.do(c, "POST", "/api/admin/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if resp.StatusCode != 200 {
		h.t.Fatalf("login: %d", resp.StatusCode)
	}
	resp, _ = h.do(c, "POST", "/api/admin/auth/mfa-challenge", map[string]any{
		"code": mustTOTP(h.t, secret, h.now()),
	})
	if resp.StatusCode != 200 {
		h.t.Fatalf("mfa-challenge: %d", resp.StatusCode)
	}
	return c
}

// ─── Tests ──────────────────────────────────────────────────────────

func TestLoginWrongPassword(t *testing.T) {
	h := newHarness(t)
	h.seedAdmin("admin@x.com", "rightpass1234", true)
	c := h.client()
	resp, _ := h.do(c, "POST", "/api/admin/auth/login", map[string]any{
		"email": "admin@x.com", "password": "wrongpass",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLoginRequiresMFA(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "password1234", true)

	c := h.client()
	resp, _ := h.do(c, "POST", "/api/admin/auth/login", map[string]any{
		"email": "admin@x.com", "password": "password1234",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("login: %d", resp.StatusCode)
	}
	// Stage-2 route — should be blocked until MFA verifies.
	resp, _ = h.do(c, "GET", "/api/admin/customers", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 before mfa, got %d", resp.StatusCode)
	}
	// Stage-1 routes work.
	resp, _ = h.do(c, "GET", "/api/admin/auth/me", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("/me before mfa should be 200, got %d", resp.StatusCode)
	}
	// Verify MFA.
	resp, _ = h.do(c, "POST", "/api/admin/auth/mfa-challenge", map[string]any{
		"code": mustTOTP(t, secret, h.now()),
	})
	if resp.StatusCode != 200 {
		t.Fatalf("mfa: %d", resp.StatusCode)
	}
	// Stage-2 now works.
	resp, _ = h.do(c, "GET", "/api/admin/customers", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("/customers after mfa: %d", resp.StatusCode)
	}
}

func TestMFAEnrollmentFlow(t *testing.T) {
	h := newHarness(t)
	h.seedAdmin("newadmin@x.com", "password1234", false /* not enrolled */)
	c := h.client()
	// Login.
	resp, _ := h.do(c, "POST", "/api/admin/auth/login", map[string]any{
		"email": "newadmin@x.com", "password": "password1234",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("login: %d", resp.StatusCode)
	}
	// Setup MFA.
	resp, body := h.do(c, "POST", "/api/admin/auth/mfa-setup", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("setup: %d (%s)", resp.StatusCode, body)
	}
	var setup struct {
		Secret string `json:"secret"`
	}
	_ = json.Unmarshal(body, &setup)
	if setup.Secret == "" {
		t.Fatal("no secret returned")
	}
	// Verify.
	resp, _ = h.do(c, "POST", "/api/admin/auth/mfa-setup-verify", map[string]any{
		"code": mustTOTP(t, setup.Secret, h.now()),
	})
	if resp.StatusCode != 200 {
		t.Fatalf("setup-verify: %d", resp.StatusCode)
	}
	// Stage-2 should work now.
	resp, _ = h.do(c, "GET", "/api/admin/customers", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("post-enroll stage-2 fail: %d", resp.StatusCode)
	}
}

func TestSessionIdleExpiry(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "password1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "password1234", secret)
	h.advance(adminportal.SessionIdleTTL + time.Minute)
	resp, _ := h.do(c, "GET", "/api/admin/auth/me", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestListCustomersFilters(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "password1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "password1234", secret)
	now := h.now()
	for _, e := range []string{"alpha@a.com", "beta@a.com", "gamma@b.com"} {
		_ = h.store.CreateCustomer(context.Background(), &store.Customer{
			ID: "cust_" + e, Email: e, Status: "active", CreatedAt: now, UpdatedAt: now,
		})
	}
	// All.
	resp, body := h.do(c, "GET", "/api/admin/customers", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list: %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), "alpha@a.com") || !strings.Contains(string(body), "gamma@b.com") {
		t.Fatalf("missing customers: %s", body)
	}
	// Filter.
	resp, body = h.do(c, "GET", "/api/admin/customers?q=gamma", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list q: %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), "gamma@b.com") || strings.Contains(string(body), "alpha@a.com") {
		t.Fatalf("filter wrong: %s", body)
	}
}

func TestCustomerDetailAggregation(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "password1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "password1234", secret)
	now := h.now()
	_ = h.store.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust1", Email: "c@x.com", Status: "active", CreatedAt: now, UpdatedAt: now,
	})
	_ = h.store.CreateSubscription(context.Background(), &store.Subscription{
		ID: "sub1", CustomerID: "cust1", Tier: "professional", Status: "active",
		CurrentPeriodStart: now, CurrentPeriodEnd: now.Add(30 * 24 * time.Hour),
		CreatedAt: now, UpdatedAt: now,
	})
	_ = h.store.UpsertDeployment(context.Background(), &store.Deployment{
		ID: "dep1", CustomerID: "cust1", DeploymentID: "dep-1",
		FirstSeenAt: now, LastSeenAt: now, CreatedAt: now,
	})
	resp, body := h.do(c, "GET", "/api/admin/customers/cust1", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("detail: %d (%s)", resp.StatusCode, body)
	}
	for _, want := range []string{"sub1", "dep-1", "c@x.com"} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("missing %q in detail: %s", want, body)
		}
	}
}

func TestSuspendCustomerDropsPortalSessions(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "password1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "password1234", secret)
	now := h.now()
	_ = h.store.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust1", Email: "c@x.com", Status: "active", CreatedAt: now, UpdatedAt: now,
	})
	// Pretend the customer is logged into the portal.
	_ = h.store.CreatePortalSession(context.Background(), &store.PortalSession{
		ID: "psess1", CustomerID: "cust1",
		CreatedAt: now, LastSeen: now, ExpiresAt: now.Add(time.Hour),
	})
	resp, _ := h.do(c, "PATCH", "/api/admin/customers/cust1", map[string]any{"status": "suspended"})
	if resp.StatusCode != 200 {
		t.Fatalf("suspend: %d", resp.StatusCode)
	}
	if _, err := h.store.GetPortalSession(context.Background(), "psess1"); err == nil {
		t.Fatal("portal session should be deleted on suspend")
	}
}

func TestFlagDeployment(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "password1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "password1234", secret)
	now := h.now()
	_ = h.store.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust1", Email: "c@x.com", Status: "active", CreatedAt: now, UpdatedAt: now,
	})
	_ = h.store.UpsertDeployment(context.Background(), &store.Deployment{
		ID: "dep1", CustomerID: "cust1", DeploymentID: "dep-1",
		FirstSeenAt: now, LastSeenAt: now, CreatedAt: now,
	})
	resp, _ := h.do(c, "POST", "/api/admin/deployments/dep-1/flag", map[string]any{
		"flagged": true, "reason": "suspicious activity",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("flag: %d", resp.StatusCode)
	}
	d, _ := h.store.GetDeployment(context.Background(), "dep-1")
	if !d.FlaggedForReview || d.FlagReason != "suspicious activity" {
		t.Fatalf("flag not persisted: %+v", d)
	}
}

func TestAuditPagination(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "password1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "password1234", secret)
	// Login already wrote 2 entries (login_stage1, login_mfa_verified).
	resp, body := h.do(c, "GET", "/api/admin/audit?limit=10", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("audit: %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), "login_stage1") {
		t.Fatalf("missing login entry: %s", body)
	}
}

func TestUnauthenticated(t *testing.T) {
	h := newHarness(t)
	c := h.client()
	resp, _ := h.do(c, "GET", "/api/admin/customers", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestChangePasswordRotatesSessions(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "oldpassword1234", true)
	a := h.loginAndVerifyMFA("admin@x.com", "oldpassword1234", secret)
	b := h.loginAndVerifyMFA("admin@x.com", "oldpassword1234", secret)
	resp, _ := h.do(a, "POST", "/api/admin/auth/change-password", map[string]any{
		"old_password": "oldpassword1234", "new_password": "newpassword1234",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("change-password: %d", resp.StatusCode)
	}
	// A still works.
	resp, _ = h.do(a, "GET", "/api/admin/customers", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("A after change: %d", resp.StatusCode)
	}
	// B is dead.
	resp, _ = h.do(b, "GET", "/api/admin/customers", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("B should be 401, got %d", resp.StatusCode)
	}
}

func TestCreateOfflineSubscription(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "password1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "password1234", secret)
	now := h.now()
	_ = h.store.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust1", Email: "c@x.com", Status: "active", CreatedAt: now, UpdatedAt: now,
	})
	resp, body := h.do(c, "POST", "/api/admin/customers/cust1/subscriptions", map[string]any{
		"tier": "enterprise",
		"current_period_end": now.Add(365 * 24 * time.Hour),
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d (%s)", resp.StatusCode, body)
	}
	subs, _ := h.store.ListSubscriptionsByCustomer(context.Background(), "cust1")
	if len(subs) != 1 || !subs[0].ManualOffline || subs[0].Tier != "enterprise" {
		t.Fatalf("subs wrong: %+v", subs)
	}
}

func TestAdminCreateEnrollmentToken(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "password1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "password1234", secret)
	now := h.now()
	_ = h.store.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust1", Email: "c@x.com", Status: "active", CreatedAt: now, UpdatedAt: now,
	})
	_ = h.store.CreateSubscription(context.Background(), &store.Subscription{
		ID: "sub1", CustomerID: "cust1", Tier: "professional", Status: "active",
		CreatedAt: now, UpdatedAt: now,
	})
	resp, body := h.do(c, "POST", "/api/admin/customers/cust1/enrollment-tokens", map[string]any{
		"subscription_id": "sub1", "label": "prod", "ttl_hours": 24,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d (%s)", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "NP-ENROLL-") {
		t.Fatalf("missing plaintext: %s", body)
	}
}
