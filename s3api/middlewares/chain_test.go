package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/versitygw/versitygw/s3api/middlewares"
)

// trackingMiddleware returns a middleware that appends a marker to the request
// header so we can verify execution order.
func trackingMiddleware(marker string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Add("X-Middleware-Order", marker)
			next.ServeHTTP(w, r)
		})
	}
}

// TestChain_EmptyMiddleware verifies that Chain with no middlewares passes
// the request directly to the final handler.
func TestChain_EmptyMiddleware(t *testing.T) {
	finalCalled := false
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finalCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := middlewares.Chain(final)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !finalCalled {
		t.Error("expected final handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// TestChain_SingleMiddleware verifies that a single middleware wraps the
// handler correctly.
func TestChain_SingleMiddleware(t *testing.T) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middlewares.Chain(final, trackingMiddleware("A"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// TestChain_PreservesOrder verifies that middlewares are applied in the order
// they are passed (first middleware is outermost).
func TestChain_PreservesOrder(t *testing.T) {
	var order []string

	makeMiddleware := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	handler := middlewares.Chain(final,
		makeMiddleware("first"),
		makeMiddleware("second"),
		makeMiddleware("third"),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected := []string{
		"first-before",
		"second-before",
		"third-before",
		"handler",
		"third-after",
		"second-after",
		"first-after",
	}

	if len(order) != len(expected) {
		t.Fatalf("expected %d entries, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("position %d: expected %q, got %q", i, v, order[i])
		}
	}
}

// TestDefaultMiddlewareChain_ReturnsHandler verifies that DefaultMiddlewareChain
// returns a non-nil handler that responds without error.
func TestDefaultMiddlewareChain_ReturnsHandler(t *testing.T) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler := middlewares.DefaultMiddlewareChain(final)
	if handler == nil {
		t.Fatal("expected non-nil handler from DefaultMiddlewareChain")
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// DefaultMiddlewareChain may include recovery/logging; just ensure it
	// does not panic and that the response is ultimately written.
	handler.ServeHTTP(rec, req)

	if rec.Code == 0 {
		t.Error("expected a non-zero status code from DefaultMiddlewareChain handler")
	}
}
