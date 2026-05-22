package obs

import (
	"encoding/json"
	"net/http"
)

// JSON404 wraps a *http.ServeMux so requests that don't match any
// registered pattern return a JSON 404 body instead of stdlib's
// plain-text "404 page not found\n".
//
// Uses ServeMux.Handler() to detect no-match before dispatching —
// avoids the buffer-and-rewrite dance that would require wrapping
// every ResponseWriter. Method mismatches (405) are passed through
// to the stdlib handler which serves a sensible plain-text body
// plus an Allow header — that response stays as-is.
//
// Mount example:
//
//	root := http.NewServeMux()
//	root.Handle("/api/portal/", obs.JSON404(portalMux))
func JSON404(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			writeJSONStatus(w, http.StatusNotFound, "not found")
			return
		}
		mux.ServeHTTP(w, r)
	})
}

// JSONStatus writes a {"error": msg} JSON response with the given
// HTTP status. Exported so endpoint-specific 404 handlers (e.g. the
// /metrics one) can use the same shape.
func JSONStatus(w http.ResponseWriter, status int, msg string) {
	writeJSONStatus(w, status, msg)
}

func writeJSONStatus(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
