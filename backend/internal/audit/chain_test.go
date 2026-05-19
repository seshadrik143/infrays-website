package audit

import (
	"context"
	"testing"
)

func TestMemoryAppendAndVerify(t *testing.T) {
	ctx := context.Background()
	log := NewMemory()

	for i := 0; i < 5; i++ {
		_, err := log.Append(ctx, Entry{
			EventType: "test.event",
			Actor:     "system",
			Payload:   map[string]any{"i": i},
		})
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	seq, reason, err := log.Verify(ctx)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if seq != 0 {
		t.Fatalf("chain broken at seq %d: %s", seq, reason)
	}
	if got := log.Count(ctx); got != 5 {
		t.Errorf("count: got %d, want 5", got)
	}
}

func TestMemoryDetectsTampering(t *testing.T) {
	ctx := context.Background()
	log := NewMemory()
	for i := 0; i < 3; i++ {
		_, _ = log.Append(ctx, Entry{EventType: "t", Payload: map[string]any{"i": i}})
	}
	// Tamper with entry 2's payload
	log.entries[1].Payload["i"] = 999
	seq, _, err := log.Verify(ctx)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if seq != 2 {
		t.Errorf("expected break at seq 2, got %d", seq)
	}
}

func TestMemoryTailOrdering(t *testing.T) {
	ctx := context.Background()
	log := NewMemory()
	for i := 0; i < 10; i++ {
		_, _ = log.Append(ctx, Entry{EventType: "t", Payload: map[string]any{"i": i}})
	}
	tail, err := log.Tail(ctx, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(tail) != 3 {
		t.Fatalf("tail len: got %d", len(tail))
	}
	if tail[0].Seq != 8 || tail[2].Seq != 10 {
		t.Errorf("tail order: got seqs %d/%d, want 8/10", tail[0].Seq, tail[2].Seq)
	}
}
