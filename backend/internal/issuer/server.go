// Package issuer is the license issuer service. Exposes the HTTP
// endpoints documented at docs/api/issuer.md (TODO):
//
//   POST /v1/enroll                    — token → JWS license
//   POST /v1/refresh                   — current license → re-signed
//   GET  /v1/entitlements/{id}.json    — feature manifest
//   GET  /v1/well-known/keys           — public keys + kill-switch
//   POST /v1/webhooks/stripe           — Stripe events (Phase 51)
//   POST /internal/admin/*             — staff endpoints (Phase 52.5)
//
// Constructed via NewServer + Routes(). Tests use httptest.NewServer
// to exercise the full stack against an in-memory Store.
package issuer

import (
	"net/http"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/signing"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// StripeBillHandler is the narrow http.Handler interface that the
// Stripe webhook + Checkout handlers satisfy. We keep this an
// interface (rather than a concrete *stripebill.Handler) to avoid
// pulling the stripe-go dep into this package.
type StripeBillHandler interface {
	http.Handler
}

// Config holds the issuer's wiring. Constructed by main; tests
// construct it directly.
type Config struct {
	Store                store.Store
	Audit                audit.Log
	Signer               signing.Signer
	IssuerURL            string           // e.g. "license.infrays.org" — embedded in JWS iss claim
	DefaultGraceDays     int              // license grace_until = expires_at + N days
	RefreshIntervalHours int              // hint to clients
	Now                  func() time.Time // injectable for tests

	// Phase 51: Stripe wiring. Both optional — when nil, the
	// /v1/webhooks/stripe and /v1/checkout/create-session routes are
	// not registered. Lets the issuer run without Stripe credentials
	// during pre-launch dev.
	StripeWebhook  StripeBillHandler
	StripeCheckout StripeBillHandler
}

// Server is the HTTP handler container. Keep it concise — handlers
// reach into Config directly for shared state.
type Server struct {
	cfg Config
}

func NewServer(cfg Config) *Server {
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	if cfg.IssuerURL == "" {
		cfg.IssuerURL = "license.infrays.org"
	}
	if cfg.DefaultGraceDays == 0 {
		cfg.DefaultGraceDays = 90
	}
	if cfg.RefreshIntervalHours == 0 {
		cfg.RefreshIntervalHours = 24
	}
	return &Server{cfg: cfg}
}

// Routes returns the mux populated with all issuer endpoints. Caller
// is responsible for wrapping with TLS, rate limiting, request
// logging, etc. — those are deploy-environment concerns.
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/enroll", s.handleEnroll)
	mux.HandleFunc("POST /v1/refresh", s.handleRefresh)
	mux.HandleFunc("GET /v1/entitlements/", s.handleEntitlements)
	mux.HandleFunc("GET /v1/well-known/keys", s.handleWellKnownKeys)
	mux.HandleFunc("GET /healthz", s.handleHealthz)

	// Phase 51: Stripe — registered only when configured.
	if s.cfg.StripeWebhook != nil {
		mux.Handle("POST /v1/webhooks/stripe", s.cfg.StripeWebhook)
	}
	if s.cfg.StripeCheckout != nil {
		mux.Handle("POST /v1/checkout/create-session", s.cfg.StripeCheckout)
	}

	// Admin (auth handled in middleware; for Phase 49 these are
	// behind a shared-secret header — Phase 52.5 wires real RBAC).
	mux.HandleFunc("POST /internal/admin/customers", s.handleAdminCreateCustomer)
	mux.HandleFunc("POST /internal/admin/subscriptions", s.handleAdminCreateSubscription)
	mux.HandleFunc("POST /internal/admin/enrollment-tokens", s.handleAdminCreateEnrollmentToken)
	mux.HandleFunc("POST /internal/admin/licenses/revoke", s.handleAdminRevokeLicense)
	mux.HandleFunc("POST /internal/admin/entitlement-sets", s.handleAdminCreateEntitlementSet)
	mux.HandleFunc("GET /internal/admin/audit", s.handleAdminAuditTail)

	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
