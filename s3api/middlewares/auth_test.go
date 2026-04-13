package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseAuthHeader_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, err := ParseAuthHeader(req)
	if err != ErrMissingAuthHeader {
		t.Fatalf("expected ErrMissingAuthHeader, got %v", err)
	}
}

func TestParseAuthHeader_UnsupportedMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AuthorizationHeader, "Bearer sometoken")
	_, err := ParseAuthHeader(req)
	if err != ErrUnsupportedAuthMethod {
		t.Fatalf("expected ErrUnsupportedAuthMethod, got %v", err)
	}
}

func TestParseAuthHeader_ValidV4(t *testing.T) {
	validHeader := `AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20230101/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-date, Signature=abcdef1234567890`
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AuthorizationHeader, validHeader)

	result, err := ParseAuthHeader(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessKey != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("expected access key AKIAIOSFODNN7EXAMPLE, got %s", result.AccessKey)
	}
	if result.Region != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", result.Region)
	}
	if result.Service != "s3" {
		t.Errorf("expected service s3, got %s", result.Service)
	}
	if result.Signature != "abcdef1234567890" {
		t.Errorf("expected signature abcdef1234567890, got %s", result.Signature)
	}
	if len(result.SignedHeaders) != 2 {
		t.Errorf("expected 2 signed headers, got %d", len(result.SignedHeaders))
	}
}

func TestAuthMiddleware_PassesWithoutHeader(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	AuthMiddleware(next).ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
}

func TestAuthMiddleware_RejectsMalformed(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AuthorizationHeader, "AWS4-HMAC-SHA256 BadData")
	rec := httptest.NewRecorder()
	AuthMiddleware(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
