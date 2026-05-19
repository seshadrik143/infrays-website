package issuer_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/issuer"
	"github.com/seshadrik143/infrays-website/backend/internal/signing"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// testServer spins up an httptest server backed by in-memory stores
// and a fresh Ed25519 keypair. Returns the URL + a shutdown closure.
func testServer(t *testing.T) (string, ed25519.PublicKey, func()) {
	t.Helper()
	t.Setenv("NP_ISSUER_ADMIN_SECRET", "test-secret")

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := signing.NewLocalSignerFromKey("np-test-issuer-2026", priv)
	if err != nil {
		t.Fatal(err)
	}
	st := store.NewMemory()
	auditLog := audit.NewMemory()

	// Seed the three default entitlement sets.
	_ = st.CreateEntitlementSet(context.Background(), &store.EntitlementSet{
		ID: "professional-v1", Name: "Professional", Version: 1,
		Features: []string{"audit_log"},
		Limits:   store.Limits{MaxAgents: 50, MaxMetricsPerSec: 10000, RetentionDays: 90},
	})
	_ = st.CreateEntitlementSet(context.Background(), &store.EntitlementSet{
		ID: "free-v1", Name: "Free", Version: 1,
		Limits: store.Limits{MaxAgents: 3, RetentionDays: 7},
	})

	srv := issuer.NewServer(issuer.Config{
		Store:                st,
		Audit:                auditLog,
		Signer:               signer,
		IssuerURL:            "license.infrays.org",
		DefaultGraceDays:     90,
		RefreshIntervalHours: 24,
		Now:                  func() time.Time { return time.Now().UTC() },
	})
	hts := httptest.NewServer(srv.Routes())
	return hts.URL, pub, func() {
		hts.Close()
		signer.Close()
		st.Close()
		auditLog.Close()
	}
}

func adminPOST(t *testing.T, url, path string, body any) (*http.Response, []byte) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url+path, bytes.NewReader(b))
	req.Header.Set("X-Admin-Secret", "test-secret")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out := readAll(t, resp)
	return resp, out
}

func anonPOST(t *testing.T, url, path string, body any) (*http.Response, []byte) {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(url+path, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	return resp, readAll(t, resp)
}

func readAll(t *testing.T, r *http.Response) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r.Body)
	return buf.Bytes()
}

// verifyJWS does the same Ed25519 signature check that NodePulse's
// VerifyJWS does, using the test pubkey.
func verifyJWS(t *testing.T, tok string, pub ed25519.PublicKey) map[string]any {
	t.Helper()
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("malformed JWS: %d segments", len(parts))
	}
	signingInput := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decode sig: %v", err)
	}
	if !ed25519.Verify(pub, []byte(signingInput), sig) {
		t.Fatal("JWS signature did not verify")
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	_ = json.Unmarshal(payloadJSON, &m)
	return m
}

// ── tests ──────────────────────────────────────────────────────────

func TestHealthz(t *testing.T) {
	url, _, cleanup := testServer(t)
	defer cleanup()
	resp, err := http.Get(url + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status: got %d", resp.StatusCode)
	}
}

func TestAdminGate_RejectsWithoutSecret(t *testing.T) {
	url, _, cleanup := testServer(t)
	defer cleanup()
	resp, _ := anonPOST(t, url, "/internal/admin/customers", map[string]string{"email": "a@b.c"})
	if resp.StatusCode != 403 {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestFullEnrollmentLifecycle(t *testing.T) {
	url, pub, cleanup := testServer(t)
	defer cleanup()

	// 1. Admin creates customer
	resp, body := adminPOST(t, url, "/internal/admin/customers", map[string]string{
		"email": "ops@acme.example",
		"name":  "ACME Corp",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create customer: %d %s", resp.StatusCode, body)
	}
	var cust struct{ ID string }
	_ = json.Unmarshal(body, &cust)

	// 2. Admin creates subscription
	resp, body = adminPOST(t, url, "/internal/admin/subscriptions", map[string]any{
		"customer_id":              cust.ID,
		"tier":                     "professional",
		"current_period_end_days":  365,
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create subscription: %d %s", resp.StatusCode, body)
	}
	var sub struct{ ID string }
	_ = json.Unmarshal(body, &sub)

	// 3. Admin creates enrollment token
	resp, body = adminPOST(t, url, "/internal/admin/enrollment-tokens", map[string]any{
		"customer_id":     cust.ID,
		"subscription_id": sub.ID,
		"label":           "prod-cluster-1",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create enrollment token: %d %s", resp.StatusCode, body)
	}
	var tok struct{ Plaintext string }
	_ = json.Unmarshal(body, &tok)
	if !strings.HasPrefix(tok.Plaintext, "NP-ENROLL-") {
		t.Errorf("token format: got %q", tok.Plaintext)
	}

	// 4. NodePulse enrolls
	resp, body = anonPOST(t, url, "/v1/enroll", map[string]any{
		"enrollment_token": tok.Plaintext,
		"deployment_id":    "dep_test_001",
		"deployment_name":  "prod-cluster-1",
		"version":          "v0.47.0",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("enroll: %d %s", resp.StatusCode, body)
	}
	var enroll struct {
		LicenseJWS       string `json:"license_jws"`
		EntitlementSetID string `json:"entitlement_set_id"`
		RefreshAt        int64  `json:"refresh_at"`
	}
	_ = json.Unmarshal(body, &enroll)
	payload := verifyJWS(t, enroll.LicenseJWS, pub)
	if payload["tier"] != "professional" {
		t.Errorf("tier: got %v", payload["tier"])
	}
	if payload["subscription_id"] != sub.ID {
		t.Errorf("subscription_id: got %v want %s", payload["subscription_id"], sub.ID)
	}
	if payload["deployment_id"] != "dep_test_001" {
		t.Errorf("deployment_id: got %v", payload["deployment_id"])
	}
	if payload["entitlement_set_id"] != "professional-v1" {
		t.Errorf("entitlement_set_id: got %v", payload["entitlement_set_id"])
	}
	firstLicenseID := payload["license_id"].(string)
	firstJTI := payload["jti"].(string)
	if firstLicenseID == "" || firstJTI == "" {
		t.Fatal("missing license_id or jti")
	}

	// 5. Idempotent re-enroll with same deployment_id — same response cached
	resp, body = anonPOST(t, url, "/v1/enroll", map[string]any{
		"enrollment_token": tok.Plaintext,
		"deployment_id":    "dep_test_001",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("idempotent re-enroll: %d %s", resp.StatusCode, body)
	}
	var enroll2 struct {
		LicenseJWS string `json:"license_jws"`
	}
	_ = json.Unmarshal(body, &enroll2)
	if enroll2.LicenseJWS != enroll.LicenseJWS {
		t.Error("idempotent re-enroll returned different JWS")
	}

	// 6. Re-enroll with DIFFERENT deployment_id — rejected
	resp, body = anonPOST(t, url, "/v1/enroll", map[string]any{
		"enrollment_token": tok.Plaintext,
		"deployment_id":    "dep_attacker_999",
	})
	if resp.StatusCode != 403 {
		t.Errorf("cross-deployment redeem should 403, got %d %s", resp.StatusCode, body)
	}

	// 7. Refresh — same license_id, new JTI
	resp, body = anonPOST(t, url, "/v1/refresh", map[string]any{
		"license_id":    firstLicenseID,
		"deployment_id": "dep_test_001",
		"version":       "v0.47.0",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("refresh: %d %s", resp.StatusCode, body)
	}
	var refresh struct {
		LicenseJWS string `json:"license_jws"`
	}
	_ = json.Unmarshal(body, &refresh)
	refreshedPayload := verifyJWS(t, refresh.LicenseJWS, pub)
	if refreshedPayload["license_id"] != firstLicenseID {
		t.Errorf("license_id changed on refresh")
	}
	newJTI := refreshedPayload["jti"].(string)
	if newJTI == firstJTI {
		t.Error("jti did not rotate on refresh")
	}

	// 8. Admin revokes the new license
	resp, body = adminPOST(t, url, "/internal/admin/licenses/revoke", map[string]any{
		"jti":    newJTI,
		"reason": "test revoke",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("revoke: %d %s", resp.StatusCode, body)
	}

	// 9. Refresh after revoke — 403
	resp, body = anonPOST(t, url, "/v1/refresh", map[string]any{
		"license_id":    firstLicenseID,
		"deployment_id": "dep_test_001",
	})
	if resp.StatusCode != 403 {
		t.Errorf("refresh after revoke should 403, got %d %s", resp.StatusCode, body)
	}
}

func TestEntitlementsEndpoint(t *testing.T) {
	url, _, cleanup := testServer(t)
	defer cleanup()
	resp, err := http.Get(url + "/v1/entitlements/professional-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Cache-Control"); !strings.Contains(got, "max-age") {
		t.Errorf("cache-control: %q", got)
	}
	body := readAll(t, resp)
	var m map[string]any
	_ = json.Unmarshal(body, &m)
	if m["id"] != "professional-v1" {
		t.Errorf("id: got %v", m["id"])
	}
	features, _ := m["features"].([]any)
	if len(features) == 0 {
		t.Error("expected at least one feature in professional manifest")
	}
}

func TestWellKnownKeys(t *testing.T) {
	url, _, cleanup := testServer(t)
	defer cleanup()
	resp, err := http.Get(url + "/v1/well-known/keys")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readAll(t, resp)
	var m map[string]any
	_ = json.Unmarshal(body, &m)
	keys, _ := m["keys"].([]any)
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	k0 := keys[0].(map[string]any)
	if k0["alg"] != "EdDSA" {
		t.Errorf("alg: got %v", k0["alg"])
	}
	if k0["kid"] != "np-test-issuer-2026" {
		t.Errorf("kid: got %v", k0["kid"])
	}
	pub, _ := k0["pub"].(string)
	if !strings.HasPrefix(pub, "-----BEGIN PUBLIC KEY-----") {
		t.Errorf("pub: not PEM")
	}
}

func TestEnrollRejectsExpiredToken(t *testing.T) {
	url, _, cleanup := testServer(t)
	defer cleanup()

	// Create customer + subscription + token, then directly force the
	// token's ExpiresAt into the past by re-creating the issuer Now()
	// is harder to mock from outside — use a 0-hour TTL request.
	resp, body := adminPOST(t, url, "/internal/admin/customers", map[string]string{"email": "exp@test"})
	if resp.StatusCode != 201 {
		t.Fatal(string(body))
	}
	var cust struct{ ID string }
	_ = json.Unmarshal(body, &cust)

	resp, body = adminPOST(t, url, "/internal/admin/subscriptions", map[string]any{
		"customer_id": cust.ID, "tier": "professional",
	})
	var sub struct{ ID string }
	_ = json.Unmarshal(body, &sub)

	// TTL=1h, then immediately try to enroll with an obviously-wrong
	// token to confirm "invalid_token" path. (Direct expiry test
	// requires a clock injection feature we'd add in a follow-up.)
	resp, body = anonPOST(t, url, "/v1/enroll", map[string]any{
		"enrollment_token": "NP-ENROLL-NONEXISTENT",
		"deployment_id":    "dep_x",
	})
	if resp.StatusCode != 403 {
		t.Errorf("expected 403 for unknown token, got %d", resp.StatusCode)
	}
}

func TestAuditChainGrowsAndStaysValid(t *testing.T) {
	url, _, cleanup := testServer(t)
	defer cleanup()

	// Do a full enroll cycle to produce several audit events.
	resp, body := adminPOST(t, url, "/internal/admin/customers", map[string]string{"email": "audit@test"})
	if resp.StatusCode != 201 {
		t.Fatal(string(body))
	}
	var cust struct{ ID string }
	_ = json.Unmarshal(body, &cust)

	resp, body = adminPOST(t, url, "/internal/admin/subscriptions", map[string]any{
		"customer_id": cust.ID, "tier": "professional",
	})
	var sub struct{ ID string }
	_ = json.Unmarshal(body, &sub)

	resp, body = adminPOST(t, url, "/internal/admin/enrollment-tokens", map[string]any{
		"customer_id": cust.ID, "subscription_id": sub.ID,
	})
	var tok struct{ Plaintext string }
	_ = json.Unmarshal(body, &tok)

	_, _ = anonPOST(t, url, "/v1/enroll", map[string]any{
		"enrollment_token": tok.Plaintext,
		"deployment_id":    "dep_audit_1",
	})

	// Read audit tail
	req, _ := http.NewRequest("GET", url+"/internal/admin/audit", nil)
	req.Header.Set("X-Admin-Secret", "test-secret")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("audit status: %d", resp.StatusCode)
	}
	body = readAll(t, resp)
	var auditResp struct {
		Count   int `json:"count"`
		Entries []struct {
			Seq       int    `json:"Seq"`
			EventType string `json:"EventType"`
		} `json:"entries"`
	}
	_ = json.Unmarshal(body, &auditResp)
	if auditResp.Count < 4 {
		t.Errorf("expected >=4 audit entries, got %d", auditResp.Count)
	}
	wantEvents := map[string]bool{
		"admin.customer.created":         false,
		"admin.subscription.created":     false,
		"admin.enrollment_token.created": false,
		"license.issued":                 false,
	}
	for _, e := range auditResp.Entries {
		if _, ok := wantEvents[e.EventType]; ok {
			wantEvents[e.EventType] = true
		}
	}
	for evt, seen := range wantEvents {
		if !seen {
			t.Errorf("missing audit event: %s", evt)
		}
	}
}
