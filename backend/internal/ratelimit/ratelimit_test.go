package ratelimit_test

import (
	"sync"
	"testing"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/ratelimit"
)

func TestAllowWithinBurst(t *testing.T) {
	now := time.Now()
	rl := ratelimit.New(ratelimit.Config{
		Max: 3, Window: time.Minute,
		Now: func() time.Time { return now },
	})
	defer rl.Close()

	for i := 0; i < 3; i++ {
		ok, _ := rl.Allow("k")
		if !ok {
			t.Fatalf("attempt %d: expected allowed", i+1)
		}
	}
	ok, retry := rl.Allow("k")
	if ok {
		t.Fatal("attempt 4: expected blocked")
	}
	if retry <= 0 {
		t.Fatalf("expected positive retry, got %v", retry)
	}
}

func TestAllowAfterWindow(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	rl := ratelimit.New(ratelimit.Config{
		Max: 2, Window: time.Minute,
		Now: clock,
	})
	defer rl.Close()

	_, _ = rl.Allow("k")
	_, _ = rl.Allow("k")
	if ok, _ := rl.Allow("k"); ok {
		t.Fatal("expected blocked on 3rd")
	}

	now = now.Add(61 * time.Second) // past window
	if ok, _ := rl.Allow("k"); !ok {
		t.Fatal("expected allowed after window")
	}
}

func TestKeysIndependent(t *testing.T) {
	now := time.Now()
	rl := ratelimit.New(ratelimit.Config{
		Max: 1, Window: time.Minute,
		Now: func() time.Time { return now },
	})
	defer rl.Close()
	if ok, _ := rl.Allow("a"); !ok {
		t.Fatal("a #1 should allow")
	}
	if ok, _ := rl.Allow("a"); ok {
		t.Fatal("a #2 should block")
	}
	if ok, _ := rl.Allow("b"); !ok {
		t.Fatal("b #1 should allow (independent key)")
	}
}

func TestReset(t *testing.T) {
	rl := ratelimit.New(ratelimit.Config{Max: 1, Window: time.Minute})
	defer rl.Close()
	_, _ = rl.Allow("k")
	if ok, _ := rl.Allow("k"); ok {
		t.Fatal("expected blocked after burst")
	}
	rl.Reset("k")
	if ok, _ := rl.Allow("k"); !ok {
		t.Fatal("expected allowed after Reset")
	}
}

func TestRetryAfterMonotonic(t *testing.T) {
	base := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	now := base
	rl := ratelimit.New(ratelimit.Config{
		Max: 1, Window: time.Minute,
		Now: func() time.Time { return now },
	})
	defer rl.Close()

	_, _ = rl.Allow("k") // burns the one allowed slot at base
	now = base.Add(20 * time.Second)
	_, retry := rl.Allow("k")
	// First event was at base; ages out at base+60s. now=base+20s.
	// Expected retry ≈ 40s.
	if retry < 30*time.Second || retry > 45*time.Second {
		t.Fatalf("retry out of expected range: %v", retry)
	}
}

func TestConcurrentSafe(t *testing.T) {
	rl := ratelimit.New(ratelimit.Config{Max: 1000, Window: time.Minute})
	defer rl.Close()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_, _ = rl.Allow("shared")
			}
		}(i)
	}
	wg.Wait()
	// No assertion on exact state — race detector enforces correctness.
}

func TestCloseIdempotent(t *testing.T) {
	rl := ratelimit.New(ratelimit.Config{Max: 1, Window: time.Minute})
	rl.Close()
	rl.Close() // must not panic / deadlock
}

func TestDefaultsApplied(t *testing.T) {
	// Zero config still gives a usable limiter.
	rl := ratelimit.New(ratelimit.Config{})
	defer rl.Close()
	// Default Max is 5 — sixth call should block.
	for i := 0; i < 5; i++ {
		if ok, _ := rl.Allow("default"); !ok {
			t.Fatalf("expected allow within default burst, blocked at %d", i+1)
		}
	}
	if ok, _ := rl.Allow("default"); ok {
		t.Fatal("expected default Max=5 to block 6th attempt")
	}
}
