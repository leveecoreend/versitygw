package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	id1, err := GenerateRequestID()
	if err != nil {
		t.Fatalf("unexpected error generating request ID: %v", err)
	}
	if len(id1) != 32 {
		t.Errorf("expected request ID length 32, got %d", len(id1))
	}

	id2, err := GenerateRequestID()
	if err != nil {
		t.Fatalf("unexpected error generating second request ID: %v", err)
	}
	if id1 == id2 {
		t.Error("expected unique request IDs, got duplicates")
	}
}

func TestRequestIDMiddleware_SetsHeader(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("x-amz-request-id") == "" {
		t.Error("expected x-amz-request-id header to be set")
	}
}

func TestRequestIDMiddleware_PreservesExistingID(t *testing.T) {
	const existingID = "my-custom-request-id-1234"

	var capturedID string
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-amz-request-id", existingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedID != existingID {
		t.Errorf("expected captured ID %q, got %q", existingID, capturedID)
	}
	if rec.Header().Get("x-amz-request-id") != existingID {
		t.Errorf("expected response header ID %q, got %q", existingID, rec.Header().Get("x-amz-request-id"))
	}
}

func TestGetRequestID_MissingContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	id := GetRequestID(req.Context())
	if id != "" {
		t.Errorf("expected empty string for missing request ID, got %q", id)
	}
}
