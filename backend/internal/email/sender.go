// Package email is the transactional-email surface for the issuer.
//
// Three implementations satisfy the Sender interface:
//
//   - Postmark  — production. Set NP_POSTMARK_SERVER_TOKEN to enable.
//   - Noop      — default when no provider is configured. Logs and
//                 drops the message. Lets the issuer run without
//                 email credentials during pre-launch.
//   - Capture   — test-only. Stores sends in memory so tests can
//                 assert recipient / subject / body.
//
// All Sender calls are best-effort from the caller's perspective:
// email failures are logged and audited but never block the request
// that triggered them (e.g. Stripe webhook processing). Long Postmark
// timeouts won't stall the customer's checkout.
package email

import "context"

// Message is what callers hand the Sender. Both HTML and text bodies
// are required — Postmark sends them as multipart, and clients pick
// the best one. If you only have one, supply the same content twice
// (the Sender does not auto-strip HTML).
type Message struct {
	To          string // RFC 5322 address; "Name <addr@x.com>" supported
	Subject     string
	HTMLBody    string
	TextBody    string
	MessageType string // "trial_expiring", "payment_failed", etc. — for logging + telemetry
}

// Sender is the abstraction every caller uses. Construct via
// NewSenderFromEnv() in main; tests construct NewCaptureSender()
// directly.
type Sender interface {
	// Send delivers the message. Returns error on transport failure
	// (network, 5xx from provider, malformed input). Callers should
	// log + audit the error but never propagate it up the API call
	// chain — email delivery is not load-bearing.
	Send(ctx context.Context, msg Message) error

	// Name returns a short identifier for the active implementation,
	// for startup logging ("postmark", "noop", "capture").
	Name() string
}
