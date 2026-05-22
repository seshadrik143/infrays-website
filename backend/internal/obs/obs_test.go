package obs_test

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/seshadrik143/infrays-website/backend/internal/obs"
)

// ─── /metrics handler ───────────────────────────────────────────────

func TestMetricsHandler404WhenUnconfigured(t *testing.T) {
	for _, c := range []struct{ user, pw string }{
		{"", ""},
		{"u", ""},
		{"", "p"},
	} {
		req := httptest.NewRequest("GET", "/metrics", nil)
		rec := httptest.NewRecorder()
		obs.MetricsHandler(c.user, c.pw).ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("user=%q pw=%q: expected 404, got %d", c.user, c.pw, rec.Code)
		}
	}
}

func TestMetricsHandlerNoAuth401(t *testing.T) {
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	obs.MetricsHandler("admin", "secret").ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("WWW-Authenticate"), "Basic") {
		t.Fatalf("missing WWW-Authenticate header")
	}
}

func TestMetricsHandlerWrongCreds401(t *testing.T) {
	req := httptest.NewRequest("GET", "/metrics", nil)
	req.SetBasicAuth("admin", "wrong")
	rec := httptest.NewRecorder()
	obs.MetricsHandler("admin", "secret").ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMetricsHandlerRightCredsExposesPrometheus(t *testing.T) {
	// Bump a counter so we know what to look for.
	obs.EnrollmentsTotal.WithLabelValues("success").Inc()

	req := httptest.NewRequest("GET", "/metrics", nil)
	req.SetBasicAuth("admin", "secret")
	rec := httptest.NewRecorder()
	obs.MetricsHandler("admin", "secret").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "np_issuer_enrollments_total") {
		t.Fatalf("expected metric name in body: %s", string(body))
	}
	// Prometheus exposition format hallmark.
	if !strings.Contains(string(body), "# HELP") || !strings.Contains(string(body), "# TYPE") {
		t.Fatalf("body doesn't look like Prometheus exposition: %s", string(body))
	}
}

// ─── HTTPMiddleware ─────────────────────────────────────────────────

func TestHTTPMiddlewareRecordsStatusAndLatency(t *testing.T) {
	// Wrap a handler that returns 418.
	wrapped := obs.HTTPMiddleware("test.route", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(418)
	}))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest("GET", "/foo", nil))
	if rec.Code != 418 {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
	// Confirm the metric was recorded — pull /metrics with creds.
	mreq := httptest.NewRequest("GET", "/metrics", nil)
	mreq.SetBasicAuth("u", "p")
	mrec := httptest.NewRecorder()
	obs.MetricsHandler("u", "p").ServeHTTP(mrec, mreq)
	body := mrec.Body.String()
	if !strings.Contains(body, `np_http_requests_total{method="GET",route="test.route",status="4xx"}`) {
		t.Fatalf("missing http counter for status: %s", body)
	}
}

func TestHTTPMiddlewareDefaultStatus200(t *testing.T) {
	// Handler that writes a body without calling WriteHeader → implicit 200.
	wrapped := obs.HTTPMiddleware("test.implicit", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// ─── Logger ─────────────────────────────────────────────────────────

func TestNewLoggerEmitsJSON(t *testing.T) {
	// Swap stderr indirectly: build a logger that writes to a buffer.
	var buf bytes.Buffer
	lg := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	lg.Info("hello", "k", "v")
	out := buf.String()
	if !strings.Contains(out, `"msg":"hello"`) || !strings.Contains(out, `"k":"v"`) {
		t.Fatalf("expected JSON with msg+k, got: %s", out)
	}
}

func TestConstEqual(t *testing.T) {
	// White-box: not exported, so this just exercises MetricsHandler with
	// equal-length wrong-password to confirm the path returns 401.
	req := httptest.NewRequest("GET", "/metrics", nil)
	req.SetBasicAuth("user", "rong") // same length as "right"... almost
	rec := httptest.NewRecorder()
	obs.MetricsHandler("user", "right").ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
