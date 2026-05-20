package email

import (
	"context"
	"sync"
)

// CaptureSender stores every Send call in-memory. Test-only.
//
// Tests construct it directly and inspect .Messages() after the
// code under test has run, to assert that the right emails were
// triggered with the right recipients / subjects.
type CaptureSender struct {
	mu   sync.Mutex
	sent []Message
}

func NewCaptureSender() *CaptureSender { return &CaptureSender{} }

func (c *CaptureSender) Name() string { return "capture" }

func (c *CaptureSender) Send(_ context.Context, msg Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, msg)
	return nil
}

// Messages returns a snapshot of everything sent so far. The slice
// is a copy — modifying it doesn't affect the sender's internal
// state.
func (c *CaptureSender) Messages() []Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Message, len(c.sent))
	copy(out, c.sent)
	return out
}

// MessagesOfType returns only the messages with a matching
// MessageType. Convenience for tests that want to assert "exactly
// one trial_expiring email went out."
func (c *CaptureSender) MessagesOfType(msgType string) []Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []Message
	for _, m := range c.sent {
		if m.MessageType == msgType {
			out = append(out, m)
		}
	}
	return out
}

// Reset drops all captured messages. Useful between sub-tests.
func (c *CaptureSender) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = nil
}
