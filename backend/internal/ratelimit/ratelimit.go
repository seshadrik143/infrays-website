// Package ratelimit implements a simple in-memory sliding-window rate
// limiter for HTTP auth endpoints.
//
// Usage:
//
//	rl := ratelimit.New(ratelimit.Config{
//	    Max:    5,
//	    Window: 15 * time.Minute,
//	})
//	if ok, retry := rl.Allow("login:" + ip); !ok {
//	    w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
//	    http.Error(w, "rate limited", http.StatusTooManyRequests)
//	    return
//	}
//
// Use distinct key prefixes per endpoint so e.g. a flood of failed
// logins doesn't lock out password-reset for the same IP. The cleanup
// goroutine evicts stale keys every Window.
//
// Not distributed — each Fly machine has its own counters. For
// per-machine rate budgets that's fine (an attacker pinned to one
// region still hits one machine's counters; Fly's region-pinning
// means cross-machine spread isn't a free amplification).
package ratelimit

import (
	"sync"
	"time"
)

// Config controls a single Limiter.
//
//	Max: max events allowed in the rolling window
//	Window: rolling window length
//
// The limiter is "fixed-tick sliding window" — each key holds a slice
// of event timestamps; on Allow we drop entries older than Window and
// then check len < Max. Memory cost per key: 8 bytes × Max. With Max=20
// and 10k tracked keys that's ~1.6 MB.
type Config struct {
	Max    int
	Window time.Duration
	// CleanupInterval controls how often the eviction sweep runs.
	// Default Window when zero.
	CleanupInterval time.Duration
	// Now is injectable for tests; defaults to time.Now.
	Now func() time.Time
}

// Limiter is a goroutine-safe sliding-window rate limiter.
type Limiter struct {
	cfg     Config
	mu      sync.Mutex
	buckets map[string][]time.Time
	stop    chan struct{}
}

// New constructs a Limiter and starts its cleanup goroutine.
// Stop the limiter by calling Close — important in tests so
// the goroutine doesn't leak.
func New(cfg Config) *Limiter {
	if cfg.Max <= 0 {
		cfg.Max = 5
	}
	if cfg.Window <= 0 {
		cfg.Window = 15 * time.Minute
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = cfg.Window
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	rl := &Limiter{
		cfg:     cfg,
		buckets: map[string][]time.Time{},
		stop:    make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// Allow records an event for key and reports whether it's allowed.
// On rejection, retryAfter is how long until the oldest in-window
// event ages out (i.e. the wait before another attempt could succeed).
func (rl *Limiter) Allow(key string) (allowed bool, retryAfter time.Duration) {
	now := rl.cfg.Now()
	cutoff := now.Add(-rl.cfg.Window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	events := rl.buckets[key]
	// Drop expired entries from the head (slice is append-ordered so
	// older events are first).
	i := 0
	for i < len(events) && events[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		events = events[i:]
	}

	if len(events) >= rl.cfg.Max {
		// Earliest event ages out at events[0] + Window.
		retry := events[0].Add(rl.cfg.Window).Sub(now)
		if retry < time.Second {
			retry = time.Second
		}
		// Don't record this event — keeps the limiter from extending
		// the window every time the attacker probes.
		rl.buckets[key] = events
		return false, retry
	}

	events = append(events, now)
	rl.buckets[key] = events
	return true, 0
}

// Reset clears all recorded events for a key. Called after a
// successful login so the counter doesn't persist against a legitimate
// user across one bad-then-good sequence.
func (rl *Limiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.buckets, key)
}

// Close stops the cleanup goroutine. Idempotent.
func (rl *Limiter) Close() {
	select {
	case <-rl.stop:
		// Already closed.
	default:
		close(rl.stop)
	}
}

func (rl *Limiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cfg.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stop:
			return
		case <-ticker.C:
			rl.evictStale()
		}
	}
}

func (rl *Limiter) evictStale() {
	now := rl.cfg.Now()
	cutoff := now.Add(-rl.cfg.Window)
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for k, events := range rl.buckets {
		i := 0
		for i < len(events) && events[i].Before(cutoff) {
			i++
		}
		switch {
		case i == len(events):
			delete(rl.buckets, k)
		case i > 0:
			rl.buckets[k] = events[i:]
		}
	}
}
