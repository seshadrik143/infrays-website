// Command backfill-trials creates a trialing subscription for existing
// customers who signed up before the signup flow started creating one, so the
// expiry-reminder scheduler can find them.
//
// SAFE BY DEFAULT — dry run unless -apply is passed. Idempotent (skips
// customers that already have any subscription).
//
//	# preview against production
//	PG_URL=postgres://... go run ./cmd/backfill-trials
//	# actually create the rows
//	PG_URL=postgres://... go run ./cmd/backfill-trials -apply
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/backfill"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

func main() {
	pgURL := flag.String("pg-url", os.Getenv("PG_URL"), "PostgreSQL DSN (default: $PG_URL)")
	trialDays := flag.Int("trial-days", 15, "trial length in days (signup + this = trial_end)")
	limit := flag.Int("limit", 100000, "max customers to scan")
	apply := flag.Bool("apply", false, "actually create subscriptions (default: dry run)")
	flag.Parse()

	if *pgURL == "" {
		log.Fatal("backfill-trials: -pg-url or PG_URL is required")
	}

	ctx := context.Background()
	st, err := store.NewPG(ctx, *pgURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	customers, err := st.ListCustomers(ctx, "", *limit)
	if err != nil {
		log.Fatalf("list customers: %v", err)
	}

	hasSub := func(id string) bool {
		subs, err := st.ListSubscriptionsByCustomer(ctx, id)
		if err != nil {
			log.Printf("warn: list subs for %s: %v (treating as has-sub to be safe)", id, err)
			return true
		}
		return len(subs) > 0
	}

	now := time.Now().UTC()
	candidates := backfill.PlanTrialBackfill(customers, hasSub, now, *trialDays)

	fmt.Printf("scanned %d customers; %d need a backfilled trial subscription\n", len(customers), len(candidates))
	mode := "DRY RUN (no changes) — pass -apply to create"
	if *apply {
		mode = "APPLYING"
	}
	fmt.Println("mode:", mode)

	created := 0
	for _, c := range candidates {
		fmt.Printf("  %-40s trial_end=%s\n", c.Email, c.TrialEnd.Format("2006-01-02"))
		if !*apply {
			continue
		}
		sub := backfill.NewTrialSubscription("sub_"+randHex(12), c, now)
		if err := st.CreateSubscription(ctx, sub); err != nil {
			log.Printf("  ERROR creating sub for %s: %v", c.Email, err)
			continue
		}
		created++
	}
	if *apply {
		fmt.Printf("created %d trial subscriptions\n", created)
	}
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
