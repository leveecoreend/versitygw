package middlewares

import (
	"net/http"
)

type contextKey string

const authResultKey contextKey = "authResult"

// AuthMiddleware validates the presence and basic structure of AWS auth headers.
// It does not verify the signature — that is left to the backend.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow anonymous requests for operations that may not require auth
		// (e.g. pre-signed URLs handled downstream).
		authHeader := r.Header.Get(AuthorizationHeader)
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		result, err := ParseAuthHeader(r)
		if err != nil {
			requestID := GetRequestID(r.Context())
			http.Error(w, buildAuthErrorXML(err.Error(), requestID), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		ctx = setAuthResult(ctx, result)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAuthResult retrieves the parsed AuthResult from the request context.
// Returns nil if no auth result has been stored.
func GetAuthResult(r *http.Request) *AuthResult {
	val := r.Context().Value(authResultKey)
	if val == nil {
		return nil
	}
	result, ok := val.(*AuthResult)
	if !ok {
		return nil
	}
	return result
}

func setAuthResult(ctx interface{ Value(interface{}) interface{} }, result *AuthResult) interface{ Value(interface{}) interface{} } {
	return contextWithValue(ctx, result)
}

func buildAuthErrorXML(msg, requestID string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		"<Error><Code>InvalidRequest</Code><Message>" + msg +
		"</Message><RequestId>" + requestID + "</RequestId></Error>"
}
