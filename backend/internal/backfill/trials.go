// Package backfill holds one-off data migrations. PlanTrialBackfill decides
// which existing customers need a trialing subscription created retroactively
// (they signed up before signup started creating one), so the expiry-reminder
// scheduler can find them.
package backfill

import (
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// Candidate is a customer that should get a backfilled trial subscription.
type Candidate struct {
	CustomerID string
	Email      string
	TrialEnd   time.Time
}

// PlanTrialBackfill returns the customers needing a trial subscription:
// those with NO existing subscription whose trial (signup + trialDays) is
// still in the future. Customers already past their trial window are skipped —
// backdating a "trialing" sub for them would be misleading and send no
// reminders anyway. Idempotent: hasSub makes re-runs a no-op.
func PlanTrialBackfill(customers []*store.Customer, hasSub func(customerID string) bool, now time.Time, trialDays int) []Candidate {
	var out []Candidate
	for _, c := range customers {
		if hasSub(c.ID) {
			continue
		}
		trialEnd := c.CreatedAt.Add(time.Duration(trialDays) * 24 * time.Hour)
		if !trialEnd.After(now) {
			continue // trial already ended — no reminders would fire
		}
		out = append(out, Candidate{CustomerID: c.ID, Email: c.Email, TrialEnd: trialEnd})
	}
	return out
}

// NewTrialSubscription builds the subscription row for a candidate, matching
// what signup now creates for new users.
func NewTrialSubscription(id string, c Candidate, now time.Time) *store.Subscription {
	return &store.Subscription{
		ID:                 id,
		CustomerID:         c.CustomerID,
		Tier:               "professional",
		Status:             "trialing",
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   c.TrialEnd,
		TrialEnd:           c.TrialEnd,
		ManualOffline:      true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}
