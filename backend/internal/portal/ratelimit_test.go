package portal_test

import (
	"net/http"
	"testing"
)

// These tests are kept in the existing harness via portal_test.go's
// newHarness; they sit in a separate file so the rate-limit coverage
// is easy to find.

func TestSignupRateLimitedPerIP(t *testing.T) {
	h := newHarness(t)
	c := h.client()
	// signupRL default: 5 per 10 minutes. 6th request from the same
	// client (httptest server → RemoteAddr is 127.0.0.1) should 429.
	for i := 0; i < 5; i++ {
		resp, _ := h.do(c, "POST", "/api/portal/auth/signup", map[string]any{
			"email":    "rl" + string(rune('a'+i)) + "@example.com",
			"password": "password1",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("attempt %d should succeed, got %d", i+1, resp.StatusCode)
		}
	}
	resp, _ := h.do(c, "POST", "/api/portal/auth/signup", map[string]any{
		"email":    "rlOverflow@example.com",
		"password": "password1",
	})
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after burst, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Fatal("missing Retry-After header on 429")
	}
}

func TestLoginAccountLockout(t *testing.T) {
	h := newHarness(t)
	// Seed an account.
	h.signupAndVerify("locked@example.com", "rightpassword1234")
	c := h.client()
	// 5 wrong-password attempts → 5 invalid_credentials responses.
	for i := 0; i < 5; i++ {
		resp, _ := h.do(c, "POST", "/api/portal/auth/login", map[string]any{
			"email":    "locked@example.com",
			"password": "wrongpw",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("attempt %d expected 401, got %d", i+1, resp.StatusCode)
		}
	}
	// 6th attempt: even with the RIGHT password we should get 429.
	resp, _ := h.do(c, "POST", "/api/portal/auth/login", map[string]any{
		"email":    "locked@example.com",
		"password": "rightpassword1234",
	})
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 lockout, got %d", resp.StatusCode)
	}
}

func TestLoginAccountLockoutResetsOnSuccess(t *testing.T) {
	h := newHarness(t)
	h.signupAndVerify("reset@example.com", "rightpassword1234")
	c := h.client()
	// 2 wrong attempts (below the 5 threshold).
	for i := 0; i < 2; i++ {
		_, _ = h.do(c, "POST", "/api/portal/auth/login", map[string]any{
			"email":    "reset@example.com",
			"password": "wrongpw",
		})
	}
	// Now a successful login.
	resp, _ := h.do(c, "POST", "/api/portal/auth/login", map[string]any{
		"email":    "reset@example.com",
		"password": "rightpassword1234",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("success login: %d", resp.StatusCode)
	}
	// Counter should be reset — 5 more failures must all return 401,
	// not flip to 429 on the (now-stale) earlier 2.
	for i := 0; i < 5; i++ {
		resp, _ := h.do(c, "POST", "/api/portal/auth/login", map[string]any{
			"email":    "reset@example.com",
			"password": "wrongpw",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("post-reset attempt %d should be 401, got %d", i+1, resp.StatusCode)
		}
	}
}
