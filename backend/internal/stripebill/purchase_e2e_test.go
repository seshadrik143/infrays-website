package stripebill

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// purchase_e2e_test.go is the end-to-end purchase-lifecycle suite. Each
// Stripe test card produces a specific sequence of webhook events; this
// file replays those exact sequences and asserts the resulting
// application state (subscription status, customer binding, dunning
// emails). It is the offline twin of scripts/stripe-e2e.sh, which drives
// the same scenarios against real Stripe test mode.
//
// Card → behavior reference (Stripe test cards):
//
//	4242 4242 4242 4242  success                → active, invoice.paid
//	4000 0025 0000 3155  requires 3DS, succeeds  → incomplete → active
//	4000 0000 0000 3220  requires 3DS, fails     → incomplete → incomplete_expired
//	4000 0000 0000 0002  generic decline         → incomplete + invoice.payment_failed
//	4000 0000 0000 9995  insufficient_funds      → incomplete + invoice.payment_failed
//	4000 0000 0000 0341  attaches, later fails   → active then past_due → canceled
//
// All amounts are in cents; the professional tier is $199/mo = 19900.

const proPriceCents = 19900

// ── helpers ─────────────────────────────────────────────────────────

// seedPortalCustomer creates a local customer that started a self-serve
// checkout (no Stripe ID linked yet). Returns its local ID.
func seedPortalCustomer(t *testing.T, st store.Store, id, email string) string {
	t.Helper()
	now := time.Now().UTC()
	if err := st.CreateCustomer(context.Background(), &store.Customer{
		ID: id, Email: email, Status: "active", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	return id
}

// subEvent builds a customer.subscription.* data object. metaCustID,
// when non-empty, simulates the portal's checkout metadata binding.
func subEvent(stripeSubID, stripeCustID, email, metaCustID, priceID, status string, now time.Time) map[string]any {
	m := map[string]any{
		"id":       stripeSubID,
		"customer": map[string]any{"id": stripeCustID, "email": email},
		"status":   status,
		"items": map[string]any{
			"data": []map[string]any{
				{
					"price":                map[string]any{"id": priceID},
					"current_period_start": now.Unix(),
					"current_period_end":   now.Add(30 * 24 * time.Hour).Unix(),
				},
			},
		},
	}
	if metaCustID != "" {
		m["metadata"] = map[string]any{"customer_id": metaCustID}
	}
	return m
}

// invoiceEvent builds an invoice.* data object. The dunning-email path
// resolves the customer by email, so the customer ref carries one (real
// Stripe expands it on invoice events we subscribe to).
func invoiceEvent(invID, stripeCustID, email string, attempt, amountDue int64) map[string]any {
	return map[string]any{
		"id":            invID,
		"customer":      map[string]any{"id": stripeCustID, "email": email},
		"amount_due":    amountDue,
		"amount_paid":   int64(0),
		"attempt_count": attempt,
		"currency":      "usd",
	}
}

// assertSubStatus fetches the subscription by Stripe ID and checks its
// stored status + binding to the expected local customer.
func assertSubStatus(t *testing.T, st store.Store, stripeSubID, wantStatus, wantCustomerID string) {
	t.Helper()
	got, err := st.GetSubscriptionByStripeID(context.Background(), stripeSubID)
	if err != nil || got == nil {
		t.Fatalf("subscription %s not found: %v", stripeSubID, err)
	}
	if got.Status != wantStatus {
		t.Errorf("status: got %q want %q", got.Status, wantStatus)
	}
	if wantCustomerID != "" && got.CustomerID != wantCustomerID {
		t.Errorf("customer binding: got %q want %q", got.CustomerID, wantCustomerID)
	}
}

// ── POSITIVE: card 4242 — clean successful purchase ─────────────────

func TestE2E_Card4242_SuccessfulPurchase(t *testing.T) {
	h, st, _, cap := newTestHandlerWithEmail(t)
	cid := seedPortalCustomer(t, st, "cust_e2e_ok", "ok@buyer.test")
	now := time.Now().UTC()

	// Stripe creates the subscription active immediately on a good card.
	w := postEvent(t, h, "evt_ok_sub", "customer.subscription.created",
		subEvent("sub_ok", "cus_ok", "ok@buyer.test", cid, "price_test_pro", "active", now))
	if w.Code != http.StatusOK {
		t.Fatalf("sub.created: %d %s", w.Code, w.Body.String())
	}
	w = postEvent(t, h, "evt_ok_inv", "invoice.paid",
		invoiceEvent("in_ok", "cus_ok", "ok@buyer.test", 1, proPriceCents))
	if w.Code != http.StatusOK {
		t.Fatalf("invoice.paid: %d", w.Code)
	}

	assertSubStatus(t, st, "sub_ok", "active", cid)
	// First paid conversion links the Stripe customer ID for the portal.
	c, _ := st.GetCustomer(context.Background(), cid)
	if c.StripeCustomerID != "cus_ok" {
		t.Errorf("stripe id not linked: %q", c.StripeCustomerID)
	}
	// No dunning email on a clean purchase.
	if n := len(cap.MessagesOfType("payment_failed")); n != 0 {
		t.Errorf("unexpected %d payment_failed emails", n)
	}
}

// ── POSITIVE: trial → active conversion ─────────────────────────────

func TestE2E_TrialConversion(t *testing.T) {
	h, st, _, _ := newTestHandlerWithEmail(t)
	cid := seedPortalCustomer(t, st, "cust_e2e_trial", "trial@buyer.test")
	now := time.Now().UTC()

	postEvent(t, h, "evt_tr_1", "customer.subscription.created",
		subEvent("sub_tr", "cus_tr", "trial@buyer.test", cid, "price_test_pro", "trialing", now))
	assertSubStatus(t, st, "sub_tr", "trialing", cid)

	// Trial ends, card charged successfully → active.
	postEvent(t, h, "evt_tr_2", "customer.subscription.updated",
		subEvent("sub_tr", "cus_tr", "trial@buyer.test", cid, "price_test_pro", "active", now))
	assertSubStatus(t, st, "sub_tr", "active", cid)
}

// ── POSITIVE: card 4000…3155 — 3DS required, authenticated ──────────

func TestE2E_Card3155_3DSAuthenticated(t *testing.T) {
	h, st, _, _ := newTestHandlerWithEmail(t)
	cid := seedPortalCustomer(t, st, "cust_e2e_3ds", "3ds@buyer.test")
	now := time.Now().UTC()

	// Before authentication the subscription sits incomplete.
	postEvent(t, h, "evt_3ds_1", "customer.subscription.created",
		subEvent("sub_3ds", "cus_3ds", "3ds@buyer.test", cid, "price_test_pro", "incomplete", now))
	assertSubStatus(t, st, "sub_3ds", "incomplete", cid)

	// Customer completes 3DS → Stripe activates it.
	postEvent(t, h, "evt_3ds_2", "customer.subscription.updated",
		subEvent("sub_3ds", "cus_3ds", "3ds@buyer.test", cid, "price_test_pro", "active", now))
	assertSubStatus(t, st, "sub_3ds", "active", cid)
}

// ── NEGATIVE: card 4000…0002 — generic decline ──────────────────────

func TestE2E_Card0002_GenericDecline(t *testing.T) {
	h, st, _, cap := newTestHandlerWithEmail(t)
	cid := seedPortalCustomer(t, st, "cust_e2e_decline", "decline@buyer.test")
	now := time.Now().UTC()

	// Declined card → subscription is created but stays incomplete.
	postEvent(t, h, "evt_dec_sub", "customer.subscription.created",
		subEvent("sub_dec", "cus_dec", "decline@buyer.test", cid, "price_test_pro", "incomplete", now))
	w := postEvent(t, h, "evt_dec_inv", "invoice.payment_failed",
		invoiceEvent("in_dec", "cus_dec", "decline@buyer.test", 1, proPriceCents))
	if w.Code != http.StatusOK {
		t.Fatalf("payment_failed webhook: %d", w.Code)
	}

	// Must NOT be active.
	assertSubStatus(t, st, "sub_dec", "incomplete", cid)
	// First-attempt dunning email fired exactly once.
	if n := len(cap.MessagesOfType("payment_failed")); n != 1 {
		t.Errorf("payment_failed emails: got %d want 1", n)
	}
}

// ── NEGATIVE: card 4000…9995 — insufficient funds ───────────────────

func TestE2E_Card9995_InsufficientFunds(t *testing.T) {
	h, st, _, _ := newTestHandlerWithEmail(t)
	cid := seedPortalCustomer(t, st, "cust_e2e_nsf", "nsf@buyer.test")
	now := time.Now().UTC()

	postEvent(t, h, "evt_nsf_sub", "customer.subscription.created",
		subEvent("sub_nsf", "cus_nsf", "nsf@buyer.test", cid, "price_test_pro", "incomplete", now))
	postEvent(t, h, "evt_nsf_inv", "invoice.payment_failed",
		invoiceEvent("in_nsf", "cus_nsf", "nsf@buyer.test", 1, proPriceCents))

	assertSubStatus(t, st, "sub_nsf", "incomplete", cid)
}

// ── NEGATIVE: card 4000…3220 — 3DS required, NOT authenticated ──────

func TestE2E_Card3220_3DSAbandoned(t *testing.T) {
	h, st, _, _ := newTestHandlerWithEmail(t)
	cid := seedPortalCustomer(t, st, "cust_e2e_3dsx", "3dsx@buyer.test")
	now := time.Now().UTC()

	postEvent(t, h, "evt_3dsx_1", "customer.subscription.created",
		subEvent("sub_3dsx", "cus_3dsx", "3dsx@buyer.test", cid, "price_test_pro", "incomplete", now))
	// Customer never completes 3DS; Stripe expires it.
	postEvent(t, h, "evt_3dsx_2", "customer.subscription.updated",
		subEvent("sub_3dsx", "cus_3dsx", "3dsx@buyer.test", cid, "price_test_pro", "incomplete_expired", now))

	assertSubStatus(t, st, "sub_3dsx", "incomplete_expired", cid)
}

// ── NEGATIVE: renewal failure — active → past_due → canceled ────────

func TestE2E_RenewalFailure_PastDueThenCanceled(t *testing.T) {
	h, st, _, cap := newTestHandlerWithEmail(t)
	cid := seedPortalCustomer(t, st, "cust_e2e_renew", "renew@buyer.test")
	now := time.Now().UTC()

	// Healthy active subscription first.
	postEvent(t, h, "evt_rn_1", "customer.subscription.created",
		subEvent("sub_rn", "cus_rn", "renew@buyer.test", cid, "price_test_pro", "active", now))
	assertSubStatus(t, st, "sub_rn", "active", cid)

	// Renewal charge fails → dunning email + Stripe flips to past_due.
	postEvent(t, h, "evt_rn_2", "invoice.payment_failed",
		invoiceEvent("in_rn", "cus_rn", "renew@buyer.test", 1, proPriceCents))
	postEvent(t, h, "evt_rn_3", "customer.subscription.updated",
		subEvent("sub_rn", "cus_rn", "renew@buyer.test", cid, "price_test_pro", "past_due", now))
	assertSubStatus(t, st, "sub_rn", "past_due", cid)

	// Dunning exhausted → subscription deleted → canceled + email.
	postEvent(t, h, "evt_rn_4", "customer.subscription.deleted",
		subEvent("sub_rn", "cus_rn", "renew@buyer.test", cid, "price_test_pro", "canceled", now))
	assertSubStatus(t, st, "sub_rn", "canceled", cid)

	if n := len(cap.MessagesOfType("payment_failed")); n != 1 {
		t.Errorf("payment_failed emails: got %d want 1", n)
	}
	if n := len(cap.MessagesOfType("subscription_canceled")); n != 1 {
		t.Errorf("subscription_canceled emails: got %d want 1", n)
	}
}

// ── NEGATIVE: duplicate webhook delivery is idempotent ──────────────

func TestE2E_DuplicateWebhookDelivery(t *testing.T) {
	h, st, _, _ := newTestHandlerWithEmail(t)
	cid := seedPortalCustomer(t, st, "cust_e2e_dup", "dup@buyer.test")
	now := time.Now().UTC()
	ev := subEvent("sub_dup", "cus_dup", "dup@buyer.test", cid, "price_test_pro", "active", now)

	postEvent(t, h, "evt_dup_same", "customer.subscription.created", ev)
	postEvent(t, h, "evt_dup_same", "customer.subscription.created", ev) // replay same event id

	subs, _ := st.ListSubscriptionsByCustomer(context.Background(), cid)
	if len(subs) != 1 {
		t.Errorf("idempotency: got %d subscriptions want 1", len(subs))
	}
}

// ── NEGATIVE: out-of-order delivery (updated before created) ────────

func TestE2E_OutOfOrderDelivery(t *testing.T) {
	h, st, _, _ := newTestHandlerWithEmail(t)
	cid := seedPortalCustomer(t, st, "cust_e2e_ooo", "ooo@buyer.test")
	now := time.Now().UTC()

	// "updated" arrives first — handler treats latest state as truth and
	// upserts, so the subscription still lands correctly.
	postEvent(t, h, "evt_ooo_1", "customer.subscription.updated",
		subEvent("sub_ooo", "cus_ooo", "ooo@buyer.test", cid, "price_test_pro", "active", now))
	assertSubStatus(t, st, "sub_ooo", "active", cid)

	// A late "created" with stale (trialing) state must not resurrect it
	// below active if it carries older info — here we just assert no
	// duplicate row is produced.
	postEvent(t, h, "evt_ooo_2", "customer.subscription.created",
		subEvent("sub_ooo", "cus_ooo", "ooo@buyer.test", cid, "price_test_pro", "active", now))
	subs, _ := st.ListSubscriptionsByCustomer(context.Background(), cid)
	if len(subs) != 1 {
		t.Errorf("out-of-order produced %d rows, want 1", len(subs))
	}
}
