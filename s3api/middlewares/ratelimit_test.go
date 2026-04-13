package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultRateLimitConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()
	if cfg.RequestsPerSecond != 100 {
		t.Errorf("expected RequestsPerSecond=100, got %d", cfg.RequestsPerSecond)
	}
	if cfg.BurstSize != 200 {
		t.Errorf("expected BurstSize=200, got %d", cfg.BurstSize)
	}
}

func TestRateLimitMiddleware_AllowsNormalTraffic(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 100, BurstSize: 10}
	middleware := RateLimitMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware_BlocksExcessiveTraffic(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 1, BurstSize: 1}
	middleware := RateLimitMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"

	// First request should succeed (burst allows 1)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Errorf("expected first request to succeed, got %d", rr1.Code)
	}

	// Subsequent requests should be rate limited
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rr2.Code)
	}
}

func TestRateLimitMiddleware_DifferentIPsIndependent(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 1, BurstSize: 1}
	middleware := RateLimitMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, ip := range []string{"1.1.1.1:80", "2.2.2.2:80", "3.3.3.3:80"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("ip %s: expected 200, got %d", ip, rr.Code)
		}
	}
}
