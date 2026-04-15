package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const RequestIDKey contextKey = "requestID"

// GenerateRequestID creates a random 16-byte hex request ID.
func GenerateRequestID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RequestIDMiddleware injects a unique request ID into each request context
// and sets the x-amz-request-id response header.
// Note: if the incoming request already carries an x-amz-request-id header
// we reuse it so that client-generated IDs are preserved in logs/traces.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("x-amz-request-id")
		if requestID == "" {
			var err error
			requestID, err = GenerateRequestID()
			if err != nil {
				// Fall back to a static marker so downstream code always
				// has a non-empty string to work with.
				requestID = "unknown"
			}
		}

		w.Header().Set("x-amz-request-id", requestID)

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if no request ID has been set.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
