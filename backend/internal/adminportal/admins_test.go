package adminportal_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/pquerna/otp/totp"
	"github.com/seshadrik143/infrays-website/backend/internal/adminportal"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

func TestAdminMgmtForbiddenForNonAdmins(t *testing.T) {
	h := newHarness(t)
	// Seed a support-role user with MFA enrolled.
	secret := seedAdminWithRole(t, h, "support@x.com", "supportpass1234", "support")
	c := h.loginAndVerifyMFA("support@x.com", "supportpass1234", secret)

	resp, body := h.do(c, "GET", "/api/admin/admins", nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("list: expected 403, got %d (%s)", resp.StatusCode, body)
	}
	resp, body = h.do(c, "POST", "/api/admin/admins", map[string]any{
		"email": "x@x.com", "role": "support",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("create: expected 403, got %d (%s)", resp.StatusCode, body)
	}
}

func TestCreateAdminReturnsTempPasswordOnce(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "rightpass1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "rightpass1234", secret)

	resp, body := h.do(c, "POST", "/api/admin/admins", map[string]any{
		"email": "newadmin@x.com", "role": "support",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d (%s)", resp.StatusCode, body)
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	temp, _ := got["temp_password"].(string)
	if temp == "" {
		t.Fatal("missing temp_password")
	}
	mcp, _ := got["must_change_password"].(bool)
	if !mcp {
		t.Fatal("must_change_password should be true")
	}

	// List should include the new admin.
	resp, body = h.do(c, "GET", "/api/admin/admins", nil)
	if !strings.Contains(string(body), "newadmin@x.com") {
		t.Fatalf("list missing new admin: %s", body)
	}

	// New admin can log in with the temp password.
	nc := h.client()
	resp, _ = h.do(nc, "POST", "/api/admin/auth/login", map[string]any{
		"email": "newadmin@x.com", "password": temp,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("new admin login: %d", resp.StatusCode)
	}
}

func TestDuplicateAdminConflict(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "rightpass1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "rightpass1234", secret)
	resp, _ := h.do(c, "POST", "/api/admin/admins", map[string]any{
		"email": "dup@x.com", "role": "support",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first create: %d", resp.StatusCode)
	}
	resp, _ = h.do(c, "POST", "/api/admin/admins", map[string]any{
		"email": "dup@x.com", "role": "support",
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestDisableAdminDropsTheirSessions(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "adminpass1234", true)
	adminClient := h.loginAndVerifyMFA("admin@x.com", "adminpass1234", secret)

	// Create + log in a second admin.
	resp, body := h.do(adminClient, "POST", "/api/admin/admins", map[string]any{
		"email": "victim@x.com", "role": "admin",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d (%s)", resp.StatusCode, body)
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	victimID, _ := got["id"].(string)
	temp, _ := got["temp_password"].(string)

	// Victim logs in (stage 1 — no MFA enrolled yet).
	victimClient := h.client()
	resp, _ = h.do(victimClient, "POST", "/api/admin/auth/login", map[string]any{
		"email": "victim@x.com", "password": temp,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("victim login: %d", resp.StatusCode)
	}
	// /me works.
	resp, _ = h.do(victimClient, "GET", "/api/admin/auth/me", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("victim /me before disable: %d", resp.StatusCode)
	}

	// Admin disables the victim.
	resp, _ = h.do(adminClient, "POST", "/api/admin/admins/"+victimID+"/status", map[string]any{
		"status": "disabled",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("disable: %d", resp.StatusCode)
	}

	// Victim's session is gone.
	resp, _ = h.do(victimClient, "GET", "/api/admin/auth/me", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("victim /me after disable: expected 401, got %d", resp.StatusCode)
	}

	// Victim can't log in either.
	resp, _ = h.do(h.client(), "POST", "/api/admin/auth/login", map[string]any{
		"email": "victim@x.com", "password": temp,
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("disabled login expected 403, got %d", resp.StatusCode)
	}
}

func TestCannotSelfDisable(t *testing.T) {
	h := newHarness(t)
	id, secret := h.seedAdmin("admin@x.com", "rightpass1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "rightpass1234", secret)
	resp, _ := h.do(c, "POST", "/api/admin/admins/"+id+"/status", map[string]any{"status": "disabled"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCannotSelfDelete(t *testing.T) {
	h := newHarness(t)
	id, secret := h.seedAdmin("admin@x.com", "rightpass1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "rightpass1234", secret)
	resp, _ := h.do(c, "DELETE", "/api/admin/admins/"+id, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestResetPasswordReturnsNewTempAndDropsSessions(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "rightpass1234", true)
	adminClient := h.loginAndVerifyMFA("admin@x.com", "rightpass1234", secret)

	resp, body := h.do(adminClient, "POST", "/api/admin/admins", map[string]any{
		"email": "user@x.com", "role": "support",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d (%s)", resp.StatusCode, body)
	}
	var created map[string]any
	_ = json.Unmarshal(body, &created)
	uid := created["id"].(string)
	oldTemp := created["temp_password"].(string)

	// Reset password.
	resp, body = h.do(adminClient, "POST", "/api/admin/admins/"+uid+"/reset-password", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("reset: %d", resp.StatusCode)
	}
	var rst map[string]any
	_ = json.Unmarshal(body, &rst)
	newTemp := rst["temp_password"].(string)
	if newTemp == oldTemp || newTemp == "" {
		t.Fatal("expected new distinct temp password")
	}
	// Old temp no longer valid.
	resp, _ = h.do(h.client(), "POST", "/api/admin/auth/login", map[string]any{
		"email": "user@x.com", "password": oldTemp,
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("old temp expected 401, got %d", resp.StatusCode)
	}
	// New temp works.
	resp, _ = h.do(h.client(), "POST", "/api/admin/auth/login", map[string]any{
		"email": "user@x.com", "password": newTemp,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("new temp login: %d", resp.StatusCode)
	}
}

// seedAdminWithRole creates an MFA-enrolled admin user with the given
// role. Returns the TOTP secret so the caller can drive the MFA flow.
func seedAdminWithRole(t *testing.T, h *harness, email, password, role string) string {
	t.Helper()
	hash, err := adminportal.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "test", AccountName: email, SecretSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	a := &store.AdminUser{
		ID: adminportal.NewAdminID(), Email: email, PasswordHash: hash, Role: role,
		Status: "active", MFASecret: key.Secret(), MFAEnrolled: true,
		CreatedAt: h.now(),
	}
	if err := h.store.CreateAdminUser(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	return key.Secret()
}
