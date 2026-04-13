package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultTimeoutConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	if cfg.Duration != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.Duration)
	}
}

func TestTimeoutMiddleware_CompletesInTime(t *testing.T) {
	cfg := TimeoutConfig{Duration: 500 * time.Millisecond}

	handler := TimeoutMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTimeoutMiddleware_ExceedsTimeout(t *testing.T) {
	cfg := TimeoutConfig{Duration: 50 * time.Millisecond}

	handler := TimeoutMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestTimeoutMiddleware_ZeroDurationUsesDefault(t *testing.T) {
	cfg := TimeoutConfig{Duration: 0}

	handler := TimeoutMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestTimeoutMiddleware_ContextCancelledPropagated(t *testing.T) {
	cfg := TimeoutConfig{Duration: 100 * time.Millisecond}
	var ctxErr error

	handler := TimeoutMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		ctxErr = r.Context().Err()
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if ctxErr == nil {
		t.Error("expected context error, got nil")
	}
}
