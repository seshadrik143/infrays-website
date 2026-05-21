package stripebill

import (
	"errors"
	"strings"

	stripe "github.com/stripe/stripe-go/v82"
	billingportal "github.com/stripe/stripe-go/v82/billingportal/session"
)

// PortalSessionCreator issues Stripe Billing Portal Session URLs.
// Customers click "Manage subscription" in the portal and get
// redirected to Stripe's hosted portal where they can update cards,
// cancel, view invoices, etc.
//
// Separate from CheckoutHandler so the issuer can run with one
// configured but not the other (e.g. checkout disabled post-launch).
type PortalSessionCreator struct {
	apiKey    string
	returnURL string // where Stripe sends them after the portal session ends
}

func NewPortalSessionCreator(apiKey, appURL string) (*PortalSessionCreator, error) {
	if apiKey == "" {
		return nil, errors.New("stripebill: empty Stripe API key for portal session creator")
	}
	stripe.Key = apiKey
	if appURL == "" {
		appURL = "https://app.infrays.org"
	}
	return &PortalSessionCreator{
		apiKey:    apiKey,
		returnURL: strings.TrimRight(appURL, "/") + "/account",
	}, nil
}

// CreateSession returns the Stripe-hosted portal URL for the given
// Stripe customer ID. Caller is responsible for resolving
// stripeCustomerID from the authenticated user (e.g. via the portal
// session → store.Customer.StripeCustomerID).
func (p *PortalSessionCreator) CreateSession(stripeCustomerID string) (string, error) {
	if stripeCustomerID == "" {
		return "", errors.New("stripebill: empty stripe_customer_id")
	}
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeCustomerID),
		ReturnURL: stripe.String(p.returnURL),
	}
	sess, err := billingportal.New(params)
	if err != nil {
		return "", err
	}
	return sess.URL, nil
}
