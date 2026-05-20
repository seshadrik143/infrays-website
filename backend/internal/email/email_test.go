package email

import (
	"context"
	"strings"
	"testing"
)

func TestCaptureSender_StoresMessages(t *testing.T) {
	c := NewCaptureSender()
	if c.Name() != "capture" {
		t.Errorf("name: got %q", c.Name())
	}
	for i := 0; i < 3; i++ {
		_ = c.Send(context.Background(), Message{To: "a@b", Subject: "test", MessageType: "trial_expiring"})
	}
	if len(c.Messages()) != 3 {
		t.Errorf("count: got %d, want 3", len(c.Messages()))
	}
	if got := c.MessagesOfType("trial_expiring"); len(got) != 3 {
		t.Errorf("by type: got %d, want 3", len(got))
	}
	if got := c.MessagesOfType("welcome"); len(got) != 0 {
		t.Errorf("wrong type: got %d, want 0", len(got))
	}
	c.Reset()
	if len(c.Messages()) != 0 {
		t.Error("Reset should clear")
	}
}

func TestNoopSender_AlwaysSucceeds(t *testing.T) {
	n := NewNoopSender()
	if n.Name() != "noop" {
		t.Errorf("name: got %q", n.Name())
	}
	if err := n.Send(context.Background(), Message{To: "x@y", Subject: "s"}); err != nil {
		t.Errorf("noop send failed: %v", err)
	}
}

func TestRenderWelcome(t *testing.T) {
	msg, err := RenderWelcome("alice@example.com", WelcomeData{
		CustomerName: "Alice",
		AppURL:       "https://app.test",
		DocsURL:      "https://app.test/docs",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if msg.MessageType != "welcome" {
		t.Errorf("type: got %q", msg.MessageType)
	}
	if msg.To != "alice@example.com" {
		t.Errorf("to: got %q", msg.To)
	}
	if !strings.Contains(msg.Subject, "Welcome") {
		t.Errorf("subject: got %q", msg.Subject)
	}
	if !strings.Contains(msg.HTMLBody, "Alice") {
		t.Errorf("html missing customer name")
	}
	if !strings.Contains(msg.HTMLBody, "https://app.test/deployments") {
		t.Errorf("html missing AppURL substitution")
	}
	if !strings.Contains(msg.TextBody, "Alice") {
		t.Errorf("text missing customer name")
	}
}

func TestRenderTrialExpiring_SubjectPluralization(t *testing.T) {
	cases := []struct {
		days     int
		wantSubj string
	}{
		{1, "expires in 1 day"},
		{3, "expires in 3 days"},
		{7, "expires in 7 days"},
	}
	for _, tc := range cases {
		msg, err := RenderTrialExpiring("x@y", TrialExpiringData{
			CustomerName: "Bob",
			DaysLeft:     tc.days,
			AppURL:       "https://app.test",
			UpgradeURL:   "https://app.test/billing",
		})
		if err != nil {
			t.Fatalf("days=%d: %v", tc.days, err)
		}
		if !strings.Contains(msg.Subject, tc.wantSubj) {
			t.Errorf("days=%d: subject %q, want substring %q", tc.days, msg.Subject, tc.wantSubj)
		}
	}
}

func TestRenderPaymentFailed(t *testing.T) {
	msg, err := RenderPaymentFailed("charlie@test.com", PaymentFailedData{
		CustomerName:  "Charlie",
		AmountDueUSD:  "$49.00",
		AppURL:        "https://app.test",
		UpdateCardURL: "https://app.test/billing",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg.HTMLBody, "$49.00") {
		t.Errorf("html missing amount")
	}
	if !strings.Contains(msg.HTMLBody, "https://app.test/billing") {
		t.Errorf("html missing update card URL")
	}
}

func TestRenderSubscriptionCanceled(t *testing.T) {
	msg, err := RenderSubscriptionCanceled("dee@test", SubscriptionCanceledData{
		CustomerName: "Dee",
		AccessEndsOn: "January 1, 2027",
		AppURL:       "https://app.test",
		ContactURL:   "https://app.test/contact",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg.HTMLBody, "January 1, 2027") {
		t.Errorf("html missing access-ends date")
	}
	if !strings.Contains(msg.TextBody, "January 1, 2027") {
		t.Errorf("text missing access-ends date")
	}
}

func TestRenderEnrollmentToken(t *testing.T) {
	msg, err := RenderEnrollmentTokenCreated("ops@acme", EnrollmentTokenCreatedData{
		CustomerName:    "ACME Ops",
		EnrollmentToken: "NP-ENROLL-ABCDEFGH",
		Label:           "prod-cluster-east",
		ExpiresIn:       "24 hours",
		DocsURL:         "https://app.test/docs/enroll",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg.HTMLBody, "NP-ENROLL-ABCDEFGH") {
		t.Errorf("html missing token")
	}
	if !strings.Contains(msg.HTMLBody, "prod-cluster-east") {
		t.Errorf("html missing label")
	}
	if !strings.Contains(msg.HTMLBody, "24 hours") {
		t.Errorf("html missing TTL")
	}
}

func TestNewSenderFromEnv_NoopWithoutToken(t *testing.T) {
	t.Setenv("NP_POSTMARK_SERVER_TOKEN", "")
	s := NewSenderFromEnv()
	if s.Name() != "noop" {
		t.Errorf("expected noop, got %s", s.Name())
	}
}

func TestNewSenderFromEnv_PostmarkWithToken(t *testing.T) {
	t.Setenv("NP_POSTMARK_SERVER_TOKEN", "fake-token-xyz")
	s := NewSenderFromEnv()
	if s.Name() != "postmark" {
		t.Errorf("expected postmark, got %s", s.Name())
	}
}
