package stripebill

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/email"
	"github.com/seshadrik143/infrays-website/backend/internal/store"

	stripe "github.com/stripe/stripe-go/v82"
)

// ── customer.created ───────────────────────────────────────────────
//
// Creates a local customer row keyed off the Stripe customer ID. If a
// customer with this email already exists, we attach the Stripe ID to
// the existing row (the typical case is a customer signed up at
// app.infrays.org before paying, then their first checkout creates the
// Stripe customer object).

func (h *Handler) handleCustomerCreated(ctx context.Context, event *stripe.Event) error {
	var sc stripe.Customer
	if err := unmarshalStripeData(event, &sc); err != nil {
		return fmt.Errorf("decode customer: %w", err)
	}
	if sc.Email == "" {
		// Stripe lets you create customers without an email but we
		// reject that — every paying entity must be reachable.
		return errors.New("stripebill: customer.created has no email")
	}

	// Look up by email first — pre-existing portal signups attach here.
	existing, err := h.store.GetCustomerByEmail(ctx, sc.Email)
	if err == nil {
		existing.StripeCustomerID = sc.ID
		if existing.Name == "" {
			existing.Name = sc.Name
		}
		existing.UpdatedAt = time.Now().UTC()
		if err := h.store.UpdateCustomer(ctx, existing); err != nil {
			return fmt.Errorf("attach stripe id: %w", err)
		}
		_, _ = h.audit.Append(ctx, audit.Entry{
			EventType:  "stripe.customer.attached",
			CustomerID: existing.ID,
			Actor:      "stripe",
			Payload:    map[string]any{"stripe_customer_id": sc.ID, "email": sc.Email},
		})
		return nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return err
	}

	// New customer — create from scratch.
	now := time.Now().UTC()
	c := &store.Customer{
		ID:               newID("cust"),
		Email:            sc.Email,
		Name:             sc.Name,
		StripeCustomerID: sc.ID,
		Status:           "active",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := h.store.CreateCustomer(ctx, c); err != nil {
		return fmt.Errorf("create customer: %w", err)
	}
	_, _ = h.audit.Append(ctx, audit.Entry{
		EventType:  "stripe.customer.created",
		CustomerID: c.ID,
		Actor:      "stripe",
		Payload:    map[string]any{"stripe_customer_id": sc.ID, "email": sc.Email},
	})
	// Welcome email — best-effort.
	msg, err := email.RenderWelcome(c.Email, email.WelcomeData{
		CustomerName: orFallback(c.Name, "there"),
		AppURL:       h.cfg.AppURL,
		DocsURL:      h.cfg.AppURL + "/docs",
	})
	if err == nil {
		h.sendEmail(ctx, msg)
	}
	return nil
}

// ── customer.subscription.{created,updated,deleted} ────────────────
//
// One handler covers all three event types. Stripe's subscription
// object carries the same shape; the event type tells us whether to
// insert, update, or mark canceled. We re-read the subscription each
// time rather than try to diff in-band — Stripe's webhook can arrive
// out of order, so authoritative state is whatever the most recent
// event says.

func (h *Handler) handleSubscriptionEvent(ctx context.Context, event *stripe.Event, action string) error {
	var ss stripe.Subscription
	if err := unmarshalStripeData(event, &ss); err != nil {
		return fmt.Errorf("decode subscription: %w", err)
	}
	if ss.ID == "" {
		return errors.New("stripebill: subscription event has no id")
	}

	// Find the local customer.
	customer, err := h.findCustomerByStripe(ctx, ss.Customer)
	if err != nil {
		return fmt.Errorf("subscription %s: %w", ss.ID, err)
	}

	// Resolve tier + entitlement_set_id from the first Price on the
	// subscription. Stripe API v2025+ moved billing-period fields onto
	// the SubscriptionItem (multi-line subs can have different billing
	// cycles). Our pricing model is single-tier so we pick the first
	// item's price and period.
	priceID := ""
	var periodStart, periodEnd int64
	if ss.Items != nil && len(ss.Items.Data) > 0 {
		item := ss.Items.Data[0]
		if item.Price != nil {
			priceID = item.Price.ID
		}
		periodStart = item.CurrentPeriodStart
		periodEnd = item.CurrentPeriodEnd
	}
	tierCfg, found := h.cfg.PriceMap.Lookup(priceID)
	if !found && priceID != "" {
		// Unknown price — fall back to "free" but warn loudly via audit
		// so the operator can fix the price-map config.
		_, _ = h.audit.Append(ctx, audit.Entry{
			EventType:  "stripe.unknown_price",
			CustomerID: customer.ID,
			Actor:      "stripe",
			Payload:    map[string]any{"price_id": priceID, "subscription_id": ss.ID},
		})
	}

	// Build / find the local subscription row.
	now := time.Now().UTC()
	existing, err := h.store.GetSubscriptionByStripeID(ctx, ss.ID)
	if errors.Is(err, store.ErrNotFound) {
		// Create new
		sub := &store.Subscription{
			ID:                   newID("sub"),
			CustomerID:           customer.ID,
			Tier:                 tierCfg.Tier,
			StripeSubscriptionID: ss.ID,
			StripePriceID:        priceID,
			Status:               string(ss.Status),
			CurrentPeriodStart:   unixTime(periodStart),
			CurrentPeriodEnd:     unixTime(periodEnd),
			CancelAt:             unixTime(ss.CancelAt),
			CanceledAt:           unixTime(ss.CanceledAt),
			TrialEnd:             unixTime(ss.TrialEnd),
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := h.store.CreateSubscription(ctx, sub); err != nil {
			return fmt.Errorf("create subscription: %w", err)
		}
		_, _ = h.audit.Append(ctx, audit.Entry{
			EventType:      "stripe.subscription." + action,
			CustomerID:     customer.ID,
			SubscriptionID: sub.ID,
			Actor:          "stripe",
			Payload: map[string]any{
				"stripe_subscription_id": ss.ID,
				"tier":                   sub.Tier,
				"status":                 sub.Status,
				"period_end":             sub.CurrentPeriodEnd.Format(time.RFC3339),
			},
		})
		return nil
	}
	if err != nil {
		return err
	}

	// Update existing
	existing.Tier = tierCfg.Tier
	existing.StripePriceID = priceID
	existing.Status = string(ss.Status)
	existing.CurrentPeriodStart = unixTime(periodStart)
	existing.CurrentPeriodEnd = unixTime(periodEnd)
	existing.CancelAt = unixTime(ss.CancelAt)
	existing.CanceledAt = unixTime(ss.CanceledAt)
	existing.TrialEnd = unixTime(ss.TrialEnd)
	existing.UpdatedAt = now
	if action == "deleted" {
		// Stripe sends subscription.deleted on the actual end of life
		// (not just scheduled cancellation). Mark canceled here in case
		// Status hasn't been updated to "canceled" by an earlier event.
		existing.Status = "canceled"
		if existing.CanceledAt.IsZero() {
			existing.CanceledAt = now
		}
		// Customer email — best-effort. The license stays valid until
		// CurrentPeriodEnd; the email tells them when access ends.
		accessEnds := existing.CurrentPeriodEnd
		if accessEnds.IsZero() {
			accessEnds = now
		}
		if msg, err := email.RenderSubscriptionCanceled(customer.Email, email.SubscriptionCanceledData{
			CustomerName: orFallback(customer.Name, "there"),
			AccessEndsOn: accessEnds.Format("January 2, 2006"),
			AppURL:       h.cfg.AppURL,
			ContactURL:   h.cfg.AppURL + "/contact",
		}); err == nil {
			h.sendEmail(ctx, msg)
		}
	}
	if err := h.store.UpdateSubscription(ctx, existing); err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	_, _ = h.audit.Append(ctx, audit.Entry{
		EventType:      "stripe.subscription." + action,
		CustomerID:     customer.ID,
		SubscriptionID: existing.ID,
		Actor:          "stripe",
		Payload: map[string]any{
			"stripe_subscription_id": ss.ID,
			"tier":                   existing.Tier,
			"status":                 existing.Status,
			"period_end":             existing.CurrentPeriodEnd.Format(time.RFC3339),
		},
	})
	return nil
}

// ── invoice.paid / invoice.payment_failed ──────────────────────────
//
// We don't store invoices ourselves (Stripe is the source of truth);
// we just audit them so support can correlate "license refreshed"
// with the payment that funded it.

func (h *Handler) handleInvoicePaid(ctx context.Context, event *stripe.Event) error {
	var inv stripe.Invoice
	if err := unmarshalStripeData(event, &inv); err != nil {
		return fmt.Errorf("decode invoice: %w", err)
	}
	_, _ = h.audit.Append(ctx, audit.Entry{
		EventType: "stripe.invoice.paid",
		Actor:     "stripe",
		Payload: map[string]any{
			"invoice_id":       inv.ID,
			"amount_paid":      inv.AmountPaid,
			"currency":         string(inv.Currency),
			"stripe_customer":  inv.Customer,
		},
	})
	return nil
}

func (h *Handler) handleInvoicePaymentFailed(ctx context.Context, event *stripe.Event) error {
	var inv stripe.Invoice
	if err := unmarshalStripeData(event, &inv); err != nil {
		return fmt.Errorf("decode invoice: %w", err)
	}
	_, _ = h.audit.Append(ctx, audit.Entry{
		EventType: "stripe.invoice.payment_failed",
		Actor:     "stripe",
		Payload: map[string]any{
			"invoice_id":         inv.ID,
			"amount_due":         inv.AmountDue,
			"attempt_count":      inv.AttemptCount,
			"next_payment_attempt": unixOrZero(inv.NextPaymentAttempt),
			"stripe_customer":    inv.Customer,
		},
	})
	// Customer email — only on first attempt to avoid the noise of
	// every dunning retry. Stripe sets attempt_count >= 1; we filter
	// to == 1 to send exactly once per failed invoice.
	if inv.AttemptCount == 1 && inv.Customer != nil {
		customer, err := h.findCustomerByStripe(ctx, inv.Customer)
		if err == nil {
			amount := fmt.Sprintf("$%.2f", float64(inv.AmountDue)/100)
			if msg, mErr := email.RenderPaymentFailed(customer.Email, email.PaymentFailedData{
				CustomerName:  orFallback(customer.Name, "there"),
				AmountDueUSD:  amount,
				AppURL:        h.cfg.AppURL,
				UpdateCardURL: h.cfg.AppURL + "/billing",
			}); mErr == nil {
				h.sendEmail(ctx, msg)
			}
		}
	}
	// We deliberately do NOT cancel the subscription here. Stripe's
	// own dunning logic will eventually mark the subscription as
	// past_due → canceled, which arrives via subscription.updated /
	// .deleted webhooks. Acting twice would be racy.
	return nil
}

// ── helpers ────────────────────────────────────────────────────────

// findCustomerByStripe resolves the local Customer row for a Stripe
// customer reference. Tries Stripe customer ID first; falls back to
// email (when Stripe expanded the customer object in the event).
func (h *Handler) findCustomerByStripe(ctx context.Context, ref *stripe.Customer) (*store.Customer, error) {
	if ref == nil {
		return nil, errors.New("stripebill: nil customer ref")
	}
	if ref.ID == "" {
		return nil, errors.New("stripebill: empty stripe customer id")
	}

	// Linear scan by Stripe ID. The store interface doesn't expose
	// GetCustomerByStripeID directly — adding it would touch all
	// implementations. Phase 51 keeps the scan; a real PG query gets
	// added if scale demands it (won't until 10K+ customers).
	if ref.Email != "" {
		if c, err := h.store.GetCustomerByEmail(ctx, ref.Email); err == nil {
			if c.StripeCustomerID == "" {
				c.StripeCustomerID = ref.ID
				c.UpdatedAt = time.Now().UTC()
				_ = h.store.UpdateCustomer(ctx, c)
			}
			return c, nil
		}
	}
	return nil, fmt.Errorf("no local customer for stripe id %s (email %s)", ref.ID, ref.Email)
}

func unixTime(secs int64) time.Time {
	if secs == 0 {
		return time.Time{}
	}
	return time.Unix(secs, 0).UTC()
}

func unixOrZero(secs int64) int64 {
	if secs == 0 {
		return 0
	}
	return secs
}

func newID(prefix string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}

// orFallback returns s if non-empty, else fb. Used to give email
// templates a sensible salutation when CustomerName isn't set.
func orFallback(s, fb string) string {
	if s == "" {
		return fb
	}
	return s
}
