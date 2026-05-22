package obs

import (
	"net/http"
	"strings"
)

// TrailingSlash normalizes API request paths by stripping a single
// trailing slash before dispatch. Without this, Go's stdlib
// http.ServeMux treats "POST /api/portal/auth/login" and
// "POST /api/portal/auth/login/" as different routes — the former
// hits the handler, the latter 404s.
//
// Scoping: only normalize API paths (/api/* and /v1/*). The SPA
// fallback may want to serve specific slash variants (it doesn't
// today, but that's its decision), so we leave non-API paths
// untouched.
//
// "/" itself is never normalized (it has no trailing slash to strip).
func TrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if len(p) > 1 && strings.HasSuffix(p, "/") &&
			(strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/v1/") || strings.HasPrefix(p, "/internal/")) {
			r2 := r.Clone(r.Context())
			r2.URL.Path = strings.TrimRight(p, "/")
			next.ServeHTTP(w, r2)
			return
		}
		next.ServeHTTP(w, r)
	})
}
