package stripebill

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

// CheckoutHandler issues Stripe Checkout Session URLs for self-serve
// signup. The customer's browser is redirected there from the pricing
// page; payment success bounces back to <app_url>/welcome with
// ?session_id=cs_..., and the matching subscription appears via the
// customer.subscription.created webhook within a few seconds.
//
// This handler is intentionally minimal — Stripe's hosted checkout
// page handles card UI, 3DS, fraud, SCA, etc. We just construct the
// session.
type CheckoutHandler struct {
	apiKey  string
	appURL  string // base URL of app.infrays.org (where success/cancel redirects land)
	priceMap *TierMapping
}

// NewCheckoutHandler returns a handler ready to serve
// POST /v1/checkout/create-session.
//
// apiKey is the Stripe secret key (env NP_STRIPE_SECRET_KEY).
// appURL is the public portal URL — used for success / cancel
// redirect URLs.
func NewCheckoutHandler(apiKey, appURL string, priceMap *TierMapping) (*CheckoutHandler, error) {
	if apiKey == "" {
		return nil, errors.New("stripebill: empty Stripe API key — refusing to register checkout handler")
	}
	if appURL == "" {
		appURL = "https://app.infrays.org"
	}
	// stripe-go's API key is set as a global. Setting it here is
	// idempotent for repeated calls.
	stripe.Key = apiKey
	if priceMap == nil {
		priceMap = NewTierMapping(nil)
	}
	return &CheckoutHandler{
		apiKey:   apiKey,
		appURL:   strings.TrimRight(appURL, "/"),
		priceMap: priceMap,
	}, nil
}

// createSessionReq is the POST body shape.
type createSessionReq struct {
	PriceID       string `json:"price_id"`        // REQUIRED — must be in cfg.PriceMap
	CustomerEmail string `json:"customer_email"`  // pre-fill on Checkout; optional
	TrialDays     int    `json:"trial_days"`      // 0 = no trial. Cap at 30.
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type createSessionResp struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// ServeHTTP handles POST /v1/checkout/create-session.
func (h *CheckoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		http.Error(w, `{"error":"read"}`, http.StatusBadRequest)
		return
	}
	var req createSessionReq
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if req.PriceID == "" {
		http.Error(w, `{"error":"price_id required"}`, http.StatusBadRequest)
		return
	}
	// Verify the price is known to our tier map. Stops drive-by abuse
	// where someone passes a malicious price_id to test our Stripe key.
	if _, ok := h.priceMap.Lookup(req.PriceID); !ok && len(h.priceMap.byPriceID) > 0 {
		http.Error(w, `{"error":"unknown price_id"}`, http.StatusBadRequest)
		return
	}
	if req.TrialDays > 30 {
		req.TrialDays = 30
	}

	sessionID, url, err := h.createSession(req.PriceID, req.CustomerEmail, req.TrialDays, req.Metadata)
	if err != nil {
		http.Error(w, `{"error":"stripe error","detail":"`+err.Error()+`"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(createSessionResp{SessionID: sessionID, URL: url})
}

// createSession builds and creates a Stripe Checkout Session for a
// known Price ID. Shared by the public ServeHTTP path (pricing page)
// and CreateCheckoutSession (authenticated portal path). Callers are
// responsible for validating the priceID against the tier map.
func (h *CheckoutHandler) createSession(priceID, customerEmail string, trialDays int, metadata map[string]string) (sessionID, url string, err error) {
	successURL := h.appURL + "/welcome?session_id={CHECKOUT_SESSION_ID}"
	cancelURL := h.appURL + "/pricing?canceled=1"

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(priceID), Quantity: stripe.Int64(1)},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}
	if customerEmail != "" {
		params.CustomerEmail = stripe.String(customerEmail)
	}
	// Propagate metadata to BOTH the session and the resulting
	// subscription. Subscription-level metadata is what the webhook
	// reads to bind the subscription to an existing local customer
	// (customer_id), so it must survive onto the Subscription object.
	if trialDays > 0 || len(metadata) > 0 {
		sd := &stripe.CheckoutSessionSubscriptionDataParams{}
		if trialDays > 0 {
			sd.TrialPeriodDays = stripe.Int64(int64(trialDays))
		}
		if len(metadata) > 0 {
			sd.Metadata = metadata
		}
		params.SubscriptionData = sd
	}
	if len(metadata) > 0 {
		params.Metadata = metadata
	}

	_ = context.Background() // hooks for future tracing
	sess, err := session.New(params)
	if err != nil {
		return "", "", err
	}
	return sess.ID, sess.URL, nil
}

// CreateCheckoutSession issues a Checkout Session URL for an
// authenticated portal customer who picked a tier + interval. The
// price is resolved server-side from the tier map (the client never
// supplies a raw Price ID), and customerID is stamped into session
// metadata so the webhook binds the resulting subscription to the
// existing customer instead of creating a duplicate.
//
// No trial days are granted here — the customer already consumed the
// signup trial; this is a paid conversion. Satisfies
// portal.CheckoutCreator.
func (h *CheckoutHandler) CreateCheckoutSession(tier, interval, customerEmail, customerID string) (string, error) {
	priceID, ok := h.priceMap.LookupTier(tier, interval)
	if !ok {
		return "", errors.New("stripebill: no Stripe price configured for tier/interval")
	}
	meta := map[string]string{}
	if customerID != "" {
		meta["customer_id"] = customerID
	}
	_, url, err := h.createSession(priceID, customerEmail, 0, meta)
	if err != nil {
		return "", err
	}
	return url, nil
}
