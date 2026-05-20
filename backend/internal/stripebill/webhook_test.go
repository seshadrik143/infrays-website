package stripebill

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/email"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

const testWebhookSecret = "whsec_test_secret_for_unit_tests_only"

// signStripeEvent computes the Stripe-Signature header value for a
// JSON body. Mirrors what stripe-go's webhook package expects:
//
//	t=<unix-ts>,v1=<hex hmac-sha256 of "<ts>.<body>" keyed by secret>
//
// stripe-go also accepts v0 (deprecated) — we only generate v1.
func signStripeEvent(t *testing.T, body []byte, ts time.Time) string {
	t.Helper()
	tsStr := fmt.Sprintf("%d", ts.Unix())
	signedPayload := tsStr + "." + string(body)
	mac := hmac.New(sha256.New, []byte(testWebhookSecret))
	mac.Write([]byte(signedPayload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return "t=" + tsStr + ",v1=" + sig
}

func newTestHandler(t *testing.T) (*Handler, store.Store, audit.Log) {
	t.Helper()
	h, st, al, _ := newTestHandlerWithEmail(t)
	return h, st, al
}

// newTestHandlerWithEmail also returns the CaptureSender so tests
// can assert on the emails the handler would have sent.
func newTestHandlerWithEmail(t *testing.T) (*Handler, store.Store, audit.Log, *email.CaptureSender) {
	t.Helper()
	st := store.NewMemory()
	al := audit.NewMemory()
	cap := email.NewCaptureSender()
	pm := NewTierMapping(map[string]TierConfig{
		"price_test_pro": {Tier: "professional", EntitlementSetID: "professional-v1"},
		"price_test_ent": {Tier: "enterprise", EntitlementSetID: "enterprise-v1"},
	})
	h, err := NewHandler(Config{
		WebhookSecret: testWebhookSecret,
		PriceMap:      pm,
		Email:         cap,
		AppURL:        "https://app.test",
	}, st, al)
	if err != nil {
		t.Fatal(err)
	}
	return h, st, al, cap
}

// expectedStripeAPIVersion must match what the linked stripe-go SDK
// expects (it's a constant inside the SDK). Real Stripe webhooks
// include this in the event body; without it stripe-go's
// ConstructEventWithOptions rejects the event with a version-mismatch
// error. The exact value moves with SDK upgrades — keep this string
// in sync with the major version bump.
const expectedStripeAPIVersion = "2025-08-27.basil"

func postEvent(t *testing.T, h http.Handler, eventID, eventType string, data any) *httptest.ResponseRecorder {
	t.Helper()
	dataJSON, _ := json.Marshal(data)
	body, _ := json.Marshal(map[string]any{
		"id":          eventID,
		"object":      "event",
		"type":        eventType,
		"created":     time.Now().Unix(),
		"api_version": expectedStripeAPIVersion,
		"data":        map[string]any{"object": json.RawMessage(dataJSON)},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", signStripeEvent(t, body, time.Now()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

// ── tests ──────────────────────────────────────────────────────────

func TestWebhook_RejectsMissingSignature(t *testing.T) {
	h, _, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/stripe", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d want 400", w.Code)
	}
}

func TestWebhook_RejectsBadSignature(t *testing.T) {
	h, _, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/stripe", bytes.NewReader([]byte(`{"id":"evt_x","type":"foo","created":0,"data":{}}`)))
	req.Header.Set("Stripe-Signature", "t=1,v1=deadbeef")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d want 400 (bad sig)", w.Code)
	}
}

func TestWebhook_IdempotentReplay(t *testing.T) {
	h, st, _ := newTestHandler(t)
	// Seed a customer to attach the subscription to.
	now := time.Now().UTC()
	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust_local_1", Email: "acme@test", StripeCustomerID: "cus_stripe_1",
		Status: "active", CreatedAt: now, UpdatedAt: now,
	})

	sub := map[string]any{
		"id":       "sub_stripe_1",
		"customer": map[string]any{"id": "cus_stripe_1", "email": "acme@test"},
		"status":   "active",
		"items": map[string]any{
			"data": []map[string]any{
				{
					"price":                  map[string]any{"id": "price_test_pro"},
					"current_period_start":   now.Unix(),
					"current_period_end":     now.Add(30 * 24 * time.Hour).Unix(),
				},
			},
		},
	}

	w1 := postEvent(t, h, "evt_dedup_001", "customer.subscription.created", sub)
	if w1.Code != http.StatusOK {
		t.Fatalf("first post: got %d body=%s", w1.Code, w1.Body.String())
	}

	w2 := postEvent(t, h, "evt_dedup_001", "customer.subscription.created", sub)
	if w2.Code != http.StatusOK {
		t.Fatalf("replay: got %d", w2.Code)
	}
	if !bytes.Contains(w2.Body.Bytes(), []byte("already-processed")) {
		t.Errorf("replay body should include 'already-processed', got %s", w2.Body.String())
	}

	// Only ONE subscription row should exist despite two posts.
	subs, _ := st.ListSubscriptionsByCustomer(context.Background(), "cust_local_1")
	if len(subs) != 1 {
		t.Errorf("subscription count after replay: got %d want 1", len(subs))
	}
}

func TestWebhook_CustomerCreated_NewCustomer(t *testing.T) {
	h, st, _ := newTestHandler(t)
	customer := map[string]any{
		"id":    "cus_new_001",
		"email": "newcustomer@test.com",
		"name":  "New Customer",
	}
	w := postEvent(t, h, "evt_cust_new_001", "customer.created", customer)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
	c, err := st.GetCustomerByEmail(context.Background(), "newcustomer@test.com")
	if err != nil {
		t.Fatalf("customer not created: %v", err)
	}
	if c.StripeCustomerID != "cus_new_001" {
		t.Errorf("stripe id: got %q", c.StripeCustomerID)
	}
	if c.Name != "New Customer" {
		t.Errorf("name: got %q", c.Name)
	}
}

func TestWebhook_CustomerCreated_AttachToExisting(t *testing.T) {
	h, st, _ := newTestHandler(t)
	// Pre-existing local customer (portal signup BEFORE first checkout).
	now := time.Now().UTC()
	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust_existing_001", Email: "existing@test.com",
		Status: "active", CreatedAt: now, UpdatedAt: now,
	})

	customer := map[string]any{
		"id":    "cus_stripe_existing_001",
		"email": "existing@test.com",
		"name":  "Existing User",
	}
	w := postEvent(t, h, "evt_cust_attach_001", "customer.created", customer)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}
	c, _ := st.GetCustomerByEmail(context.Background(), "existing@test.com")
	if c.StripeCustomerID != "cus_stripe_existing_001" {
		t.Errorf("stripe id not attached: %q", c.StripeCustomerID)
	}
	if c.ID != "cust_existing_001" {
		t.Errorf("should reuse existing customer row, got %q", c.ID)
	}
}

func TestWebhook_SubscriptionLifecycle(t *testing.T) {
	h, st, _ := newTestHandler(t)
	now := time.Now().UTC()
	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust_lc", Email: "lifecycle@test", StripeCustomerID: "cus_lc",
		Status: "active", CreatedAt: now, UpdatedAt: now,
	})

	// created
	sub := map[string]any{
		"id":       "sub_lc",
		"customer": map[string]any{"id": "cus_lc", "email": "lifecycle@test"},
		"status":   "active",
		"items": map[string]any{
			"data": []map[string]any{
				{
					"price":                map[string]any{"id": "price_test_pro"},
					"current_period_start": now.Unix(),
					"current_period_end":   now.Add(30 * 24 * time.Hour).Unix(),
				},
			},
		},
	}
	w := postEvent(t, h, "evt_lc_001", "customer.subscription.created", sub)
	if w.Code != http.StatusOK {
		t.Fatalf("created: got %d body=%s", w.Code, w.Body.String())
	}
	got, _ := st.GetSubscriptionByStripeID(context.Background(), "sub_lc")
	if got == nil || got.Tier != "professional" {
		t.Fatalf("subscription not stored as professional: %+v", got)
	}

	// updated: tier change to enterprise
	sub["items"] = map[string]any{
		"data": []map[string]any{
			{
				"price":                map[string]any{"id": "price_test_ent"},
				"current_period_start": now.Unix(),
				"current_period_end":   now.Add(30 * 24 * time.Hour).Unix(),
			},
		},
	}
	w = postEvent(t, h, "evt_lc_002", "customer.subscription.updated", sub)
	if w.Code != http.StatusOK {
		t.Fatalf("updated: got %d", w.Code)
	}
	got, _ = st.GetSubscriptionByStripeID(context.Background(), "sub_lc")
	if got.Tier != "enterprise" {
		t.Errorf("tier after update: got %q want enterprise", got.Tier)
	}

	// deleted
	sub["status"] = "canceled"
	sub["canceled_at"] = now.Add(1 * time.Hour).Unix()
	w = postEvent(t, h, "evt_lc_003", "customer.subscription.deleted", sub)
	if w.Code != http.StatusOK {
		t.Fatalf("deleted: got %d", w.Code)
	}
	got, _ = st.GetSubscriptionByStripeID(context.Background(), "sub_lc")
	if got.Status != "canceled" {
		t.Errorf("status after delete: got %q want canceled", got.Status)
	}
}

func TestWebhook_UnknownPriceFallsBackToFree(t *testing.T) {
	h, st, _ := newTestHandler(t)
	now := time.Now().UTC()
	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust_unk", Email: "unk@test", StripeCustomerID: "cus_unk",
		Status: "active", CreatedAt: now, UpdatedAt: now,
	})

	sub := map[string]any{
		"id":       "sub_unk",
		"customer": map[string]any{"id": "cus_unk", "email": "unk@test"},
		"status":   "active",
		"items": map[string]any{
			"data": []map[string]any{
				{
					"price":                map[string]any{"id": "price_typo_999"},
					"current_period_start": now.Unix(),
					"current_period_end":   now.Add(30 * 24 * time.Hour).Unix(),
				},
			},
		},
	}
	w := postEvent(t, h, "evt_unk_001", "customer.subscription.created", sub)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
	got, _ := st.GetSubscriptionByStripeID(context.Background(), "sub_unk")
	if got.Tier != "free" {
		t.Errorf("fallback tier: got %q want free", got.Tier)
	}
}

func TestWebhook_UnknownEventType_NoOp(t *testing.T) {
	h, st, _ := newTestHandler(t)
	w := postEvent(t, h, "evt_unknown", "tax_rate.created", map[string]any{"id": "txr_001"})
	if w.Code != http.StatusOK {
		t.Errorf("unknown event type should still 200, got %d", w.Code)
	}
	// No side effects expected
	_, err := st.GetWebhookEvent(context.Background(), "evt_unknown")
	if err != nil {
		t.Errorf("event record should exist: %v", err)
	}
}

func TestTierMapping_ParseFromEnv(t *testing.T) {
	tm, err := ParseTierMappingFromEnv("price_a=professional:professional-v1, price_b=enterprise:enterprise-v1")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := tm.Lookup("price_a")
	if a.Tier != "professional" || a.EntitlementSetID != "professional-v1" {
		t.Errorf("got %+v", a)
	}
	b, _ := tm.Lookup("price_b")
	if b.Tier != "enterprise" {
		t.Errorf("got %+v", b)
	}

	// Empty input → empty mapping, no error
	empty, err := ParseTierMappingFromEnv("")
	if err != nil {
		t.Errorf("empty input: %v", err)
	}
	if empty == nil {
		t.Errorf("empty mapping should not be nil")
	}

	// Bad input → error
	if _, err := ParseTierMappingFromEnv("garbage"); err == nil {
		t.Errorf("expected error for malformed input")
	}
}

func TestCheckout_RejectsMissingPriceID(t *testing.T) {
	h, err := NewCheckoutHandler("sk_test_fake", "https://app.test", NewTierMapping(nil))
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/checkout/create-session", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
}

func TestCheckout_RejectsUnknownPriceID(t *testing.T) {
	pm := NewTierMapping(map[string]TierConfig{
		"price_known": {Tier: "professional", EntitlementSetID: "professional-v1"},
	})
	h, _ := NewCheckoutHandler("sk_test_fake", "https://app.test", pm)
	req := httptest.NewRequest(http.MethodPost, "/v1/checkout/create-session", bytes.NewReader([]byte(`{"price_id":"price_unknown"}`)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
}

func TestHandler_RejectsEmptyWebhookSecret(t *testing.T) {
	_, err := NewHandler(Config{WebhookSecret: ""}, store.NewMemory(), audit.NewMemory())
	if err == nil {
		t.Error("expected error for empty webhook secret")
	}
}

// ── email-trigger tests ────────────────────────────────────────────

func TestEmail_CustomerCreatedSendsWelcome(t *testing.T) {
	h, _, _, cap := newTestHandlerWithEmail(t)
	customer := map[string]any{
		"id":    "cus_new_e2e",
		"email": "newcust@test.com",
		"name":  "New Customer",
	}
	w := postEvent(t, h, "evt_email_001", "customer.created", customer)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", w.Code, w.Body.String())
	}
	sent := cap.MessagesOfType("welcome")
	if len(sent) != 1 {
		t.Fatalf("welcome emails: got %d, want 1", len(sent))
	}
	if sent[0].To != "newcust@test.com" {
		t.Errorf("welcome To: got %q", sent[0].To)
	}
	if !contains(sent[0].HTMLBody, "New Customer") {
		t.Errorf("welcome body missing name")
	}
}

func TestEmail_SubscriptionDeletedSendsCancelNotification(t *testing.T) {
	h, st, _, cap := newTestHandlerWithEmail(t)
	now := time.Now().UTC()
	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust_cancel", Email: "canceler@test.com", Name: "Canceler",
		StripeCustomerID: "cus_cancel", Status: "active",
		CreatedAt: now, UpdatedAt: now,
	})

	sub := map[string]any{
		"id":          "sub_cancel",
		"customer":    map[string]any{"id": "cus_cancel", "email": "canceler@test.com"},
		"status":      "active",
		"canceled_at": now.Unix(),
		"items": map[string]any{
			"data": []map[string]any{
				{
					"price":                map[string]any{"id": "price_test_pro"},
					"current_period_start": now.Unix(),
					"current_period_end":   now.Add(30 * 24 * time.Hour).Unix(),
				},
			},
		},
	}
	// First create the subscription
	postEvent(t, h, "evt_cancel_001", "customer.subscription.created", sub)
	cap.Reset() // ignore any welcome / creation emails

	// Now delete it
	w := postEvent(t, h, "evt_cancel_002", "customer.subscription.deleted", sub)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", w.Code, w.Body.String())
	}
	sent := cap.MessagesOfType("subscription_canceled")
	if len(sent) != 1 {
		t.Fatalf("cancel emails: got %d, want 1", len(sent))
	}
	if sent[0].To != "canceler@test.com" {
		t.Errorf("cancel To: got %q", sent[0].To)
	}
}

func TestEmail_PaymentFailedFirstAttemptOnly(t *testing.T) {
	h, st, _, cap := newTestHandlerWithEmail(t)
	now := time.Now().UTC()
	_ = st.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust_pf", Email: "payfail@test.com", Name: "Pay Fail",
		StripeCustomerID: "cus_pf", Status: "active",
		CreatedAt: now, UpdatedAt: now,
	})

	// First failed attempt → email
	invoice1 := map[string]any{
		"id":             "in_001",
		"amount_due":     4900,
		"attempt_count":  1,
		"customer":       map[string]any{"id": "cus_pf", "email": "payfail@test.com"},
	}
	w := postEvent(t, h, "evt_pf_001", "invoice.payment_failed", invoice1)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", w.Code, w.Body.String())
	}
	if got := cap.MessagesOfType("payment_failed"); len(got) != 1 {
		t.Fatalf("first attempt: got %d emails, want 1", len(got))
	}

	// Second + third attempts (dunning retries) → no email
	invoice2 := map[string]any{
		"id":             "in_001",
		"amount_due":     4900,
		"attempt_count":  2,
		"customer":       map[string]any{"id": "cus_pf", "email": "payfail@test.com"},
	}
	cap.Reset()
	postEvent(t, h, "evt_pf_002", "invoice.payment_failed", invoice2)
	if got := cap.MessagesOfType("payment_failed"); len(got) != 0 {
		t.Errorf("second attempt: got %d emails, want 0", len(got))
	}
}

// contains is a local helper because strings.Contains isn't imported.
func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}

func TestCheckoutHandler_RejectsEmptyAPIKey(t *testing.T) {
	_, err := NewCheckoutHandler("", "https://app.test", NewTierMapping(nil))
	if err == nil {
		t.Error("expected error for empty API key")
	}
}
