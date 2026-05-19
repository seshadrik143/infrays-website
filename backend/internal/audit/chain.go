// Package audit is the hash-chained event log for the issuer service.
//
// Every state-changing operation (enrollment consumed, license issued,
// license refreshed, license revoked, deployment flagged, admin
// action, ...) appends an entry. Each entry's Hash is SHA-256 over
// (Seq || PrevHash || canonical(Payload)), forming a tamper-evident
// chain. Mid-chain deletion or payload tampering surfaces as a hash
// mismatch on Verify.
//
// Phase 49 ships the in-memory implementation. PostgreSQL adapter
// follows the same pattern as NodePulse's server/auditlog package:
// pg_advisory_xact_lock per-issuer-instance serializes the SELECT-
// tail + INSERT-new transaction.
package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Entry is one row in the audit log.
type Entry struct {
	Seq            uint64    // monotonic; 1-based
	PrevHash       string    // hex SHA-256 of previous entry's Hash; "" for the first
	Hash           string    // hex SHA-256 of (Seq || PrevHash || canonical(Payload))
	EventType      string    // "enrollment.consumed" | "license.issued" | "license.refreshed" | "license.revoked" | ...
	CustomerID     string
	SubscriptionID string
	LicenseID      string
	DeploymentID   string
	Actor          string // "system" | admin email | customer email
	Payload        map[string]any
	CreatedAt      time.Time
}

// Log is the storage abstraction for the chain.
type Log interface {
	// Append writes a new entry, computing Seq/PrevHash/Hash. Returns
	// the persisted Entry with those fields filled.
	Append(ctx context.Context, e Entry) (*Entry, error)

	// Tail returns the most recent N entries in chronological order.
	Tail(ctx context.Context, n int) ([]*Entry, error)

	// Verify walks the chain and returns the first mismatch (Seq + reason),
	// or (0, "") if the chain is intact.
	Verify(ctx context.Context) (uint64, string, error)

	// Count returns the total entries logged.
	Count(ctx context.Context) int

	Close() error
}

// Memory is an in-memory Log. Goroutine-safe.
type Memory struct {
	mu      sync.Mutex
	entries []*Entry
}

func NewMemory() *Memory { return &Memory{} }

func (m *Memory) Append(_ context.Context, e Entry) (*Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e.Seq = uint64(len(m.entries) + 1)
	if e.Seq == 1 {
		e.PrevHash = ""
	} else {
		e.PrevHash = m.entries[len(m.entries)-1].Hash
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	h, err := chainHash(e)
	if err != nil {
		return nil, err
	}
	e.Hash = h
	cp := e
	m.entries = append(m.entries, &cp)
	return &cp, nil
}

func (m *Memory) Tail(_ context.Context, n int) ([]*Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if n <= 0 || n > len(m.entries) {
		n = len(m.entries)
	}
	start := len(m.entries) - n
	out := make([]*Entry, n)
	for i := 0; i < n; i++ {
		cp := *m.entries[start+i]
		out[i] = &cp
	}
	return out, nil
}

func (m *Memory) Verify(_ context.Context) (uint64, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var prev string
	for i, e := range m.entries {
		if e.Seq != uint64(i+1) {
			return e.Seq, fmt.Sprintf("seq gap: expected %d, got %d", i+1, e.Seq), nil
		}
		if e.PrevHash != prev {
			return e.Seq, "prev_hash mismatch", nil
		}
		want, err := chainHash(*e)
		if err != nil {
			return e.Seq, "rehash failed: " + err.Error(), nil
		}
		if want != e.Hash {
			return e.Seq, "hash mismatch (payload tampered?)", nil
		}
		prev = e.Hash
	}
	return 0, "", nil
}

func (m *Memory) Count(_ context.Context) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

func (m *Memory) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = nil
	return nil
}

// chainHash computes SHA-256 over Seq || PrevHash || canonical(Payload).
// Payload is canonicalised by sorting map keys; JSON's default ordering
// is not stable for maps.
func chainHash(e Entry) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, "%d|%s|", e.Seq, e.PrevHash)
	canon, err := canonicalJSON(map[string]any{
		"event_type":      e.EventType,
		"customer_id":     e.CustomerID,
		"subscription_id": e.SubscriptionID,
		"license_id":      e.LicenseID,
		"deployment_id":   e.DeploymentID,
		"actor":           e.Actor,
		"created_at":      e.CreatedAt.Format(time.RFC3339Nano),
		"payload":         e.Payload,
	})
	if err != nil {
		return "", err
	}
	h.Write(canon)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// canonicalJSON marshals v with sorted map keys at every level.
// Sufficient for our payloads which are shallow JSON objects.
func canonicalJSON(v any) ([]byte, error) {
	normalized := normalizeForCanonical(v)
	return json.Marshal(normalized)
}

func normalizeForCanonical(v any) any {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make([]any, 0, 2*len(keys))
		for _, k := range keys {
			out = append(out, k)
			out = append(out, normalizeForCanonical(t[k]))
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = normalizeForCanonical(e)
		}
		return out
	default:
		return t
	}
}

// Compile-time interface assertion.
var _ Log = (*Memory)(nil)

// ErrChainBroken is returned by callers that hard-fail on a Verify
// mismatch. The Log interface itself just reports the seq + reason.
var ErrChainBroken = errors.New("audit: chain integrity broken")
