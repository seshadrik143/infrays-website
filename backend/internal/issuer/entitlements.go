package issuer

import (
	"net/http"
	"strings"
)

// GET /v1/entitlements/{id}.json
// Returns the feature manifest as a cacheable JSON document. Cache-
// Control allows CDNs to memoize for an hour (manifests are
// immutable once published; new manifests get new IDs).
func (s *Server) handleEntitlements(w http.ResponseWriter, r *http.Request) {
	// Extract id from path. Go 1.22 mux pattern would have given us
	// PathValue("id"), but our route is /v1/entitlements/ (catch-all)
	// so we slice manually.
	id := strings.TrimPrefix(r.URL.Path, "/v1/entitlements/")
	id = strings.TrimSuffix(id, ".json")
	if id == "" {
		writeErr(w, http.StatusNotFound, "not_found", "missing entitlement set id")
		return
	}

	es, err := s.cfg.Store.GetEntitlementSet(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "entitlement set not found")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=3600, immutable")
	writeJSON(w, http.StatusOK, map[string]any{
		"id":       es.ID,
		"name":     es.Name,
		"version":  es.Version,
		"features": es.Features,
		"limits": map[string]any{
			"max_agents":          es.Limits.MaxAgents,
			"max_metrics_per_sec": es.Limits.MaxMetricsPerSec,
			"max_log_gb_per_day":  es.Limits.MaxLogGBPerDay,
			"max_alert_rules":     es.Limits.MaxAlertRules,
			"retention_days":      es.Limits.RetentionDays,
		},
	})
}
