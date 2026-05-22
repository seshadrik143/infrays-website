package obs_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/seshadrik143/infrays-website/backend/internal/obs"
)

// ─── BodyLimit ──────────────────────────────────────────────────────

func TestBodyLimitRejectsOversizeBody(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to read past the limit; MaxBytesReader returns an error.
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(200)
	})
	wrapped := obs.BodyLimit(64, inner) // 64-byte cap for the test
	rec := httptest.NewRecorder()
	body := strings.NewReader(strings.Repeat("A", 1024))
	wrapped.ServeHTTP(rec, httptest.NewRequest("POST", "/", body))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestBodyLimitAllowsUnderCap(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(204)
	})
	wrapped := obs.BodyLimit(1024, inner)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest("POST", "/", bytes.NewReader([]byte("small"))))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

// ─── TrailingSlash ──────────────────────────────────────────────────

func TestTrailingSlashStripsAPIPaths(t *testing.T) {
	hit := ""
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/portal/auth/login", func(_ http.ResponseWriter, r *http.Request) {
		hit = r.URL.Path
	})
	wrapped := obs.TrailingSlash(mux)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest("POST", "/api/portal/auth/login/", nil))
	if hit != "/api/portal/auth/login" {
		t.Fatalf("expected normalized path, got %q", hit)
	}
}

func TestTrailingSlashLeavesNonAPIAlone(t *testing.T) {
	hit := ""
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login/", func(_ http.ResponseWriter, r *http.Request) {
		hit = r.URL.Path
	})
	wrapped := obs.TrailingSlash(mux)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest("GET", "/login/", nil))
	if hit != "/login/" {
		t.Fatalf("expected un-normalized path, got %q", hit)
	}
}

// ─── JSON404 ────────────────────────────────────────────────────────

func TestJSON404OnUnmatchedRoute(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /foo", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	wrapped := obs.JSON404(mux)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest("GET", "/missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected JSON content-type, got %q", rec.Header().Get("Content-Type"))
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "not found" {
		t.Fatalf("expected {error:not found}, got %v", body)
	}
}

func TestJSON404PassesThroughMatchedRoute(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /foo", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("hello"))
	})
	wrapped := obs.JSON404(mux)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest("GET", "/foo", nil))
	if rec.Code != 200 || rec.Body.String() != "hello" {
		t.Fatalf("matched route should pass through: %d %q", rec.Code, rec.Body.String())
	}
}

// ─── DenyCrossOrigin ────────────────────────────────────────────────

func TestDenyCrossOriginRejectsOPTIONS(t *testing.T) {
	wrapped := obs.DenyCrossOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("inner handler should not be called for OPTIONS")
	}))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest("OPTIONS", "/api/portal/auth/login", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") == "" {
		t.Fatal("expected Allow header")
	}
	// No Access-Control-Allow-Origin header — browsers reject preflight.
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("Access-Control-Allow-Origin should not be set")
	}
}

func TestDenyCrossOriginPassesThroughOtherMethods(t *testing.T) {
	called := false
	wrapped := obs.DenyCrossOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(204)
	}))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest("POST", "/api/portal/auth/login", nil))
	if !called {
		t.Fatal("inner handler should be called for POST")
	}
	if rec.Code != 204 {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
