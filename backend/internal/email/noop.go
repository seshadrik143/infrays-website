package email

import (
	"context"
	"log"
)

// NoopSender logs the call and drops the message. Used when no
// provider is configured — issuer runs without email during early
// dev / pre-Postmark setup.
//
// Operator sees the would-have-been-sent in logs so they know the
// trigger paths are firing correctly even before Postmark is wired.
type NoopSender struct{}

func NewNoopSender() *NoopSender { return &NoopSender{} }

func (n *NoopSender) Name() string { return "noop" }

func (n *NoopSender) Send(_ context.Context, msg Message) error {
	log.Printf("email[noop]: would send type=%q to=%q subject=%q (NP_POSTMARK_SERVER_TOKEN unset)",
		msg.MessageType, msg.To, msg.Subject)
	return nil
}
