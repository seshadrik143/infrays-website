package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"time"
)

// BrevoSender posts to https://api.brevo.com/v3/smtp/email. Stdlib HTTP
// only — matches the PostmarkSender approach, no extra SDK dependency.
//
// Reference: https://developers.brevo.com/reference/sendtransacemail
type BrevoSender struct {
	apiKey string
	from   brevoAddr
	client *http.Client
}

type brevoAddr struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

const brevoURL = "https://api.brevo.com/v3/smtp/email"

// NewBrevoSender returns a ready Sender. fromAddr must be a verified
// sender in your Brevo account. Empty fromAddr defaults to
// "infraYS <no-reply@infrays.org>".
func NewBrevoSender(apiKey, fromAddr string) *BrevoSender {
	if fromAddr == "" {
		fromAddr = "infraYS <no-reply@infrays.org>"
	}
	return &BrevoSender{
		apiKey: apiKey,
		from:   parseBrevoAddr(fromAddr),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (b *BrevoSender) Name() string { return "brevo" }

func (b *BrevoSender) Send(ctx context.Context, msg Message) error {
	if msg.To == "" {
		return errors.New("email: empty To")
	}
	if msg.Subject == "" {
		return errors.New("email: empty Subject")
	}

	type payload struct {
		Sender      brevoAddr   `json:"sender"`
		To          []brevoAddr `json:"to"`
		Subject     string      `json:"subject"`
		HTMLContent string      `json:"htmlContent,omitempty"`
		TextContent string      `json:"textContent,omitempty"`
		Tags        []string    `json:"tags,omitempty"`
	}
	p := payload{
		Sender:      b.from,
		To:          []brevoAddr{parseBrevoAddr(msg.To)},
		Subject:     msg.Subject,
		HTMLContent: msg.HTMLBody,
		TextContent: msg.TextBody,
	}
	if msg.MessageType != "" {
		p.Tags = []string{msg.MessageType}
	}

	rawBody, _ := json.Marshal(p)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, brevoURL, bytes.NewReader(rawBody))
	if err != nil {
		return fmt.Errorf("email: build request: %w", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("api-key", b.apiKey)

	resp, err := b.client.Do(req)
	if err != nil {
		slog.Error("brevo send transport failure",
			"message_type", msg.MessageType, "to", msg.To, "err", err)
		return fmt.Errorf("email: brevo %s: %w", msg.MessageType, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error("brevo send rejected",
			"message_type", msg.MessageType, "to", msg.To,
			"status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("email: brevo status=%d body=%s", resp.StatusCode, string(respBody))
	}
	slog.Info("brevo send ok",
		"message_type", msg.MessageType, "to", msg.To, "status", resp.StatusCode)
	return nil
}

// parseBrevoAddr splits "Name <email>" or bare "email" into a brevoAddr.
func parseBrevoAddr(s string) brevoAddr {
	if addr, err := mail.ParseAddress(s); err == nil {
		return brevoAddr{Name: addr.Name, Email: addr.Address}
	}
	return brevoAddr{Email: s}
}
