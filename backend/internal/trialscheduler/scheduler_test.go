package trialscheduler

import (
	"context"
	"testing"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/email"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// fixedClock returns a controlled "now" so tests can run trials that
// "expire" at deterministic times without sleeping. Each test sets
// nowFn before calling RunOnce.

func setup(t *testing.T, now time.Time) (*Scheduler, store.Store, *email.CaptureSender, audit.Log) {
	t.Helper()
	st := store.NewMemory()
	mail := email.NewCaptureSender()
	al := audit.NewMemory()
	s, err := New(Config{
		Store:         st,
		Email:         mail,
		Audit:         al,
		AppURL:        "https://app.test",
		CheckInterval: time.Hour,
		Thresholds:    []int{30, 7, 1},
		LookaheadDays: 31,
		Now:           func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	return s, st, mail, al
}

// seedTrial creates a customer + subscription whose TrialEnd is at
// `daysFromNow` days after `now`. Returns the subscription ID.
func seedTrial(t *testing.T, st store.Store, now time.Time, daysFromNow int) string {
	t.Helper()
	custID := "cust_t_" + t.Name()
	subID := "sub_t_" + t.Name()
	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: custID, Email: "trial@test.com", Name: "Trial User",
		Status: "active", CreatedAt: now, UpdatedAt: now,
	})
	_ = st.CreateSubscription(context.Background(), &store.Subscription{
		ID: subID, CustomerID: custID, Tier: "professional",
		Status:   "trialing",
		TrialEnd: now.Add(time.Duration(daysFromNow) * 24 * time.Hour).Add(time.Hour),
		// +1h fudge ensures `daysFromNow` rounds correctly: a sub
		// 7 days away should fire the 7-day reminder, not 6.
		CreatedAt: now, UpdatedAt: now,
	})
	return subID
}

// ── tests ──────────────────────────────────────────────────────────

func TestRunOnce_NoTrialsNoSends(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	s, _, mail, _ := setup(t, now)
	sent, err := s.RunOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sent != 0 || len(mail.Messages()) != 0 {
		t.Errorf("got %d sent, %d messages; want 0", sent, len(mail.Messages()))
	}
}

func TestRunOnce_30DayThresholdFires(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	s, st, mail, _ := setup(t, now)
	subID := seedTrial(t, st, now, 30) // trial ends in 30 days

	sent, err := s.RunOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sent != 1 {
		t.Errorf("sent: got %d want 1", sent)
	}
	msgs := mail.MessagesOfType("trial_expiring")
	if len(msgs) != 1 {
		t.Errorf("trial_expiring count: got %d want 1", len(msgs))
	}

	// Subscription was updated with the threshold record.
	sub, _ := st.GetSubscription(context.Background(), subID)
	if !intIn(30, sub.TrialRemindersSent) {
		t.Errorf("threshold not recorded: got %v", sub.TrialRemindersSent)
	}
}

func TestRunOnce_Idempotent_NoSecondSend(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	s, st, mail, _ := setup(t, now)
	seedTrial(t, st, now, 30)

	_, _ = s.RunOnce(context.Background())
	mail.Reset()
	_, _ = s.RunOnce(context.Background())
	if len(mail.Messages()) != 0 {
		t.Errorf("second tick sent %d emails; want 0 (already-sent dedup)", len(mail.Messages()))
	}
}

func TestRunOnce_CrossesAllThresholds(t *testing.T) {
	// Simulate ticks as the trial-end approaches: 30d, 8d, 2d, 0.5d.
	// Each crossing point should fire exactly one new reminder.
	custID := "cust_cross"
	subID := "sub_cross"

	st := store.NewMemory()
	mail := email.NewCaptureSender()
	al := audit.NewMemory()

	// Trial ends at this fixed timestamp; we move "now" forward.
	trialEnd := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: custID, Email: "x@test", Status: "active",
		CreatedAt: trialEnd.Add(-100 * 24 * time.Hour),
		UpdatedAt: trialEnd.Add(-100 * 24 * time.Hour),
	})
	_ = st.CreateSubscription(context.Background(), &store.Subscription{
		ID: subID, CustomerID: custID, Tier: "professional",
		Status: "trialing", TrialEnd: trialEnd,
		CreatedAt: trialEnd.Add(-100 * 24 * time.Hour),
		UpdatedAt: trialEnd.Add(-100 * 24 * time.Hour),
	})

	tick := func(now time.Time) int {
		s, _ := New(Config{
			Store: st, Email: mail, Audit: al,
			AppURL: "https://app.test", CheckInterval: time.Hour,
			Thresholds: []int{30, 7, 1}, LookaheadDays: 31,
			Now: func() time.Time { return now },
		})
		sent, _ := s.RunOnce(context.Background())
		return sent
	}

	// Tick A: 35 days before trial end → no fire
	if sent := tick(trialEnd.Add(-35 * 24 * time.Hour)); sent != 0 {
		t.Errorf("tick A: got %d want 0", sent)
	}
	// Tick B: 25 days before → 30-day fires
	if sent := tick(trialEnd.Add(-25 * 24 * time.Hour)); sent != 1 {
		t.Errorf("tick B: got %d want 1", sent)
	}
	// Tick C: 10 days before → still in 30-day band, no new fire
	if sent := tick(trialEnd.Add(-10 * 24 * time.Hour)); sent != 0 {
		t.Errorf("tick C: got %d want 0", sent)
	}
	// Tick D: 5 days before → 7-day fires
	if sent := tick(trialEnd.Add(-5 * 24 * time.Hour)); sent != 1 {
		t.Errorf("tick D: got %d want 1", sent)
	}
	// Tick E: 12 hours before → 1-day fires
	if sent := tick(trialEnd.Add(-12 * time.Hour)); sent != 1 {
		t.Errorf("tick E: got %d want 1", sent)
	}
	// Tick F: same as E → no fire (1-day already sent)
	if sent := tick(trialEnd.Add(-1 * time.Hour)); sent != 0 {
		t.Errorf("tick F: got %d want 0", sent)
	}

	if total := len(mail.Messages()); total != 3 {
		t.Errorf("total reminders across cycle: got %d want 3", total)
	}
	sub, _ := st.GetSubscription(context.Background(), subID)
	if len(sub.TrialRemindersSent) != 3 {
		t.Errorf("final reminders recorded: got %v want [30 7 1]", sub.TrialRemindersSent)
	}
}

func TestRunOnce_SkipsCanceledSubscriptions(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	s, st, mail, _ := setup(t, now)

	// Seed a canceled sub still inside the trial window — shouldn't get a reminder.
	custID := "cust_canceled"
	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: custID, Email: "c@test", Status: "active",
		CreatedAt: now, UpdatedAt: now,
	})
	_ = st.CreateSubscription(context.Background(), &store.Subscription{
		ID: "sub_canceled", CustomerID: custID, Tier: "professional",
		Status:    "canceled",
		TrialEnd:  now.Add(7 * 24 * time.Hour),
		CreatedAt: now, UpdatedAt: now,
	})

	sent, _ := s.RunOnce(context.Background())
	if sent != 0 || len(mail.Messages()) != 0 {
		t.Errorf("canceled sub should not get reminders; got %d sent", sent)
	}
}

func TestRunOnce_SkipsTrialAlreadyEnded(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	s, st, mail, _ := setup(t, now)

	// Trial that ended yesterday — outside the (now, cutoff] window.
	custID := "cust_past"
	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: custID, Email: "p@test", Status: "active",
		CreatedAt: now, UpdatedAt: now,
	})
	_ = st.CreateSubscription(context.Background(), &store.Subscription{
		ID: "sub_past", CustomerID: custID, Tier: "professional",
		Status:    "trialing",
		TrialEnd:  now.Add(-24 * time.Hour),
		CreatedAt: now, UpdatedAt: now,
	})

	sent, _ := s.RunOnce(context.Background())
	if sent != 0 || len(mail.Messages()) != 0 {
		t.Errorf("past-end trial should be filtered out; got %d sent", sent)
	}
}

func TestNew_RejectsMissingDeps(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{"no store", Config{Email: email.NewNoopSender(), Audit: audit.NewMemory()}},
		{"no email", Config{Store: store.NewMemory(), Audit: audit.NewMemory()}},
		{"no audit", Config{Store: store.NewMemory(), Email: email.NewNoopSender()}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := New(c.cfg); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestStats_Increments(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	s, st, _, _ := setup(t, now)
	seedTrial(t, st, now, 7)

	if ticks, sent := s.Stats(); ticks != 0 || sent != 0 {
		t.Errorf("pre-tick: got %d/%d", ticks, sent)
	}
	_, _ = s.RunOnce(context.Background())
	if ticks, sent := s.Stats(); ticks != 1 || sent != 1 {
		t.Errorf("post-tick: got %d ticks %d sent; want 1/1", ticks, sent)
	}
}
