// Package trialscheduler is the periodic job that sends 30/7/1-day-
// before-trial-end reminder emails to SaaS customers.
//
// Lifecycle:
//
//   1. Issuer starts → NewScheduler + Start()
//   2. Every CheckInterval (default 1h), Tick() runs:
//      a. Query subs with TrialEnd in the next 30 days
//      b. For each, compute days-until-trial-end
//      c. For each threshold (30, 7, 1) where:
//           days-until-trial-end ≤ threshold
//           AND threshold NOT in sub.TrialRemindersSent
//         → send the reminder email, append threshold, update sub
//   3. Stop() cancels the goroutine and waits for it to exit
//
// Idempotency: each threshold fires exactly once per subscription via
// the TrialRemindersSent list. Lost-update races on the same threshold
// are theoretically possible if Tick() ran twice concurrently, but
// Start() runs a single goroutine — they don't.
//
// **The scheduler does NOT cancel trials**. When the trial ends,
// Stripe's own state machine flips the subscription status; the
// existing webhook handler picks that up.
package trialscheduler

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/email"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// Default reminder thresholds (days before trial end). Configurable
// via Config.Thresholds — set to []int{} to disable reminders while
// keeping the scheduler running (rare).
var DefaultThresholds = []int{30, 7, 1}

// Config bundles the scheduler's wiring. Tick interval defaults to
// 1h, which gives ~1-hour granularity on reminder firing — fine for
// 30/7/1 thresholds that round to whole days anyway.
type Config struct {
	Store         store.Store
	Email         email.Sender
	Audit         audit.Log
	AppURL        string
	CheckInterval time.Duration // default 1h
	Thresholds    []int         // default {30, 7, 1}
	LookaheadDays int           // how far ahead to scan for trials ending; default = max(Thresholds) + 1
	Now           func() time.Time // injected for tests
}

type Scheduler struct {
	cfg Config

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}

	// Counters exposed for tests / metrics.
	remindersSent uint64
	ticks         uint64
}

func New(cfg Config) (*Scheduler, error) {
	if cfg.Store == nil {
		return nil, errors.New("trialscheduler: nil store")
	}
	if cfg.Email == nil {
		return nil, errors.New("trialscheduler: nil email sender")
	}
	if cfg.Audit == nil {
		return nil, errors.New("trialscheduler: nil audit log")
	}
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = time.Hour
	}
	if len(cfg.Thresholds) == 0 {
		cfg.Thresholds = DefaultThresholds
	}
	if cfg.LookaheadDays == 0 {
		// Generous default: largest threshold + 1 day. Catches a sub
		// whose TrialEnd is at "30d - 1 minute" without our window
		// missing it.
		max := 0
		for _, t := range cfg.Thresholds {
			if t > max {
				max = t
			}
		}
		cfg.LookaheadDays = max + 1
	}
	if cfg.AppURL == "" {
		cfg.AppURL = "https://app.infrays.org"
	}
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	return &Scheduler{cfg: cfg}, nil
}

// Start kicks off the periodic tick goroutine. Idempotent — calling
// Start when already running is a no-op.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return
	}
	loopCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.done = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.done)
		// First tick on a short delay so a fresh issuer doesn't sit
		// idle for an hour on startup before catching up.
		time.Sleep(10 * time.Second)
		s.runOnce(loopCtx)

		t := time.NewTicker(s.cfg.CheckInterval)
		defer t.Stop()
		for {
			select {
			case <-loopCtx.Done():
				return
			case <-t.C:
				s.runOnce(loopCtx)
			}
		}
	}()
}

// Stop cancels the loop and waits up to 5s for it to exit.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	done := s.done
	s.cancel = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}
}

// RunOnce executes a single tick synchronously. Exposed for tests
// and for ad-hoc operator-triggered runs via the admin portal.
// Returns the number of reminders sent in this tick.
func (s *Scheduler) RunOnce(ctx context.Context) (int, error) {
	return s.runOnceCounting(ctx)
}

func (s *Scheduler) runOnce(ctx context.Context) {
	if _, err := s.runOnceCounting(ctx); err != nil {
		log.Printf("trialscheduler: tick failed: %v", err)
	}
}

func (s *Scheduler) runOnceCounting(ctx context.Context) (int, error) {
	s.mu.Lock()
	s.ticks++
	s.mu.Unlock()

	now := s.cfg.Now()
	cutoff := now.Add(time.Duration(s.cfg.LookaheadDays) * 24 * time.Hour)

	subs, err := s.cfg.Store.ListSubscriptionsWithTrialEndIn(ctx, now, cutoff)
	if err != nil {
		return 0, err
	}

	sent := 0
	for _, sub := range subs {
		// status filter — only fire reminders for trials still active.
		// A "canceled" or "past_due" trial gets handled by the regular
		// Stripe webhook path; double-emailing would be noise.
		if sub.Status != "trialing" && sub.Status != "active" {
			continue
		}

		daysLeft := int(sub.TrialEnd.Sub(now).Hours() / 24)
		// Rounding: a trial ending in 23 hours = "0 days left" by
		// integer division, but we want to fire the 1-day reminder.
		// Round UP for thresholds: if there's any time remaining,
		// daysLeft is at least 1.
		if daysLeft < 1 && sub.TrialEnd.After(now) {
			daysLeft = 1
		}

		// Find the LOWEST unfired threshold that matches the current
		// daysLeft. "Lowest matching" is the right semantic:
		//
		//   - If trial is 7 days out on first tick (no reminders sent),
		//     30 and 7 both match (daysLeft ≤ both). The 30-day
		//     reminder was "missed" — we never tick'd while daysLeft
		//     was 30. Sending it now would be backdated and confusing.
		//     Send only the 7-day one. Mark both 30 and 7 as sent so
		//     the 30-day never fires retroactively.
		//
		//   - Subsequent tick at daysLeft=1: only 1 is unfired AND
		//     matches → fire 1-day reminder.
		bestThreshold := -1
		for _, threshold := range s.cfg.Thresholds {
			if daysLeft > threshold {
				continue // too far out for this threshold
			}
			if intIn(threshold, sub.TrialRemindersSent) {
				continue // already sent
			}
			if bestThreshold == -1 || threshold < bestThreshold {
				bestThreshold = threshold
			}
		}
		if bestThreshold == -1 {
			continue // nothing to send for this sub on this tick
		}

		if err := s.sendReminder(ctx, sub, bestThreshold, daysLeft); err != nil {
			log.Printf("trialscheduler: send reminder sub=%s threshold=%d: %v", sub.ID, bestThreshold, err)
			continue
		}
		sent++

		// Mark bestThreshold AND all higher unfired thresholds as
		// sent. The higher ones were "missed" — by the time we
		// tick'd, daysLeft was already past them. Suppress them so
		// they don't fire retroactively on a future tick.
		newReminders := append([]int{}, sub.TrialRemindersSent...)
		for _, threshold := range s.cfg.Thresholds {
			if threshold >= bestThreshold && !intIn(threshold, newReminders) {
				newReminders = append(newReminders, threshold)
			}
		}
		sub.TrialRemindersSent = newReminders
		sub.UpdatedAt = now
		if err := s.cfg.Store.UpdateSubscription(ctx, sub); err != nil {
			log.Printf("trialscheduler: update sub=%s after reminder: %v", sub.ID, err)
		}
	}

	s.mu.Lock()
	s.remindersSent += uint64(sent)
	s.mu.Unlock()
	return sent, nil
}

func (s *Scheduler) sendReminder(ctx context.Context, sub *store.Subscription, threshold, daysLeft int) error {
	customer, err := s.cfg.Store.GetCustomer(ctx, sub.CustomerID)
	if err != nil {
		return err
	}
	if customer.Email == "" {
		return errors.New("customer has no email")
	}

	msg, err := email.RenderTrialExpiring(customer.Email, email.TrialExpiringData{
		CustomerName: orFallback(customer.Name, "there"),
		DaysLeft:     threshold, // use the threshold for cleaner copy ("expires in 7 days" not "expires in 6")
		AppURL:       s.cfg.AppURL,
		UpgradeURL:   s.cfg.AppURL + "/billing",
	})
	if err != nil {
		return err
	}
	if err := s.cfg.Email.Send(ctx, msg); err != nil {
		// Log + audit but don't fail the tick — other subs in this
		// batch still get processed.
		_, _ = s.cfg.Audit.Append(ctx, audit.Entry{
			EventType:      "email.send_failed",
			CustomerID:     customer.ID,
			SubscriptionID: sub.ID,
			Actor:          "trialscheduler",
			Payload: map[string]any{
				"type":      "trial_expiring",
				"threshold": threshold,
				"error":     err.Error(),
			},
		})
		return err
	}
	_, _ = s.cfg.Audit.Append(ctx, audit.Entry{
		EventType:      "trial.reminder_sent",
		CustomerID:     customer.ID,
		SubscriptionID: sub.ID,
		Actor:          "trialscheduler",
		Payload: map[string]any{
			"threshold":  threshold,
			"days_left":  daysLeft,
			"trial_end":  sub.TrialEnd.Format(time.RFC3339),
			"email_to":   customer.Email,
		},
	})
	return nil
}

// Stats returns the counters since process start. For metrics
// / admin / debugging.
func (s *Scheduler) Stats() (ticks, reminders uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ticks, s.remindersSent
}

func intIn(n int, xs []int) bool {
	for _, x := range xs {
		if x == n {
			return true
		}
	}
	return false
}

func orFallback(s, fb string) string {
	if s == "" {
		return fb
	}
	return s
}
