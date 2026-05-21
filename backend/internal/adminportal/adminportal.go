// Package adminportal serves the infraYS staff portal at
// /admin/* (mounted on the same Go binary as the issuer + customer
// portal). Cookie-authenticated with TOTP MFA enforcement.
//
// Auth model:
//   - Stage 1: POST /api/admin/auth/login (email + password) → 200 with
//     a session cookie that has MFAVerified=false. The only authenticated
//     route this session can hit is /api/admin/auth/mfa-challenge.
//   - Stage 2: POST /api/admin/auth/mfa-challenge (code) → flips
//     MFAVerified=true on the session. All other routes become reachable.
//   - First login: if the admin hasn't enrolled MFA yet, /me reports
//     mfa_enrolled=false, and the SPA forces enrollment via
//     /api/admin/auth/mfa-setup (returns secret + otpauth:// URI for the
//     QR code) followed by /api/admin/auth/mfa-setup-verify (code).
//
// Session TTL: 30 min idle, 8 hr absolute lifetime. Both checked on every
// request by the middleware.
package adminportal

import (
	"net/http"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

const (
	SessionCookieName = "np_admin_session"
	SessionIdleTTL    = 30 * time.Minute
	SessionMaxLife    = 8 * time.Hour
)

// Config holds the admin portal's wiring.
//
// Offline-license MINTING isn't done via the admin portal in v1 —
// admins create offline Subscriptions + enrollment tokens, and the
// existing issuer enroll path mints the JWS when a deployment redeems
// the token. That keeps signer custody confined to the issuer code path.
type Config struct {
	Store      store.Store
	Audit      audit.Log
	IssuerURL  string           // for audit-trail context
	Now        func() time.Time // injectable for tests
	Secure     bool             // emit Secure cookies (true behind TLS)
	TOTPIssuer string           // shown in the authenticator app, e.g. "infraYS"
}

// Server is the admin portal HTTP handler container.
type Server struct {
	cfg Config
}

func NewServer(cfg Config) *Server {
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	if cfg.TOTPIssuer == "" {
		cfg.TOTPIssuer = "infraYS"
	}
	return &Server{cfg: cfg}
}

// Routes returns the admin mux. Routes:
//
//   Public (no session needed):
//     POST /api/admin/auth/login
//
//   Stage-1 session (logged in, MFA pending):
//     POST /api/admin/auth/mfa-challenge
//     POST /api/admin/auth/mfa-setup
//     POST /api/admin/auth/mfa-setup-verify
//     GET  /api/admin/auth/me
//     POST /api/admin/auth/logout
//
//   Stage-2 session (logged in AND MFA verified):
//     GET  /api/admin/customers
//     GET  /api/admin/customers/{id}
//     PATCH /api/admin/customers/{id}    — set status (suspend/reactivate)
//     POST /api/admin/customers/{id}/subscriptions  — create offline sub
//     POST /api/admin/customers/{id}/enrollment-tokens — admin-issued
//     POST /api/admin/deployments/{id}/flag — toggle flagged_for_review
//     POST /api/admin/licenses/{jti}/revoke
//     GET  /api/admin/audit
//     POST /api/admin/auth/change-password
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/admin/auth/login", s.handleLogin)

	mux.HandleFunc("GET /api/admin/auth/me", s.requireStage1(s.handleMe))
	mux.HandleFunc("POST /api/admin/auth/logout", s.requireStage1(s.handleLogout))
	mux.HandleFunc("POST /api/admin/auth/mfa-challenge", s.requireStage1(s.handleMFAChallenge))
	mux.HandleFunc("POST /api/admin/auth/mfa-setup", s.requireStage1(s.handleMFASetup))
	mux.HandleFunc("POST /api/admin/auth/mfa-setup-verify", s.requireStage1(s.handleMFASetupVerify))

	mux.HandleFunc("POST /api/admin/auth/change-password", s.requireStage2(s.handleChangePassword))
	mux.HandleFunc("GET /api/admin/customers", s.requireStage2(s.handleListCustomers))
	mux.HandleFunc("GET /api/admin/customers/{id}", s.requireStage2(s.handleGetCustomer))
	mux.HandleFunc("PATCH /api/admin/customers/{id}", s.requireStage2(s.handleUpdateCustomer))
	mux.HandleFunc("POST /api/admin/customers/{id}/subscriptions", s.requireStage2(s.handleCreateOfflineSubscription))
	mux.HandleFunc("POST /api/admin/customers/{id}/enrollment-tokens", s.requireStage2(s.handleAdminCreateEnrollmentToken))
	mux.HandleFunc("POST /api/admin/deployments/{id}/flag", s.requireStage2(s.handleFlagDeployment))
	mux.HandleFunc("POST /api/admin/licenses/{jti}/revoke", s.requireStage2(s.handleRevokeLicense))
	mux.HandleFunc("GET /api/admin/audit", s.requireStage2(s.handleListAudit))

	return mux
}
