package portal

import (
	"encoding/json"
	"net/http"
	"strings"
)

// checkoutRequest is the POST body for self-serve checkout. The client
// names a tier + billing interval; it never supplies a raw Stripe
// Price ID — the server resolves that from the tier map so a malicious
// client can't drive checkout into an arbitrary price.
type checkoutRequest struct {
	Tier     string `json:"tier"`     // "starter" | "pro" (resolved server-side)
	Interval string `json:"interval"` // "month" | "annual"; empty → month
}

// handleListPlans returns the purchasable tiers configured on the
// issuer so the picker only offers plans that resolve to a real Stripe
// price. Returns an empty list (not an error) when none are configured.
func (s *Server) handleListPlans(w http.ResponseWriter, r *http.Request) {
	plans := s.cfg.Plans
	if plans == nil {
		plans = []PlanOption{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"plans": plans})
}

// handleCreateCheckoutSession starts a Stripe Checkout Session for the
// authenticated customer and returns its URL. The browser follows the
// redirect to Stripe's hosted checkout; on success Stripe bounces back
// to AppURL/welcome and the subscription lands via webhook.
//
// The customer's email is taken from the session (not the request) so
// it can't be spoofed, and the customer ID is passed through to the
// session metadata so the webhook binds the new subscription to this
// existing customer rather than creating a duplicate.
func (s *Server) handleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Tier = strings.TrimSpace(strings.ToLower(req.Tier))
	req.Interval = strings.TrimSpace(strings.ToLower(req.Interval))
	if req.Tier == "" {
		writeError(w, http.StatusBadRequest, "tier required")
		return
	}
	if req.Interval != "" && req.Interval != "month" && req.Interval != "annual" {
		writeError(w, http.StatusBadRequest, "interval must be month or annual")
		return
	}

	cust := customerFromContext(r.Context())
	// Require a verified email before taking money — mirrors the
	// enrollment-token gate and stops checkout on unconfirmed accounts.
	if cust.EmailVerifiedAt.IsZero() {
		writeError(w, http.StatusForbidden, "verify your email before subscribing")
		return
	}

	url, err := s.cfg.Checkout.CreateCheckoutSession(req.Tier, req.Interval, cust.Email, cust.ID)
	if err != nil {
		// Unknown tier/interval is a client problem; everything else
		// is an upstream Stripe failure.
		if strings.Contains(err.Error(), "no Stripe price configured") {
			writeError(w, http.StatusBadRequest, "unknown plan")
			return
		}
		writeError(w, http.StatusBadGateway, "stripe checkout session failed")
		return
	}

	s.appendAudit("portal.checkout_session_created", cust, map[string]any{
		"tier":     req.Tier,
		"interval": req.Interval,
	})
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}
