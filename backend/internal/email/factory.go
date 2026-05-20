package email

import (
	"log"
	"os"
)

// NewSenderFromEnv constructs the appropriate Sender based on env
// vars. Called once from cmd/issuer/main.go at startup.
//
//	NP_POSTMARK_SERVER_TOKEN  — required to enable Postmark
//	NP_EMAIL_FROM             — sender address (defaults to no-reply@infrays.org)
//
// When NP_POSTMARK_SERVER_TOKEN is empty, returns a NoopSender and
// logs a warning at startup. Issuer still functions — email
// triggers run, the sender just no-ops.
func NewSenderFromEnv() Sender {
	token := os.Getenv("NP_POSTMARK_SERVER_TOKEN")
	if token == "" {
		log.Println("⚠  email: NP_POSTMARK_SERVER_TOKEN unset — emails will NOT be sent (noop sender)")
		return NewNoopSender()
	}
	from := os.Getenv("NP_EMAIL_FROM")
	log.Printf("email: postmark sender configured (from=%q)", from)
	return NewPostmarkSender(token, from)
}
