// Package portal serves the customer self-service portal at
// app.infrays.org. Cookie-authenticated. Customers manage their
// subscriptions, deployments, enrollment tokens, and account here.
//
// Routes are registered by Routes() on a fresh ServeMux. The issuer
// main wires the mux as a sub-handler so /api/portal/* and the SPA
// fallback at / share one Go binary.
//
// Auth model: simple session cookie (np_portal_session), 7-day TTL,
// SameSite=Lax, HttpOnly, Secure when behind TLS. No refresh tokens,
// no remember-me. Activity rolls the TTL forward.
//
// Password storage: bcrypt cost 12. Email verification + password
// reset use single-use SHA-256 tokens (plaintext emailed once, hash
// stored on the customer row).
package portal

import (
	"net/http"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/email"
	"github.com/seshadrik143/infrays-website/backend/internal/ratelimit"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

const (
	SessionCookieName = "np_portal_session"
	SessionTTL        = 7 * 24 * time.Hour
	TokenTTL          = 24 * time.Hour // verify-email / reset-password
)

// BillingPortalCreator is the narrow surface the portal needs from
// the Stripe billing-portal session creator. Kept here as an interface
// so the portal package doesn't import stripe-go directly.
type BillingPortalCreator interface {
	CreateSession(stripeCustomerID string) (string, error)
}

// CheckoutCreator is the narrow surface the portal needs to start a
// self-serve Stripe Checkout for an authenticated customer. The tier +
// interval are resolved to a Stripe Price ID by the implementation
// (the client never supplies a raw Price ID). Kept as an interface so
// the portal package doesn't import stripe-go directly.
type CheckoutCreator interface {
	CreateCheckoutSession(tier, interval, customerEmail, customerID string) (url string, err error)
}

// Config holds the portal's wiring. AppURL is the public origin used
// to build verification / reset links inside email templates.
type Config struct {
	Store         store.Store
	Audit         audit.Log
	Email         email.Sender
	BillingPortal BillingPortalCreator // optional — nil disables the route
	Checkout      CheckoutCreator      // optional — nil disables self-serve checkout
	AppURL        string
	Secure        bool             // emit Secure cookies (true behind TLS)
	Now           func() time.Time // injectable for tests
}

// Server is the portal HTTP handler container.
type Server struct {
	cfg Config

	// Per-IP rate limiters for auth-adjacent endpoints. Tight on login
	// + reset; looser on signup (legit traffic isn't bursty there
	// either, but we want to avoid annoying false-positives if a
	// shared NAT bursts a few new signups).
	loginIPRL *Limiter
	signupRL  *Limiter
	resetRL   *Limiter
	verifyRL  *Limiter
	// Per-account lockout for customer logins: 5 fails within 15min
	// → reject for 15min from the most recent miss.
	loginAccountRL *Limiter
}

// Limiter is a thin alias so handler code doesn't import ratelimit
// directly — keeps the dependency surface small.
type Limiter = ratelimit.Limiter

func NewServer(cfg Config) *Server {
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	if cfg.AppURL == "" {
		cfg.AppURL = "https://app.infrays.org"
	}
	return &Server{
		cfg: cfg,
		// Customer-facing endpoints. Bursts allowed because legitimate
		// users sometimes mistype passwords; cliff-edge after the
		// budget is exhausted.
		loginIPRL:      ratelimit.New(ratelimit.Config{Max: 10, Window: 5 * time.Minute, Now: cfg.Now}),
		signupRL:       ratelimit.New(ratelimit.Config{Max: 5, Window: 10 * time.Minute, Now: cfg.Now}),
		resetRL:        ratelimit.New(ratelimit.Config{Max: 5, Window: 15 * time.Minute, Now: cfg.Now}),
		verifyRL:       ratelimit.New(ratelimit.Config{Max: 10, Window: 5 * time.Minute, Now: cfg.Now}),
		loginAccountRL: ratelimit.New(ratelimit.Config{Max: 5, Window: 15 * time.Minute, Now: cfg.Now}),
	}
}

// Close stops the rate-limiter cleanup goroutines. Tests should call
// it; in production the process exits and goroutines die anyway.
func (s *Server) Close() {
	s.loginIPRL.Close()
	s.signupRL.Close()
	s.resetRL.Close()
	s.verifyRL.Close()
	s.loginAccountRL.Close()
}

// Routes returns the portal mux. All routes mounted under /api/portal
// (the issuer main strips the prefix before delegating).
//
// Public routes — no session required:
//
//	POST /api/portal/auth/signup
//	POST /api/portal/auth/login
//	POST /api/portal/auth/verify-email
//	POST /api/portal/auth/request-password-reset
//	POST /api/portal/auth/reset-password
//
// Authenticated routes — session cookie required:
//
//	POST /api/portal/auth/logout
//	GET  /api/portal/auth/me
//	POST /api/portal/auth/change-password
//	POST /api/portal/auth/resend-verification
//
// Data routes added in Task #85.
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/portal/auth/signup", s.withRateLimit(s.signupRL, "portal.signup", s.handleSignup))
	// Login uses per-IP limit at the wrapper plus per-account lockout
	// inside the handler (handler also resets the account counter on
	// success).
	mux.HandleFunc("POST /api/portal/auth/login", s.withRateLimit(s.loginIPRL, "portal.login.ip", s.handleLogin))
	mux.HandleFunc("POST /api/portal/auth/verify-email", s.withRateLimit(s.verifyRL, "portal.verify", s.handleVerifyEmail))
	mux.HandleFunc("POST /api/portal/auth/request-password-reset", s.withRateLimit(s.resetRL, "portal.reset.request", s.handleRequestPasswordReset))
	mux.HandleFunc("POST /api/portal/auth/reset-password", s.withRateLimit(s.resetRL, "portal.reset.confirm", s.handleResetPassword))

	mux.HandleFunc("POST /api/portal/auth/logout", s.requireSession(s.handleLogout))
	mux.HandleFunc("GET /api/portal/auth/me", s.requireSession(s.handleMe))
	mux.HandleFunc("POST /api/portal/auth/change-password", s.requireSession(s.handleChangePassword))
	mux.HandleFunc("POST /api/portal/auth/resend-verification", s.requireSession(s.handleResendVerification))

	// Data routes — all session-protected.
	mux.HandleFunc("GET /api/portal/subscriptions", s.requireSession(s.handleListSubscriptions))
	mux.HandleFunc("GET /api/portal/deployments", s.requireSession(s.handleListDeployments))
	mux.HandleFunc("GET /api/portal/enrollment-tokens", s.requireSession(s.handleListEnrollmentTokens))
	mux.HandleFunc("POST /api/portal/enrollment-tokens", s.requireSession(s.handleCreateEnrollmentToken))
	mux.HandleFunc("POST /api/portal/enrollment-tokens/revoke", s.requireSession(s.handleRevokeEnrollmentToken))
	mux.HandleFunc("GET /api/portal/licenses", s.requireSession(s.handleListLicenses))
	mux.HandleFunc("GET /api/portal/offline-license", s.requireSession(s.handleOfflineLicense))
	mux.HandleFunc("PATCH /api/portal/account", s.requireSession(s.handleUpdateAccount))

	// Stripe billing portal redirect — only when configured.
	if s.cfg.BillingPortal != nil {
		mux.HandleFunc("POST /api/portal/billing-portal-url", s.requireSession(s.handleBillingPortalURL))
	}

	// Self-serve checkout — only when Stripe checkout is configured.
	if s.cfg.Checkout != nil {
		mux.HandleFunc("POST /api/portal/checkout-session", s.requireSession(s.handleCreateCheckoutSession))
	}

	return mux
}
