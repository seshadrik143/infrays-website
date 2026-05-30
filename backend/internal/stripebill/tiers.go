// Package stripebill is the issuer-side Stripe integration:
//
//   - Webhook ingestion with signature verification + idempotency
//   - Event → store state mapping
//   - Checkout Session creation for self-serve signup
//   - Tier ↔ Price ID ↔ entitlement_set_id mapping
//
// Stripe is the source of truth for subscription state. The issuer
// listens for events, mirrors them into its own DB, and uses that
// state when minting / refusing license refreshes.
//
// Package name avoids collision with `stripe-go` itself (which would
// shadow `stripe` as an identifier).
package stripebill

import (
	"fmt"
	"strings"
)

// TierMapping ties a Stripe Price ID to the local tier + entitlement
// set. Configured at issuer startup from environment variables OR a
// JSON file; lookups are read-only at runtime.
//
//	infraYS Stripe Products / Prices (set up via Stripe Dashboard):
//	  price_1Abc...   → free        (free-v1)
//	  price_1Def...   → professional (professional-v1)
//	  price_1Ghi...   → enterprise   (enterprise-v1)
//
// Both monthly and annual price IDs may map to the same tier — annual
// gets a discount in Stripe but the entitlement is the same. Keep
// multiple Price IDs pointing at one tier in the config.
type TierMapping struct {
	// byPriceID maps Stripe Price IDs to a (tier, entitlement_set_id)
	// pair. Multiple Prices may map to the same tier.
	byPriceID map[string]TierConfig

	// byTier is the reverse index used for self-serve checkout: given a
	// tier + billing interval, resolve back to a Stripe Price ID. Built
	// from the same config as byPriceID (see ParseTierMappingFromEnv).
	// Empty when the mapping was built via NewTierMapping (dev/tests).
	byTier map[string][]PriceOption
}

// TierConfig is what byPriceID resolves to.
type TierConfig struct {
	Tier             string // "free" | "professional" | "enterprise"
	EntitlementSetID string // "professional-v1"
}

// PriceOption is one purchasable Price ID for a tier at a given billing
// interval. A tier typically has one "month" and one "annual" option.
type PriceOption struct {
	PriceID  string // Stripe Price ID
	Interval string // "month" | "annual"
}

// PlanInfo is a tier the portal can present for self-serve purchase,
// with its available billing intervals. Returned by ListPlans.
type PlanInfo struct {
	Tier             string        // "professional" | ...
	EntitlementSetID string        // "professional-v1"
	Options          []PriceOption // available intervals
}

// NewTierMapping returns a TierMapping initialised from a config map.
// Empty map is allowed but means every webhook will fall back to the
// "free" tier — useful for local dev / pre-Stripe-account testing.
func NewTierMapping(byPriceID map[string]TierConfig) *TierMapping {
	cp := make(map[string]TierConfig, len(byPriceID))
	for k, v := range byPriceID {
		cp[k] = v
	}
	return &TierMapping{byPriceID: cp}
}

// Lookup returns the tier + entitlement set for a Price ID, with a
// safe fallback so an unknown price doesn't fail subscription sync.
// Unknown prices get "free" — the issuer logs a warning so the
// operator notices the missing config entry.
func (m *TierMapping) Lookup(priceID string) (TierConfig, bool) {
	if m == nil || len(m.byPriceID) == 0 {
		return TierConfig{Tier: "free", EntitlementSetID: "free-v1"}, false
	}
	cfg, ok := m.byPriceID[priceID]
	if !ok {
		return TierConfig{Tier: "free", EntitlementSetID: "free-v1"}, false
	}
	return cfg, true
}

// LookupTier resolves a (tier, interval) pair back to a Stripe Price ID
// for self-serve checkout. interval is "month" or "annual"; an empty
// interval defaults to "month". Returns ok=false when the tier or
// interval isn't configured — callers must reject the request rather
// than guessing a price.
func (m *TierMapping) LookupTier(tier, interval string) (string, bool) {
	if m == nil || len(m.byTier) == 0 {
		return "", false
	}
	if interval == "" {
		interval = "month"
	}
	for _, opt := range m.byTier[tier] {
		if opt.Interval == interval {
			return opt.PriceID, true
		}
	}
	return "", false
}

// ListPlans returns the configured tiers and their purchasable price
// options, for the portal's plan picker. The "free" tier is omitted —
// it isn't something a customer checks out into. Order is not
// guaranteed; callers that need a stable display order should sort.
func (m *TierMapping) ListPlans() []PlanInfo {
	if m == nil {
		return nil
	}
	out := make([]PlanInfo, 0, len(m.byTier))
	for tier, opts := range m.byTier {
		if tier == "free" {
			continue
		}
		// Entitlement set is the same across intervals for a tier; take
		// it from the first matching byPriceID entry.
		ent := ""
		for _, opt := range opts {
			if cfg, ok := m.byPriceID[opt.PriceID]; ok {
				ent = cfg.EntitlementSetID
				break
			}
		}
		cp := make([]PriceOption, len(opts))
		copy(cp, opts)
		out = append(out, PlanInfo{Tier: tier, EntitlementSetID: ent, Options: cp})
	}
	return out
}

// ParseTierMappingFromEnv parses NP_STRIPE_PRICE_MAP, format:
//
//	"price_xxx=professional:professional-v1,price_yyy=enterprise:enterprise-v1"
//
// An optional third colon segment sets the billing interval used by the
// self-serve checkout reverse index ("month" or "annual"):
//
//	"price_xxx=professional:professional-v1:month,price_ann=professional:professional-v1:annual"
//
// When the interval is omitted it defaults to "month", so older configs
// keep working unchanged. Whitespace tolerant. Returns an empty mapping
// (no error) on empty input — caller can fall through to defaults.
func ParseTierMappingFromEnv(raw string) (*TierMapping, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return NewTierMapping(nil), nil
	}
	out := map[string]TierConfig{}
	byTier := map[string][]PriceOption{}
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		kv := strings.SplitN(entry, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("stripebill: bad price-map entry %q (want price_xxx=tier:entitlement[:interval])", entry)
		}
		priceID := strings.TrimSpace(kv[0])
		// Split into up to 3 parts: tier : entitlement : interval.
		tierParts := strings.SplitN(strings.TrimSpace(kv[1]), ":", 3)
		if len(tierParts) < 2 {
			return nil, fmt.Errorf("stripebill: bad tier spec %q (want tier:entitlement[:interval])", kv[1])
		}
		tier := strings.TrimSpace(tierParts[0])
		interval := "month"
		if len(tierParts) == 3 {
			interval = strings.TrimSpace(tierParts[2])
			if interval != "month" && interval != "annual" {
				return nil, fmt.Errorf("stripebill: bad interval %q (want month or annual)", interval)
			}
		}
		out[priceID] = TierConfig{
			Tier:             tier,
			EntitlementSetID: strings.TrimSpace(tierParts[1]),
		}
		byTier[tier] = append(byTier[tier], PriceOption{PriceID: priceID, Interval: interval})
	}
	m := NewTierMapping(out)
	m.byTier = byTier
	return m, nil
}
