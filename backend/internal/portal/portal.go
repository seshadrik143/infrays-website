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

// Config holds the portal's wiring. AppURL is the public origin used
// to build verification / reset links inside email templates.
type Config struct {
	Store         store.Store
	Audit         audit.Log
	Email         email.Sender
	BillingPortal BillingPortalCreator // optional — nil disables the route
	AppURL        string
	Secure        bool             // emit Secure cookies (true behind TLS)
	Now           func() time.Time // injectable for tests
}

// Server is the portal HTTP handler container.
type Server struct {
	cfg Config
}

func NewServer(cfg Config) *Server {
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	if cfg.AppURL == "" {
		cfg.AppURL = "https://app.infrays.org"
	}
	return &Server{cfg: cfg}
}

// Routes returns the portal mux. All routes mounted under /api/portal
// (the issuer main strips the prefix before delegating).
//
// Public routes — no session required:
//   POST /api/portal/auth/signup
//   POST /api/portal/auth/login
//   POST /api/portal/auth/verify-email
//   POST /api/portal/auth/request-password-reset
//   POST /api/portal/auth/reset-password
//
// Authenticated routes — session cookie required:
//   POST /api/portal/auth/logout
//   GET  /api/portal/auth/me
//   POST /api/portal/auth/change-password
//   POST /api/portal/auth/resend-verification
//
// Data routes added in Task #85.
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/portal/auth/signup", s.handleSignup)
	mux.HandleFunc("POST /api/portal/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/portal/auth/verify-email", s.handleVerifyEmail)
	mux.HandleFunc("POST /api/portal/auth/request-password-reset", s.handleRequestPasswordReset)
	mux.HandleFunc("POST /api/portal/auth/reset-password", s.handleResetPassword)

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

	return mux
}
