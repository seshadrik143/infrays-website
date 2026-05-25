package email

import (
	"log"
	"os"
)

// NewSenderFromEnv constructs the appropriate Sender based on env
// vars. Called once from cmd/issuer/main.go at startup.
//
//	NP_POSTMARK_SERVER_TOKEN   — required to enable Postmark
//	NP_EMAIL_FROM              — sender address (defaults to no-reply@infrays.org)
//	NP_EMAIL_ALLOWED_DOMAINS   — optional, comma-separated. When set,
//	                             only recipients on these domains are
//	                             forwarded to the upstream sender;
//	                             everything else is logged and dropped.
//	                             Used during the Postmark "pending
//	                             approval" period — once approved, unset
//	                             this env var to allow all recipients.
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
	base := NewPostmarkSender(token, from)

	allowed := os.Getenv("NP_EMAIL_ALLOWED_DOMAINS")
	if allowed != "" {
		log.Printf("email: postmark sender configured (from=%q) — domain gate active, allowed=%q",
			from, allowed)
		log.Printf("email: ⚠  domain gate will be REMOVED when NP_EMAIL_ALLOWED_DOMAINS is unset (after Postmark approval)")
		return NewDomainGatedSender(base, allowed)
	}
	log.Printf("email: postmark sender configured (from=%q) — no domain gate", from)
	return base
}
