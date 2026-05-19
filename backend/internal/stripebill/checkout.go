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

	successURL := h.appURL + "/welcome?session_id={CHECKOUT_SESSION_ID}"
	cancelURL := h.appURL + "/pricing?canceled=1"

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(req.PriceID), Quantity: stripe.Int64(1)},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}
	if req.CustomerEmail != "" {
		params.CustomerEmail = stripe.String(req.CustomerEmail)
	}
	if req.TrialDays > 0 {
		params.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
			TrialPeriodDays: stripe.Int64(int64(req.TrialDays)),
		}
	}
	if len(req.Metadata) > 0 {
		params.Metadata = req.Metadata
	}

	_ = context.Background() // hooks for future tracing
	sess, err := session.New(params)
	if err != nil {
		http.Error(w, `{"error":"stripe error","detail":"`+err.Error()+`"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(createSessionResp{
		SessionID: sess.ID,
		URL:       sess.URL,
	})
}
