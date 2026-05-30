package stripebill

import (
	"sort"
	"testing"
)

func TestParseTierMappingFromEnv_IntervalsAndReverseLookup(t *testing.T) {
	raw := "price_m=professional:professional-v1:month," +
		"price_a=professional:professional-v1:annual," +
		"price_e=enterprise:enterprise-v1" // no interval → defaults to month
	m, err := ParseTierMappingFromEnv(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Forward lookup still works.
	if cfg, ok := m.Lookup("price_m"); !ok || cfg.Tier != "professional" {
		t.Fatalf("forward lookup price_m = %+v ok=%v", cfg, ok)
	}

	// Reverse lookup by tier + interval.
	cases := []struct {
		tier, interval, want string
	}{
		{"professional", "month", "price_m"},
		{"professional", "annual", "price_a"},
		{"professional", "", "price_m"},   // empty defaults to month
		{"enterprise", "month", "price_e"}, // omitted interval defaulted to month
	}
	for _, c := range cases {
		got, ok := m.LookupTier(c.tier, c.interval)
		if !ok || got != c.want {
			t.Errorf("LookupTier(%q,%q) = %q ok=%v, want %q", c.tier, c.interval, got, ok, c.want)
		}
	}

	// Unknown tier / interval must not resolve.
	if _, ok := m.LookupTier("enterprise", "annual"); ok {
		t.Errorf("enterprise annual should not resolve (not configured)")
	}
	if _, ok := m.LookupTier("bogus", "month"); ok {
		t.Errorf("bogus tier should not resolve")
	}
}

func TestListPlans_OmitsFreeAndCarriesOptions(t *testing.T) {
	raw := "price_free=free:free-v1:month," +
		"price_m=professional:professional-v1:month," +
		"price_a=professional:professional-v1:annual"
	m, err := ParseTierMappingFromEnv(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	plans := m.ListPlans()
	if len(plans) != 1 {
		t.Fatalf("want 1 plan (free omitted), got %d: %+v", len(plans), plans)
	}
	p := plans[0]
	if p.Tier != "professional" || p.EntitlementSetID != "professional-v1" {
		t.Fatalf("unexpected plan %+v", p)
	}
	intervals := []string{p.Options[0].Interval, p.Options[1].Interval}
	sort.Strings(intervals)
	if intervals[0] != "annual" || intervals[1] != "month" {
		t.Fatalf("want month+annual options, got %v", intervals)
	}
}

func TestParseTierMappingFromEnv_BadInterval(t *testing.T) {
	if _, err := ParseTierMappingFromEnv("price_x=professional:professional-v1:weekly"); err == nil {
		t.Fatalf("expected error for bad interval")
	}
}

func TestLookupTier_EmptyMappingSafe(t *testing.T) {
	m := NewTierMapping(nil) // dev/test path: no reverse index
	if _, ok := m.LookupTier("professional", "month"); ok {
		t.Errorf("empty mapping must not resolve any tier")
	}
	if plans := m.ListPlans(); len(plans) != 0 {
		t.Errorf("empty mapping ListPlans should be empty, got %v", plans)
	}
}
