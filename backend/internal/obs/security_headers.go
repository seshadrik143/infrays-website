package obs

import (
	"net/http"
	"strings"
)

// SecurityHeaders wraps a handler with a conservative set of defense-
// in-depth response headers suitable for a TLS-terminated portal +
// JSON API + SPA. The handler is the outermost wrapper — sits ABOVE
// HTTPMiddleware in the chain so headers apply to every response,
// including 404/405 fallbacks.
//
// Header rationale:
//
//   Strict-Transport-Security  Force HTTPS for a year incl. subdomains.
//                              preload directive opts us into the
//                              Chromium/Firefox preload list (one-time
//                              submission at hstspreload.org).
//
//   X-Content-Type-Options     Disable MIME sniffing — JSON APIs and
//                              hashed JS/CSS assets should never be
//                              guessed as anything else.
//
//   X-Frame-Options            DENY iframe embedding. Belt-and-braces
//                              alongside CSP frame-ancestors below.
//
//   Referrer-Policy            strict-origin-when-cross-origin —
//                              expose origin only to same-origin or
//                              cross-origin downgrade-safe contexts.
//
//   Permissions-Policy         Disable the long tail of browser APIs
//                              the portal doesn't use. Reduces
//                              cross-origin abuse surface.
//
//   Content-Security-Policy    Tight policy matching the Vite build:
//                              the SPA loads its own JS/CSS from
//                              /assets/* only. Data URIs for images
//                              (favicons/inline base64). 'unsafe-inline'
//                              on style-src is the one concession —
//                              Tailwind's runtime style injection
//                              sometimes uses inline tags. Once we
//                              audit a clean build with no inline
//                              styles, drop that.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy",
			"accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")
		h.Set("Content-Security-Policy", strings.Join([]string{
			"default-src 'self'",
			"script-src 'self'",
			"style-src 'self' 'unsafe-inline'",
			"img-src 'self' data:",
			"font-src 'self' data:",
			"connect-src 'self'",
			"frame-ancestors 'none'",
			"base-uri 'self'",
			"form-action 'self'",
		}, "; "))
		// Strip the Fly "Server: Fly/<hash>" leak. Fly's edge adds
		// it AFTER our handler returns, so this Del() only removes
		// what our binary set (we don't set one) — to actually scrub
		// the edge value you'd configure Fly's [http_service.response]
		// in fly.toml. Left as a separate operator step.
		h.Del("Server")
		next.ServeHTTP(w, r)
	})
}
