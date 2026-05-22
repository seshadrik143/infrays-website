package obs

import "net/http"

// MaxBodyBytes is the default cap applied by BodyLimit. 1 MB is
// far larger than any legitimate JSON payload to the issuer or
// portal (the biggest is an enrollment-token redemption with a
// deployment fingerprint, ~2 KB) and well below any size that
// could realistically stress a Fly machine.
//
// Even for the SPA static asset routes the cap is harmless —
// browsers don't send large request bodies for GETs.
const MaxBodyBytes = 1 << 20 // 1 MiB

// BodyLimit wraps a handler with http.MaxBytesReader so request
// bodies above maxBytes return 413 Request Entity Too Large
// (specifically, when the handler tries to read past the cap, the
// next read returns an error). Apply at the root mux so every
// endpoint is protected without per-handler bookkeeping.
//
// Setting maxBytes <= 0 falls back to the default MaxBodyBytes.
func BodyLimit(maxBytes int64, next http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = MaxBodyBytes
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next.ServeHTTP(w, r)
	})
}
