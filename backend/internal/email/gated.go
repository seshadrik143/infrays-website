package email

import (
	"context"
	"log/slog"
	"net/mail"
	"strings"
)

// DomainGatedSender wraps another Sender and only forwards messages
// whose recipient domain is on the allowlist. Anything else is logged
// and dropped (no error returned — same contract as NoopSender).
//
// Intended for use during pre-launch periods when the upstream provider
// (e.g. Postmark) is in "pending approval" state and rejects sends to
// any domain other than the sender's own. Without this gate, every
// signup attempt from a Gmail/Outlook user would call Postmark, eat a
// 412 error, and pollute the logs.
//
// Wire via the NP_EMAIL_ALLOWED_DOMAINS env var (comma-separated). When
// unset, the factory uses the underlying sender directly with no gate.
//
// After Postmark approval (or whatever the upstream restriction was)
// lifts, unset the env var / Fly secret and the gate disappears.
type DomainGatedSender struct {
	inner    Sender
	allowed  map[string]struct{} // lowercase, no leading "@"
	rendered string              // pre-rendered "infrays.org, foo.com" for logs
}

// NewDomainGatedSender wraps inner so that only recipients whose
// domain is in allowedDomains are forwarded. allowedDomains is a
// comma- or whitespace-separated list; case-insensitive; leading "@"
// is tolerated. An empty list disables the gate (everything passes).
func NewDomainGatedSender(inner Sender, allowedDomains string) Sender {
	domains := splitDomains(allowedDomains)
	if len(domains) == 0 {
		return inner
	}
	g := &DomainGatedSender{
		inner:    inner,
		allowed:  make(map[string]struct{}, len(domains)),
		rendered: strings.Join(domains, ", "),
	}
	for _, d := range domains {
		g.allowed[d] = struct{}{}
	}
	return g
}

func (g *DomainGatedSender) Name() string {
	return g.inner.Name() + "+domain-gate"
}

func (g *DomainGatedSender) Send(ctx context.Context, msg Message) error {
	domain := extractDomain(msg.To)
	if _, ok := g.allowed[domain]; !ok {
		slog.Info("email[gated]: dropped — recipient domain not on allowlist",
			"to", msg.To,
			"recipient_domain", domain,
			"allowed_domains", g.rendered,
			"message_type", msg.MessageType,
		)
		return nil
	}
	return g.inner.Send(ctx, msg)
}

// extractDomain pulls the lowercase domain out of an RFC 5322 address.
// Tolerates both bare "alice@example.com" and "Alice <alice@EXAMPLE.com>".
// Returns "" on parse failure — caller treats that as "not allowed".
func extractDomain(to string) string {
	addr, err := mail.ParseAddress(to)
	if err != nil {
		// Last-ditch: maybe it's a bare address with no display name
		// that ParseAddress choked on; try to find the '@' directly.
		if i := strings.LastIndex(to, "@"); i > 0 && i < len(to)-1 {
			return strings.ToLower(strings.TrimSpace(to[i+1:]))
		}
		return ""
	}
	at := strings.LastIndex(addr.Address, "@")
	if at < 0 {
		return ""
	}
	return strings.ToLower(addr.Address[at+1:])
}

// splitDomains turns "infrays.org, @foo.com  bar.com" into
// ["infrays.org", "foo.com", "bar.com"]. Empty tokens are skipped.
func splitDomains(s string) []string {
	if s == "" {
		return nil
	}
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == ';'
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		d := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(f), "@"))
		if d != "" {
			out = append(out, d)
		}
	}
	return out
}
