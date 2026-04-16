package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestCORSMiddleware_NoOrigin(t *testing.T) {
	mw := CORSMiddleware(DefaultCORSConfig())(newTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/bucket/key", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS header when Origin is absent")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCORSMiddleware_WildcardOrigin(t *testing.T) {
	mw := CORSMiddleware(DefaultCORSConfig())(newTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/bucket/key", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected '*', got %q", got)
	}
}

func TestCORSMiddleware_PreflightOptions(t *testing.T) {
	mw := CORSMiddleware(DefaultCORSConfig())(newTestHandler())
	req := httptest.NewRequest(http.MethodOptions, "/bucket/key", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

func TestCORSMiddleware_SpecificOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://trusted.com"},
		AllowedMethods: []string{"GET", "PUT"},
		AllowedHeaders: []string{"Authorization"},
		MaxAge:         "3600",
	}
	mw := CORSMiddleware(cfg)(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://trusted.com")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://trusted.com" {
		t.Errorf("expected 'https://trusted.com', got %q", got)
	}
}

func TestResolveOrigin_NoMatch(t *testing.T) {
	result := resolveOrigin("https://evil.com", []string{"https://trusted.com"})
	if result != "" {
		t.Errorf("expected empty string for unmatched origin, got %q", result)
	}
}

// TestCORSMiddleware_UnknownOriginRejected verifies that a request from an
// origin not in the allowed list does not receive CORS headers.
func TestCORSMiddleware_UnknownOriginRejected(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://trusted.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{},
		MaxAge:         "600",
	}
	mw := CORSMiddleware(cfg)(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/bucket/key", nil)
	req.Header.Set("Origin", "https://untrusted.com")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	// An unrecognized origin should not receive any CORS headers.
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header for unknown origin, got %q", got)
	}
	// The request should still be processed normally (not blocked at middleware level).
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for unknown origin, got %d", rec.Code)
	}
}
