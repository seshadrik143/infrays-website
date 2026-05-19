package stripebill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/store"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

// Config holds runtime knobs the issuer's Stripe handler needs.
type Config struct {
	// WebhookSecret is the signing secret Stripe gives you when you
	// configure the webhook endpoint. Set via env
	// NP_STRIPE_WEBHOOK_SECRET. Required — without it we cannot
	// verify event authenticity and MUST reject all requests.
	WebhookSecret string

	// PriceMap maps Stripe Price IDs to local tiers. Set via env
	// NP_STRIPE_PRICE_MAP (see ParseTierMappingFromEnv).
	PriceMap *TierMapping

	// MaxEventAge bounds replay-protection. Stripe webhooks include
	// a timestamp; older events are rejected even if the signature
	// is valid. Default 5 minutes (Stripe's recommendation).
	MaxEventAge time.Duration
}

// Handler wires the Stripe webhook into the issuer's store + audit.
type Handler struct {
	cfg   Config
	store store.Store
	audit audit.Log
}

// NewHandler validates config + returns a ready Handler. Returns an
// error when WebhookSecret is empty — refuse to start a half-configured
// handler that would silently accept unsigned events.
func NewHandler(cfg Config, st store.Store, al audit.Log) (*Handler, error) {
	if cfg.WebhookSecret == "" {
		return nil, errors.New("stripebill: empty webhook secret — refusing to register handler")
	}
	if cfg.MaxEventAge == 0 {
		cfg.MaxEventAge = 5 * time.Minute
	}
	if cfg.PriceMap == nil {
		cfg.PriceMap = NewTierMapping(nil)
	}
	return &Handler{cfg: cfg, store: st, audit: al}, nil
}

// ServeHTTP handles POST /v1/webhooks/stripe.
//
// Response codes:
//
//	200  Event processed (or successfully deduped as already-seen)
//	400  Signature invalid, body unparseable, or event too old
//	500  Transient internal error — Stripe retries with backoff
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MiB cap
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	sig := r.Header.Get("Stripe-Signature")
	if sig == "" {
		http.Error(w, "missing signature", http.StatusBadRequest)
		return
	}

	// Verify signature + reject events older than MaxEventAge.
	// stripe-go's ConstructEventWithOptions honors a tolerance window
	// (defaults to 5min if we pass 300). We use cfg.MaxEventAge.
	event, err := webhook.ConstructEventWithOptions(body, sig, h.cfg.WebhookSecret, webhook.ConstructEventOptions{
		Tolerance: h.cfg.MaxEventAge,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid signature: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Idempotency: insert by event ID; ErrAlreadyExists means we've
	// seen this one before. Return 200 immediately so Stripe stops
	// retrying.
	err = h.store.RecordWebhookEvent(ctx, &store.WebhookEvent{
		ID:         event.ID,
		Type:       string(event.Type),
		Payload:    body,
		ReceivedAt: time.Now().UTC(),
		Status:     "pending",
	})
	if errors.Is(err, store.ErrAlreadyExists) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"already-processed"}`))
		return
	}
	if err != nil {
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	// Process the event. Errors here become 500 so Stripe retries —
	// the event record is left in "pending" / "failed" status for
	// operator inspection.
	if err := h.dispatch(ctx, &event); err != nil {
		_ = h.store.MarkWebhookProcessed(ctx, event.ID, "failed", err.Error())
		_, _ = h.audit.Append(ctx, audit.Entry{
			EventType: "stripe.webhook.error",
			Actor:     "system",
			Payload: map[string]any{
				"event_id":   event.ID,
				"event_type": string(event.Type),
				"error":      err.Error(),
			},
		})
		http.Error(w, "processing failed", http.StatusInternalServerError)
		return
	}

	_ = h.store.MarkWebhookProcessed(ctx, event.ID, "processed", "")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// dispatch routes verified Stripe events to per-type handlers. Unknown
// types are silently accepted (recorded in webhook_events but not
// processed) so Stripe doesn't keep retrying events we don't care
// about (e.g. invoice.created when we only listen for invoice.paid).
func (h *Handler) dispatch(ctx context.Context, event *stripe.Event) error {
	switch event.Type {
	case "customer.created":
		return h.handleCustomerCreated(ctx, event)
	case "customer.subscription.created":
		return h.handleSubscriptionEvent(ctx, event, "created")
	case "customer.subscription.updated":
		return h.handleSubscriptionEvent(ctx, event, "updated")
	case "customer.subscription.deleted":
		return h.handleSubscriptionEvent(ctx, event, "deleted")
	case "invoice.paid":
		return h.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		return h.handleInvoicePaymentFailed(ctx, event)
	default:
		// Unhandled — audit only.
		_, _ = h.audit.Append(ctx, audit.Entry{
			EventType: "stripe.webhook.ignored",
			Actor:     "system",
			Payload:   map[string]any{"event_id": event.ID, "type": string(event.Type)},
		})
		return nil
	}
}

// stripeEventTime returns the event's Created time as time.Time.
// Stripe events have an integer Created field (Unix seconds).
func stripeEventTime(event *stripe.Event) time.Time {
	if event.Created == 0 {
		return time.Now().UTC()
	}
	return time.Unix(event.Created, 0).UTC()
}

// unmarshalStripeData decodes the event's Data.Raw into a typed
// Stripe object. Convenience wrapper — Stripe's SDK exposes Data.Raw
// as []byte and leaves it to the caller to know the target type.
func unmarshalStripeData(event *stripe.Event, dst any) error {
	if len(event.Data.Raw) == 0 {
		return errors.New("stripebill: empty event data")
	}
	return json.Unmarshal(event.Data.Raw, dst)
}
