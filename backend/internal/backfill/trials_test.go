package backfill

import (
	"testing"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

func TestPlanTrialBackfill(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	custs := []*store.Customer{
		{ID: "c_intrial", Email: "a@x.com", CreatedAt: now.Add(-5 * 24 * time.Hour)},   // 5 days in → 10 left
		{ID: "c_expired", Email: "b@x.com", CreatedAt: now.Add(-40 * 24 * time.Hour)},  // trial long over
		{ID: "c_hassub", Email: "c@x.com", CreatedAt: now.Add(-2 * 24 * time.Hour)},    // has a sub already
		{ID: "c_fresh", Email: "d@x.com", CreatedAt: now.Add(-1 * time.Hour)},          // just signed up
	}
	hasSub := func(id string) bool { return id == "c_hassub" }

	got := PlanTrialBackfill(custs, hasSub, now, 15)

	ids := map[string]time.Time{}
	for _, c := range got {
		ids[c.CustomerID] = c.TrialEnd
	}
	if _, ok := ids["c_intrial"]; !ok {
		t.Error("in-trial customer should be backfilled")
	}
	if _, ok := ids["c_fresh"]; !ok {
		t.Error("just-signed-up customer should be backfilled")
	}
	if _, ok := ids["c_expired"]; ok {
		t.Error("expired-trial customer should be skipped")
	}
	if _, ok := ids["c_hassub"]; ok {
		t.Error("customer with a subscription should be skipped (idempotent)")
	}
	// trial_end is signup + 15 days
	want := custs[0].CreatedAt.Add(15 * 24 * time.Hour)
	if !ids["c_intrial"].Equal(want) {
		t.Errorf("trial_end: got %v want %v", ids["c_intrial"], want)
	}
}

func TestNewTrialSubscription(t *testing.T) {
	now := time.Now()
	c := Candidate{CustomerID: "c1", Email: "x@y.com", TrialEnd: now.Add(15 * 24 * time.Hour)}
	sub := NewTrialSubscription("sub_1", c, now)
	if sub.Status != "trialing" || sub.TrialEnd != c.TrialEnd || sub.CustomerID != "c1" || !sub.ManualOffline {
		t.Errorf("unexpected subscription: %+v", sub)
	}
}
