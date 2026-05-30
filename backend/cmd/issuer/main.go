// issuer is the infraYS license issuer service. Listens on the
// address given by --addr (default :8443 with TLS, :8080 without)
// and serves the endpoints documented in
// internal/issuer/server.go's Routes().
//
// Phase 49 ships with the in-memory store. Set --pg-url to switch
// to PostgreSQL (Phase 49 follow-up).
//
// Signing key:
//
//	--signer=local --signer-key-file=<pem>     (DEV / CI ONLY)
//	--signer=gcp-kms ... (TODO)
//	--signer=vault   ... (TODO)
//
// See backend/docs/LICENSE_KEY_CUSTODY.md for production key custody.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/adminportal"
	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/email"
	"github.com/seshadrik143/infrays-website/backend/internal/issuer"
	"github.com/seshadrik143/infrays-website/backend/internal/obs"
	"github.com/seshadrik143/infrays-website/backend/internal/portal"
	"github.com/seshadrik143/infrays-website/backend/internal/signing"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
	"github.com/seshadrik143/infrays-website/backend/internal/stripebill"
	"github.com/seshadrik143/infrays-website/backend/internal/trialscheduler"
)

func main() {
	// Initialize structured JSON logger first so all subsequent boot
	// messages land in the same downstream stream. Level via env so
	// operators can crank to debug without a redeploy by `fly secrets
	// set NP_LOG_LEVEL=debug` + a single machine restart.
	obs.SetDefault(obs.NewLogger(os.Getenv("NP_LOG_LEVEL")))

	addr := flag.String("addr", ":8080", "Listen address")
	issuerURL := flag.String("issuer-url", "license.infrays.org", "Hostname embedded in JWS 'iss' claim")
	graceDays := flag.Int("grace-days", 90, "Default grace period after license expiry (days)")
	refreshHours := flag.Int("refresh-hours", 24, "Hint to clients for how often to refresh")

	signerSource := flag.String("signer", "local", "local | gcp-kms | vault")
	signerKeyFile := flag.String("signer-key-file", "", "Path to Ed25519 PEM (local source). Use --signer-key-pem-env for cloud deploys.")
	signerKeyPemEnv := flag.String("signer-key-pem-env", "", "Env var name holding PEM-encoded Ed25519 private key (e.g. NP_LICENSE_SIGNER_KEY). Preferred over --signer-key-file for cloud deploys with secrets managers (Fly.io, Cloud Run, etc.).")
	signerKID := flag.String("signer-kid", "np-dev-2026-01", "Key id embedded in JWS header")

	pgURL := flag.String("pg-url", "", "PostgreSQL DSN. When empty, uses in-memory store (lost on restart). Env override: PG_URL")

	flag.Parse()

	// Env override for pg-url so the operator can set it without
	// putting credentials in argv (which leaks to ps).
	if env := os.Getenv("PG_URL"); env != "" && *pgURL == "" {
		*pgURL = env
	}

	// Build the signer.
	var signer signing.Signer
	var err error
	switch *signerSource {
	case "local":
		switch {
		case *signerKeyPemEnv != "":
			raw := os.Getenv(*signerKeyPemEnv)
			if raw == "" {
				log.Fatalf("env %s is empty or unset", *signerKeyPemEnv)
			}
			signer, err = signing.NewLocalSignerFromPEMBytes(*signerKID, []byte(raw))
		case *signerKeyFile != "":
			signer, err = signing.LoadLocalSigner(*signerKID, *signerKeyFile)
		default:
			log.Fatal("--signer-key-file OR --signer-key-pem-env is required for --signer=local")
		}
	default:
		log.Fatalf("unsupported --signer=%s (only 'local' is wired in Phase 49; see backend/docs/LICENSE_KEY_CUSTODY.md)", *signerSource)
	}
	if err != nil {
		log.Fatalf("load signer: %v", err)
	}
	defer signer.Close()

	// Store: PG when configured, in-memory otherwise.
	var st store.Store
	if *pgURL != "" {
		pg, err := store.NewPG(context.Background(), *pgURL)
		if err != nil {
			log.Fatalf("pg: %v", err)
		}
		st = pg
		obs.Default().Info("store backend selected", "backend", "postgres")
	} else {
		st = store.NewMemory()
		obs.Default().Warn("in-memory store — state lost on restart", "remediation", "set --pg-url or PG_URL for production")
	}
	defer st.Close()

	// Audit log stays in-memory for Phase 49. PG audit follows the
	// hash-chain pattern from NodePulse Phase 38; deferred until a
	// real operator deploys this — building it now would block the
	// dev / smoke loop on Postgres.
	auditLog := audit.NewMemory()
	defer auditLog.Close()

	// Seed default entitlement sets so enrollment can succeed
	// without manual admin setup.
	seedEntitlements(st)

	// Build server + warn if admin secret is unset.
	if os.Getenv("NP_ISSUER_ADMIN_SECRET") == "" {
		obs.Default().Warn("NP_ISSUER_ADMIN_SECRET unset — admin endpoints will reject all requests")
	}

	// ── Phase 51.5: email sender ───────────────────────────────────
	// Noop sender when NP_POSTMARK_SERVER_TOKEN is unset — logs the
	// would-have-been-sent so trigger paths are still observable.
	mailer := email.NewSenderFromEnv()
	appURL := os.Getenv("NP_APP_URL")
	if appURL == "" {
		appURL = "https://app.infrays.org"
	}

	// ── Phase 51: Stripe wiring ────────────────────────────────────
	// All three env vars optional — issuer runs without Stripe
	// before sales is set up. Webhooks + checkout routes only
	// register when their respective configs are present.
	var stripeWebhook, stripeCheckout issuer.StripeBillHandler
	var checkoutCreator portal.CheckoutCreator // self-serve checkout for the portal
	priceMap, err := stripebill.ParseTierMappingFromEnv(os.Getenv("NP_STRIPE_PRICE_MAP"))
	if err != nil {
		log.Fatalf("stripe price map: %v", err)
	}
	if whSecret := os.Getenv("NP_STRIPE_WEBHOOK_SECRET"); whSecret != "" {
		wh, err := stripebill.NewHandler(stripebill.Config{
			WebhookSecret: whSecret,
			PriceMap:      priceMap,
			Email:         mailer,
			AppURL:        appURL,
		}, st, auditLog)
		if err != nil {
			log.Fatalf("stripe webhook: %v", err)
		}
		stripeWebhook = wh
		obs.Default().Info("stripe webhook handler registered")
	}
	if apiKey := os.Getenv("NP_STRIPE_SECRET_KEY"); apiKey != "" {
		ch, err := stripebill.NewCheckoutHandler(apiKey, appURL, priceMap)
		if err != nil {
			log.Fatalf("stripe checkout: %v", err)
		}
		stripeCheckout = ch
		checkoutCreator = ch // same handler serves the authenticated portal path
		obs.Default().Info("stripe checkout handler registered")
	}

	// Purchasable tiers for the portal plan picker, derived from the
	// configured price map so the UI can't offer a tier that won't
	// resolve to a Stripe price.
	var portalPlans []portal.PlanOption
	for _, p := range priceMap.ListPlans() {
		intervals := make([]string, 0, len(p.Options))
		for _, o := range p.Options {
			intervals = append(intervals, o.Interval)
		}
		portalPlans = append(portalPlans, portal.PlanOption{Tier: p.Tier, Intervals: intervals})
	}

	srv := issuer.NewServer(issuer.Config{
		Store:                st,
		Audit:                auditLog,
		Signer:               signer,
		IssuerURL:            *issuerURL,
		DefaultGraceDays:     *graceDays,
		RefreshIntervalHours: *refreshHours,
		StripeWebhook:        stripeWebhook,
		StripeCheckout:       stripeCheckout,
		Email:                mailer,
		AppURL:               appURL,
	})

	// ── Phase 52: customer portal at app.infrays.org ─────────────
	// Mounted on the same Go binary as the issuer. The mux below
	// dispatches /api/portal/* to the portal Server and everything
	// else to the issuer Server. Future task #91 will embed the
	// React SPA and add a fallback handler for /.
	var billingPortal portal.BillingPortalCreator
	if apiKey := os.Getenv("NP_STRIPE_SECRET_KEY"); apiKey != "" {
		bp, err := stripebill.NewPortalSessionCreator(apiKey, appURL)
		if err != nil {
			log.Fatalf("stripe billing portal: %v", err)
		}
		billingPortal = bp
		obs.Default().Info("stripe billing portal session creator registered")
	}
	portalSrv := portal.NewServer(portal.Config{
		Store:         st,
		Audit:         auditLog,
		Email:         mailer,
		BillingPortal: billingPortal,
		Checkout:      checkoutCreator,
		Plans:         portalPlans,
		AppURL:        appURL,
		Secure:        strings.HasPrefix(appURL, "https://"),
	})

	// ── Phase 52.5: admin portal ───────────────────────────────────
	// Bootstrap an admin user via env if none exist yet — lets a fresh
	// deploy mint the first admin without manual DB poking.
	bootstrapAdmin(context.Background(), st)
	adminSrv := adminportal.NewServer(adminportal.Config{
		Store:      st,
		Audit:      auditLog,
		IssuerURL:  *issuerURL,
		Secure:     strings.HasPrefix(appURL, "https://"),
		TOTPIssuer: "infraYS",
	})

	rootMux := http.NewServeMux()
	// JSON404 wraps each sub-mux so unmatched routes return JSON
	// "{\"error\":\"not found\"}" instead of stdlib's plain-text fallback.
	rootMux.Handle("/api/portal/", obs.HTTPMiddleware("portal.api", obs.JSON404(portalSrv.Routes())))
	rootMux.Handle("/api/admin/", obs.HTTPMiddleware("admin.api", obs.JSON404(adminSrv.Routes())))
	issuerMux := srv.Routes()
	wrappedIssuer := obs.HTTPMiddleware("issuer.api", obs.JSON404(issuerMux))
	for _, p := range []string{"/v1/", "/healthz", "/internal/", "/.well-known/"} {
		rootMux.Handle(p, wrappedIssuer)
	}
	// /metrics — opt-in via NP_METRICS_USER + NP_METRICS_PASSWORD.
	// When either is empty, the handler returns 404, never exposing
	// metrics unauthenticated. NOT wrapped in HTTPMiddleware to avoid
	// recursive instrumentation (scraping noise).
	rootMux.Handle("/metrics", obs.MetricsHandler(
		os.Getenv("NP_METRICS_USER"),
		os.Getenv("NP_METRICS_PASSWORD"),
	))
	// Everything else falls through to the embedded portal SPA — index.html
	// at "/" plus React Router's client-side routes (/login, /dashboard, ...).
	rootMux.Handle("/", obs.HTTPMiddleware("spa", portal.SPAHandler()))

	httpSrv := &http.Server{
		Addr: *addr,
		// Chain (outermost first):
		//   SecurityHeaders — defense-in-depth response headers on
		//     every response, including 404/405 fallbacks.
		//   BodyLimit — cap request bodies before handlers see them.
		//   rootMux — actual routing.
		Handler:      obs.SecurityHeaders(obs.DenyCrossOrigin(obs.BodyLimit(0, obs.TrailingSlash(rootMux)))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Phase 51.5: trial-expiring reminder scheduler ──────────────
	// Periodic job that sends 30/7/1-day-before-trial-end reminders
	// to customers in trial. Runs as a goroutine; survives without
	// real Postmark (uses the same email.Sender — falls back to
	// noop when NP_POSTMARK_SERVER_TOKEN is unset).
	trialSched, err := trialscheduler.New(trialscheduler.Config{
		Store:         st,
		Email:         mailer,
		Audit:         auditLog,
		AppURL:        appURL,
		CheckInterval: time.Hour,
		Thresholds:    []int{30, 7, 1},
	})
	if err != nil {
		log.Fatalf("trialscheduler: %v", err)
	}
	trialCtx, trialCancel := context.WithCancel(context.Background())
	trialSched.Start(trialCtx)
	obs.Default().Info("trialscheduler started", "thresholds_days", []int{30, 7, 1}, "tick", "1h")

	go func() {
		obs.Default().Info("issuer listening", "addr", *addr, "signer", *signerSource, "kid", signer.KID())
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	obs.Default().Info("issuer shutting down")
	trialCancel()
	trialSched.Stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		obs.Default().Warn("http shutdown error", "err", err)
	}
	obs.Default().Info("issuer stopped")
}

// bootstrapAdmin creates the first admin user from environment vars
// if no admin exists yet. Idempotent — does nothing if any admin row
// is already present, so operators can rotate the bootstrap env
// safely. Vars (both required):
//
//	NP_ADMIN_BOOTSTRAP_EMAIL    e.g. "admin@infrays.org"
//	NP_ADMIN_BOOTSTRAP_PASSWORD long random string; admin must change
//	                            it on first login + enroll MFA
//
// Without these the admin portal still works, but you'll need to
// seed at least one row another way before logging in.
func bootstrapAdmin(ctx context.Context, st store.Store) {
	existing, err := st.ListAdminUsers(ctx)
	if err == nil && len(existing) > 0 {
		return
	}
	email := os.Getenv("NP_ADMIN_BOOTSTRAP_EMAIL")
	password := os.Getenv("NP_ADMIN_BOOTSTRAP_PASSWORD")
	if email == "" || password == "" {
		obs.Default().Warn("admin portal: no admin users present", "remediation", "set NP_ADMIN_BOOTSTRAP_EMAIL + NP_ADMIN_BOOTSTRAP_PASSWORD to seed the first one")
		return
	}
	hash, err := adminportal.HashPassword(password)
	if err != nil {
		obs.Default().Error("admin bootstrap hash failed", "err", err)
		return
	}
	a := &store.AdminUser{
		ID:           adminportal.NewAdminID(),
		Email:        email,
		PasswordHash: hash,
		Role:         "admin",
		CreatedAt:    time.Now().UTC(),
	}
	if err := st.CreateAdminUser(ctx, a); err != nil {
		obs.Default().Error("admin bootstrap create failed", "err", err)
		return
	}
	obs.Default().Info("admin portal bootstrapped", "email", email, "next_step", "enroll MFA on first login")
}

// seedEntitlements creates the three baseline entitlement sets so an
// operator can mint licenses without first calling
// /internal/admin/entitlement-sets. Production deploys should call
// the admin endpoint with the manifest they actually want; this is
// dev/CI convenience.
func seedEntitlements(st store.Store) {
	ctx := context.Background()
	sets := []store.EntitlementSet{
		{
			ID:        "free-v1",
			Name:      "Free",
			Version:   1,
			Features:  []string{},
			Limits:    store.Limits{MaxAgents: 3, MaxMetricsPerSec: 500, MaxLogGBPerDay: 1, MaxAlertRules: 5, RetentionDays: 7},
			CreatedAt: time.Now().UTC(),
		},
		{
			ID:        "professional-v1",
			Name:      "Professional",
			Version:   1,
			Features:  []string{"audit_log", "advanced_alerts"},
			Limits:    store.Limits{MaxAgents: 50, MaxMetricsPerSec: 10000, MaxLogGBPerDay: 100, MaxAlertRules: 100, RetentionDays: 90},
			CreatedAt: time.Now().UTC(),
		},
		{
			ID:        "enterprise-v1",
			Name:      "Enterprise",
			Version:   1,
			Features:  []string{"sso_oidc", "audit_log", "compliance", "advanced_alerts", "multi_tenant", "raft_ha"},
			Limits:    store.Limits{},
			CreatedAt: time.Now().UTC(),
		},
	}
	for _, s := range sets {
		if err := st.CreateEntitlementSet(ctx, &s); err != nil && !strings.Contains(err.Error(), "already exists") {
			obs.Default().Warn("seed entitlement failed", "id", s.ID, "err", err)
		}
	}
	_ = fmt.Sprintf
}
