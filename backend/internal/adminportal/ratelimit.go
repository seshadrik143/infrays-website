package adminportal

import (
	"net/http"
	"strconv"
	"strings"
)

// ratelimitOrAbort enforces the limiter and writes 429 if blocked.
// Returns true when the request should continue.
func (s *Server) ratelimitOrAbort(w http.ResponseWriter, rl *Limiter, key string) bool {
	ok, retry := rl.Allow(key)
	if ok {
		return true
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
	writeError(w, http.StatusTooManyRequests, "too many requests — try again later")
	return false
}

// rateLimitClientIP — same trim as session.go's clientIP, kept local.
func rateLimitClientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.IndexByte(v, ','); i >= 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i > 0 {
		host = host[:i]
	}
	return host
}
