package portal

import (
	"net/http"
)

// handleBillingPortalURL creates a fresh Stripe Billing Portal session
// for the authenticated customer and returns its URL. The browser
// follows the redirect, the customer manages their subscription on
// Stripe's hosted page, and Stripe sends them back to AppURL/account
// when they're done.
//
// Customers without a Stripe customer ID (offline / sales-issued)
// get 409 Conflict — there's no Stripe-side state to portal into.
func (s *Server) handleBillingPortalURL(w http.ResponseWriter, r *http.Request) {
	cust := customerFromContext(r.Context())
	if cust.StripeCustomerID == "" {
		writeError(w, http.StatusConflict, "no stripe customer — contact support@infrays.org for offline billing changes")
		return
	}
	url, err := s.cfg.BillingPortal.CreateSession(cust.StripeCustomerID)
	if err != nil {
		writeError(w, http.StatusBadGateway, "stripe portal session failed")
		return
	}
	s.appendAudit("portal.billing_portal_session_created", cust, nil)
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}
