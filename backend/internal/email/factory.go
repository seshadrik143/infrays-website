package email

import (
	"log"
	"os"
)

// NewSenderFromEnv constructs the appropriate Sender based on env vars.
// Called once from cmd/issuer/main.go at startup.
//
//	NP_BREVO_API_KEY           — preferred provider; takes precedence when set
//	NP_POSTMARK_SERVER_TOKEN   — fallback provider when Brevo key is absent
//	NP_EMAIL_FROM              — sender address (defaults to no-reply@infrays.org)
//	NP_EMAIL_ALLOWED_DOMAINS   — optional, comma-separated. When set, only
//	                             recipients on these domains are forwarded;
//	                             everything else is logged and dropped.
//
// When neither key is set, returns a NoopSender and logs a warning.
// The issuer still functions — email triggers fire, the sender just no-ops.
func NewSenderFromEnv() Sender {
	from := os.Getenv("NP_EMAIL_FROM")
	allowed := os.Getenv("NP_EMAIL_ALLOWED_DOMAINS")

	if token := os.Getenv("NP_BREVO_API_KEY"); token != "" {
		base := NewBrevoSender(token, from)
		return wrapGate(base, allowed)
	}
	if token := os.Getenv("NP_POSTMARK_SERVER_TOKEN"); token != "" {
		base := NewPostmarkSender(token, from)
		return wrapGate(base, allowed)
	}
	log.Println("⚠  email: NP_BREVO_API_KEY and NP_POSTMARK_SERVER_TOKEN both unset — emails will NOT be sent (noop sender)")
	return NewNoopSender()
}

func wrapGate(s Sender, allowed string) Sender {
	if allowed == "" {
		log.Printf("email: %s sender configured — no domain gate", s.Name())
		return s
	}
	log.Printf("email: %s sender configured — domain gate active, allowed=%q", s.Name(), allowed)
	log.Printf("email: ⚠  domain gate will be REMOVED when NP_EMAIL_ALLOWED_DOMAINS is unset")
	return NewDomainGatedSender(s, allowed)
}
