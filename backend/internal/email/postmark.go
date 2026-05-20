package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PostmarkSender posts to https://api.postmarkapp.com/email. Stdlib
// HTTP only — no SDK dep. Postmark's API is simple enough that
// adding their Go SDK would be more weight than it's worth.
//
// Reference: https://postmarkapp.com/developer/api/email-api
type PostmarkSender struct {
	serverToken string
	from        string // "Name <email@infrays.org>" or bare "email@infrays.org"
	stream      string // "outbound" (default) | "broadcast" (for marketing emails — we don't use)
	client      *http.Client
}

const postmarkURL = "https://api.postmarkapp.com/email"

// NewPostmarkSender returns a ready Sender. fromAddr must be a
// verified sender signature in your Postmark account — Postmark
// rejects sends from unverified domains. Empty fromAddr defaults to
// "no-reply@infrays.org" which the operator must verify.
func NewPostmarkSender(serverToken, fromAddr string) *PostmarkSender {
	if fromAddr == "" {
		fromAddr = "infraYS <no-reply@infrays.org>"
	}
	return &PostmarkSender{
		serverToken: serverToken,
		from:        fromAddr,
		stream:      "outbound",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *PostmarkSender) Name() string { return "postmark" }

func (p *PostmarkSender) Send(ctx context.Context, msg Message) error {
	if msg.To == "" {
		return errors.New("email: empty To")
	}
	if msg.Subject == "" {
		return errors.New("email: empty Subject")
	}
	body := map[string]any{
		"From":          p.from,
		"To":            msg.To,
		"Subject":       msg.Subject,
		"HtmlBody":      msg.HTMLBody,
		"TextBody":      msg.TextBody,
		"MessageStream": p.stream,
	}
	if msg.MessageType != "" {
		body["Tag"] = msg.MessageType // Postmark tags = our message types
	}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, postmarkURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("email: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", p.serverToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("email: postmark %s: %w", msg.MessageType, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("email: postmark status=%d body=%s", resp.StatusCode, string(respBody))
	}
	return nil
}
