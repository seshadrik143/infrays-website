// Package obs is the observability surface for the infraYS issuer.
//
// Two things live here:
//
//  1. Prometheus metrics — counters and histograms covering enrollment,
//     refresh, license issuance / revocation, portal signups + logins,
//     admin logins, Stripe webhook processing, password resets, and
//     HTTP request latency. Exposed via the /metrics handler (gated
//     by Basic Auth in obs.MetricsHandler).
//
//  2. Structured logger — a single *slog.Logger that emits JSON to
//     stderr by default. Fly captures stderr and ships it to its log
//     drains, so the JSON ends up searchable in any downstream
//     (Honeycomb, Grafana Cloud, Loki, etc.) without per-host config.
//
// Wiring:
//
//	logger := obs.NewLogger("info")          // or "debug"
//	obs.SetDefault(logger)                   // also installs as slog default
//	mux.Handle("/metrics",
//	    obs.MetricsHandler(os.Getenv("NP_METRICS_USER"),
//	                       os.Getenv("NP_METRICS_PASSWORD")))
//	mux.Handle("/foo", obs.HTTPMiddleware("foo", fooHandler))
package obs

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Counters + histograms are package-level so handler code can just
// import this package and bump them — no DI needed. They register
// against the default Prometheus registry on init.
var (
	// Issuer
	EnrollmentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_issuer_enrollments_total", Help: "Enrollment requests by outcome."},
		[]string{"outcome"}, // success | invalid_token | already_consumed | error
	)
	RefreshesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_issuer_refreshes_total", Help: "License refresh requests by outcome."},
		[]string{"outcome"}, // success | invalid_jws | revoked | expired | error
	)
	LicensesIssued = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_issuer_licenses_issued_total", Help: "Licenses minted by tier."},
		[]string{"tier"},
	)
	LicensesRevoked = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "np_issuer_licenses_revoked_total", Help: "Licenses revoked."},
	)
	JWSSignDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "np_issuer_jws_sign_seconds", Help: "JWS sign duration.",
		Buckets: prometheus.ExponentialBuckets(0.0005, 2, 10), // 0.5ms → ~256ms
	})

	// Customer portal
	PortalSignupsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_portal_signups_total", Help: "Customer signups by outcome."},
		[]string{"outcome"}, // created | claimed | conflict | invalid
	)
	PortalLoginsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_portal_logins_total", Help: "Customer logins by outcome."},
		[]string{"outcome"}, // success | invalid_credentials | suspended
	)
	PortalVerificationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_portal_verifications_total", Help: "Email verifications by outcome."},
		[]string{"outcome"}, // success | invalid_token | expired
	)
	PortalPasswordResetsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_portal_password_resets_total", Help: "Password reset events by stage and outcome."},
		[]string{"stage", "outcome"}, // stage=requested|completed, outcome=success|invalid
	)

	// Admin portal
	AdminLoginsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_admin_logins_total", Help: "Admin login events by stage and outcome."},
		[]string{"stage", "outcome"}, // stage=stage1|mfa, outcome=success|invalid_credentials|invalid_code
	)
	AdminActionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_admin_actions_total", Help: "Admin actions executed."},
		[]string{"action"}, // suspend | reactivate | offline_subscription | enrollment_token | flag_deployment | revoke_license
	)

	// Stripe
	StripeWebhookTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_stripe_webhook_total", Help: "Stripe webhook events by outcome."},
		[]string{"event_type", "outcome"}, // outcome: processed | duplicate | invalid_signature | error
	)

	// Email
	EmailSendsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_email_sends_total", Help: "Outbound email sends by type and outcome."},
		[]string{"message_type", "outcome"}, // outcome: success | failed | noop
	)

	// HTTP — labels kept narrow (no path cardinality explosion).
	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "np_http_requests_total", Help: "HTTP requests by route and status."},
		[]string{"route", "method", "status"},
	)
	HTTPDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "np_http_request_duration_seconds", Help: "HTTP request latency by route.",
		Buckets: prometheus.DefBuckets, // 0.005 → 10
	}, []string{"route"})
)

func init() {
	prometheus.MustRegister(
		EnrollmentsTotal, RefreshesTotal, LicensesIssued, LicensesRevoked, JWSSignDuration,
		PortalSignupsTotal, PortalLoginsTotal, PortalVerificationsTotal, PortalPasswordResetsTotal,
		AdminLoginsTotal, AdminActionsTotal,
		StripeWebhookTotal,
		EmailSendsTotal,
		HTTPRequests, HTTPDuration,
	)
}

// MetricsHandler returns the /metrics http.Handler protected by Basic
// Auth. When user OR password is empty, the handler always returns 404
// — metrics are simply not exposed. This means /metrics is opt-in:
// you only see it once Fly secrets are set.
//
// When both creds are present, the handler returns:
//
//	- 200 + Prometheus exposition format when the request supplies
//	  matching Basic Auth credentials
//	- 401 + WWW-Authenticate header otherwise
func MetricsHandler(user, password string) http.Handler {
	if user == "" || password == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			JSONStatus(w, http.StatusNotFound, "not found")
		})
	}
	inner := promhttp.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || !constEqual(u, user) || !constEqual(p, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		inner.ServeHTTP(w, r)
	})
}

// constEqual is a constant-time string comparison so an attacker can't
// learn the user/password length-by-length via timing. Inputs of
// unequal length still compare in time proportional to the longer
// input (no early-exit).
func constEqual(a, b string) bool {
	if len(a) != len(b) {
		// Still walk something to avoid leaking length difference via
		// runtime, but the result is fixed false.
		var z byte
		for i := 0; i < len(a) || i < len(b); i++ {
			var ai, bi byte
			if i < len(a) {
				ai = a[i]
			}
			if i < len(b) {
				bi = b[i]
			}
			z |= ai ^ bi
		}
		_ = z
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

// HTTPMiddleware wraps an http.Handler with request logging + metrics.
// routeLabel is the metric label used for HTTPRequests / HTTPDuration;
// keep it static per mount-point (e.g. "v1.enroll", "portal.api",
// "spa") to avoid cardinality explosion from raw paths.
func HTTPMiddleware(routeLabel string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		elapsed := time.Since(start)

		HTTPRequests.WithLabelValues(routeLabel, r.Method, statusBucket(rec.status)).Inc()
		HTTPDuration.WithLabelValues(routeLabel).Observe(elapsed.Seconds())

		Default().Info("http",
			"route", routeLabel,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"latency_ms", elapsed.Milliseconds(),
			"remote", clientIP(r),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.wroteHeader {
		return
	}
	s.status = code
	s.wroteHeader = true
	s.ResponseWriter.WriteHeader(code)
}

// statusBucket reduces status cardinality for the route label combo.
// 2xx/3xx/4xx/5xx is plenty for the dashboards we care about.
func statusBucket(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	case code >= 200:
		return "2xx"
	default:
		return "1xx"
	}
}

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.IndexByte(v, ','); i >= 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}
	return r.RemoteAddr
}
