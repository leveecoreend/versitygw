package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector()
	if mc == nil {
		t.Fatal("expected non-nil MetricsCollector")
	}
}

func TestMetricsMiddleware_RecordsRequest(t *testing.T) {
	mc := NewMetricsCollector()

	handler := MetricsMiddleware(mc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestMetricsMiddleware_RecordsLatency(t *testing.T) {
	mc := NewMetricsCollector()

	handler := MetricsMiddleware(mc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	handler.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("expected latency >= 10ms, got %v", elapsed)
	}
}

func TestMetricsMiddleware_TracksErrorStatus(t *testing.T) {
	mc := NewMetricsCollector()

	handler := MetricsMiddleware(mc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodPut, "/bucket/key", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestMetricsMiddleware_MultipleRequests(t *testing.T) {
	mc := NewMetricsCollector()

	handler := MetricsMiddleware(mc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, rr.Code)
		}
	}
}

func TestMetricsMiddleware_DifferentMethods(t *testing.T) {
	mc := NewMetricsCollector()

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodHead}

	for _, method := range methods {
		handler := MetricsMiddleware(mc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(method, "/bucket", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("method %s: expected 200, got %d", method, rr.Code)
		}
	}
}
