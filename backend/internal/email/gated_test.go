package email

import (
	"context"
	"errors"
	"testing"
)

// recordingSender is a test Sender that remembers what it was given.
type recordingSender struct {
	calls []Message
	err   error
}

func (r *recordingSender) Name() string { return "recording" }
func (r *recordingSender) Send(_ context.Context, m Message) error {
	r.calls = append(r.calls, m)
	return r.err
}

func TestDomainGatedSender_EmptyAllowlist_PassesThrough(t *testing.T) {
	rec := &recordingSender{}
	s := NewDomainGatedSender(rec, "")
	if s != Sender(rec) {
		t.Fatalf("empty allowlist should return the wrapped sender unchanged")
	}
}

func TestDomainGatedSender_AllowsListedDomain(t *testing.T) {
	rec := &recordingSender{}
	s := NewDomainGatedSender(rec, "infrays.org")
	if err := s.Send(context.Background(), Message{To: "ops@infrays.org", Subject: "x"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(rec.calls) != 1 {
		t.Errorf("expected 1 inner send, got %d", len(rec.calls))
	}
}

func TestDomainGatedSender_DropsUnlistedDomain(t *testing.T) {
	rec := &recordingSender{}
	s := NewDomainGatedSender(rec, "infrays.org")
	if err := s.Send(context.Background(), Message{To: "alice@gmail.com", Subject: "x"}); err != nil {
		t.Fatalf("send should not return error for dropped recipient: %v", err)
	}
	if len(rec.calls) != 0 {
		t.Errorf("expected 0 inner sends, got %d (gate let it through)", len(rec.calls))
	}
}

func TestDomainGatedSender_DropsUnlistedEvenIfInnerWouldFail(t *testing.T) {
	rec := &recordingSender{err: errors.New("inner would have failed")}
	s := NewDomainGatedSender(rec, "infrays.org")
	if err := s.Send(context.Background(), Message{To: "x@gmail.com"}); err != nil {
		t.Errorf("gate should swallow without calling inner, got %v", err)
	}
}

func TestDomainGatedSender_MultipleDomains(t *testing.T) {
	rec := &recordingSender{}
	s := NewDomainGatedSender(rec, "infrays.org, partner.com  trusted.io")
	for _, addr := range []string{"a@infrays.org", "b@partner.com", "c@trusted.io"} {
		if err := s.Send(context.Background(), Message{To: addr, Subject: "x"}); err != nil {
			t.Errorf("send(%s): %v", addr, err)
		}
	}
	if err := s.Send(context.Background(), Message{To: "blocked@evil.com", Subject: "x"}); err != nil {
		t.Errorf("blocked send returned error: %v", err)
	}
	if len(rec.calls) != 3 {
		t.Errorf("expected 3 inner sends, got %d", len(rec.calls))
	}
}

func TestDomainGatedSender_CaseInsensitive(t *testing.T) {
	rec := &recordingSender{}
	s := NewDomainGatedSender(rec, "INFRAYS.ORG")
	if err := s.Send(context.Background(), Message{To: "Alice <alice@Infrays.org>", Subject: "x"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(rec.calls) != 1 {
		t.Errorf("case-insensitive match failed; got %d sends", len(rec.calls))
	}
}

func TestDomainGatedSender_HandlesDisplayName(t *testing.T) {
	rec := &recordingSender{}
	s := NewDomainGatedSender(rec, "infrays.org")
	if err := s.Send(context.Background(), Message{To: "Ops Team <ops@infrays.org>", Subject: "x"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(rec.calls) != 1 {
		t.Errorf("display-name form was rejected; got %d sends", len(rec.calls))
	}
}

func TestDomainGatedSender_LeadingAtTolerated(t *testing.T) {
	rec := &recordingSender{}
	s := NewDomainGatedSender(rec, "@infrays.org")
	if err := s.Send(context.Background(), Message{To: "ops@infrays.org", Subject: "x"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(rec.calls) != 1 {
		t.Errorf("leading-@ allowlist entry was not normalised; got %d sends", len(rec.calls))
	}
}

func TestDomainGatedSender_MalformedAddressDropped(t *testing.T) {
	rec := &recordingSender{}
	s := NewDomainGatedSender(rec, "infrays.org")
	if err := s.Send(context.Background(), Message{To: "not-an-email", Subject: "x"}); err != nil {
		t.Errorf("malformed send returned error: %v", err)
	}
	if len(rec.calls) != 0 {
		t.Errorf("malformed address was forwarded; got %d sends", len(rec.calls))
	}
}

func TestExtractDomain(t *testing.T) {
	cases := []struct{ in, want string }{
		{"alice@example.com", "example.com"},
		{"Alice <alice@example.com>", "example.com"},
		{"alice@EXAMPLE.com", "example.com"},
		{"  alice@example.com  ", "example.com"},
		{"not-an-email", ""},
		{"", ""},
		{"@nodomain", ""}, // bare @ with no local-part is invalid per RFC 5322
	}
	for _, c := range cases {
		if got := extractDomain(c.in); got != c.want {
			t.Errorf("extractDomain(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNewSenderFromEnv_DomainGateActive(t *testing.T) {
	t.Setenv("NP_BREVO_API_KEY", "")
	t.Setenv("NP_POSTMARK_SERVER_TOKEN", "fake-token-xyz")
	t.Setenv("NP_EMAIL_ALLOWED_DOMAINS", "infrays.org")
	s := NewSenderFromEnv()
	if s.Name() != "postmark+domain-gate" {
		t.Errorf("expected gated postmark, got %s", s.Name())
	}
}

func TestNewSenderFromEnv_BrevoGateActive(t *testing.T) {
	t.Setenv("NP_BREVO_API_KEY", "fake-brevo-key")
	t.Setenv("NP_POSTMARK_SERVER_TOKEN", "")
	t.Setenv("NP_EMAIL_ALLOWED_DOMAINS", "infrays.org")
	s := NewSenderFromEnv()
	if s.Name() != "brevo+domain-gate" {
		t.Errorf("expected gated brevo, got %s", s.Name())
	}
}
