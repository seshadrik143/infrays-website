package obs

import "net/http"

// DenyCrossOrigin makes the "no CORS" stance explicit. The portal +
// admin APIs use cookie-based auth from the same origin only — no
// browser at another origin should ever talk to them. Without this
// middleware, the implicit denial happens via stdlib's 405 (no
// OPTIONS handler registered), which would silently flip to
// permitting cross-origin requests if anyone later mounts an OPTIONS
// handler on any route.
//
// We respond to OPTIONS explicitly with 405 + Allow header listing
// the methods this binary actually uses. No Access-Control-* headers
// are ever set, so browsers reject preflights.
func DenyCrossOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Allow", "GET, POST, PATCH, DELETE")
			JSONStatus(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		next.ServeHTTP(w, r)
	})
}
