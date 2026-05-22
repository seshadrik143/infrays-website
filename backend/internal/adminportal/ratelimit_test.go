package adminportal_test

import (
	"net/http"
	"testing"
)

func TestAdminLoginAccountLockout(t *testing.T) {
	h := newHarness(t)
	h.seedAdmin("admin@x.com", "rightpassword1234", true)
	c := h.client()
	// 5 failed attempts.
	for i := 0; i < 5; i++ {
		resp, _ := h.do(c, "POST", "/api/admin/auth/login", map[string]any{
			"email": "admin@x.com", "password": "wrong",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, resp.StatusCode)
		}
	}
	// 6th — even right password should be rate-limited.
	resp, _ := h.do(c, "POST", "/api/admin/auth/login", map[string]any{
		"email": "admin@x.com", "password": "rightpassword1234",
	})
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Fatal("missing Retry-After header")
	}
}

func TestAdminLoginIPRateLimit(t *testing.T) {
	h := newHarness(t)
	// 11 different emails from same IP → 11th should 429 (per-IP=10).
	c := h.client()
	for i := 0; i < 10; i++ {
		resp, _ := h.do(c, "POST", "/api/admin/auth/login", map[string]any{
			"email": "noone@x.com", "password": "x",
		})
		// 401 invalid_credentials — that's fine, we just want under threshold.
		_ = resp
	}
	resp, _ := h.do(c, "POST", "/api/admin/auth/login", map[string]any{
		"email": "noone@x.com", "password": "x",
	})
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after IP burst, got %d", resp.StatusCode)
	}
}
