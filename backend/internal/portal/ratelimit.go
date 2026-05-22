package portal

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ratelimitOrAbort checks the limiter and writes a 429 with
// Retry-After when blocked. Returns true when the request should
// continue, false when it was already rejected.
//
// The key prefix should describe the endpoint; the suffix is whatever
// scope you want to bucket on (IP for per-IP, email for per-account).
func (s *Server) ratelimitOrAbort(w http.ResponseWriter, rl *Limiter, key string) bool {
	ok, retry := rl.Allow(key)
	if ok {
		return true
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
	writeError(w, http.StatusTooManyRequests, "too many requests — try again later")
	return false
}

// rateLimitClientIP extracts the IP for rate-limiter keying. Matches
// session.go's clientIP but lives here to keep the rate-limiting code
// self-contained.
func rateLimitClientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.IndexByte(v, ','); i >= 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}
	// Strip the port for cleaner keys when behind localhost (tests).
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i > 0 {
		host = host[:i]
	}
	return host
}

// withRateLimit wraps a handler with a per-IP rate-limit check. Used
// for endpoints where the limit can be enforced before we know the
// account (signup, request-password-reset, etc.).
func (s *Server) withRateLimit(rl *Limiter, keyPrefix string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.ratelimitOrAbort(w, rl, keyPrefix+":"+rateLimitClientIP(r)) {
			return
		}
		h(w, r)
	}
}

// silence unused-import linters if time is referenced elsewhere only.
var _ = time.Second
