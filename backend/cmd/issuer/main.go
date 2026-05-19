// issuer is the infraYS license issuer service. Listens on the
// address given by --addr (default :8443 with TLS, :8080 without)
// and serves the endpoints documented in
// internal/issuer/server.go's Routes().
//
// Phase 49 ships with the in-memory store. Set --pg-url to switch
// to PostgreSQL (Phase 49 follow-up).
//
// Signing key:
//   --signer=local --signer-key-file=<pem>     (DEV / CI ONLY)
//   --signer=gcp-kms ... (TODO)
//   --signer=vault   ... (TODO)
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

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/issuer"
	"github.com/seshadrik143/infrays-website/backend/internal/signing"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "Listen address")
	issuerURL := flag.String("issuer-url", "license.infrays.org", "Hostname embedded in JWS 'iss' claim")
	graceDays := flag.Int("grace-days", 90, "Default grace period after license expiry (days)")
	refreshHours := flag.Int("refresh-hours", 24, "Hint to clients for how often to refresh")

	signerSource := flag.String("signer", "local", "local | gcp-kms | vault")
	signerKeyFile := flag.String("signer-key-file", "", "Path to Ed25519 PEM (local source)")
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
		if *signerKeyFile == "" {
			log.Fatal("--signer-key-file is required for --signer=local")
		}
		signer, err = signing.LoadLocalSigner(*signerKID, *signerKeyFile)
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
		log.Printf("store: PostgreSQL")
	} else {
		st = store.NewMemory()
		log.Println("⚠  store: in-memory (state lost on restart; set --pg-url for production)")
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
		log.Println("⚠  NP_ISSUER_ADMIN_SECRET is not set — admin endpoints will reject all requests")
	}

	srv := issuer.NewServer(issuer.Config{
		Store:                st,
		Audit:                auditLog,
		Signer:               signer,
		IssuerURL:            *issuerURL,
		DefaultGraceDays:     *graceDays,
		RefreshIntervalHours: *refreshHours,
	})

	httpSrv := &http.Server{
		Addr:         *addr,
		Handler:      srv.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("issuer listening on %s (signer=%s kid=%s)", *addr, *signerSource, signer.KID())
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Println("issuer shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("issuer stopped")
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
			log.Printf("seed entitlement %s: %v", s.ID, err)
		}
	}
	_ = fmt.Sprintf
}
