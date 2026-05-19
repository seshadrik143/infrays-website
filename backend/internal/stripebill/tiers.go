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
}

// TierConfig is what byPriceID resolves to.
type TierConfig struct {
	Tier             string // "free" | "professional" | "enterprise"
	EntitlementSetID string // "professional-v1"
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

// ParseTierMappingFromEnv parses NP_STRIPE_PRICE_MAP, format:
//
//	"price_xxx=professional:professional-v1,price_yyy=enterprise:enterprise-v1"
//
// Whitespace tolerant. Returns an empty mapping (no error) on empty
// input — caller can fall through to defaults.
func ParseTierMappingFromEnv(raw string) (*TierMapping, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return NewTierMapping(nil), nil
	}
	out := map[string]TierConfig{}
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		kv := strings.SplitN(entry, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("stripebill: bad price-map entry %q (want price_xxx=tier:entitlement)", entry)
		}
		priceID := strings.TrimSpace(kv[0])
		tierParts := strings.SplitN(strings.TrimSpace(kv[1]), ":", 2)
		if len(tierParts) != 2 {
			return nil, fmt.Errorf("stripebill: bad tier spec %q (want tier:entitlement)", kv[1])
		}
		out[priceID] = TierConfig{
			Tier:             strings.TrimSpace(tierParts[0]),
			EntitlementSetID: strings.TrimSpace(tierParts[1]),
		}
	}
	return NewTierMapping(out), nil
}
