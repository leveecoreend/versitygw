package middlewares

import (
	"net/http"
	"strings"
)

// CORSConfig holds the configuration for CORS middleware.
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         string
}

// DefaultCORSConfig returns a permissive CORS configuration suitable for S3-compatible APIs.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTIONS"},
		// Including common S3 headers plus ETag and Range for partial content support
		AllowedHeaders: []string{"Authorization", "Content-Type", "Content-MD5", "x-amz-date", "x-amz-content-sha256", "x-amz-security-token", "ETag", "Range"},
		MaxAge:         "86400",
	}
}

// CORSMiddleware adds CORS headers to every response and handles preflight OPTIONS requests.
func CORSMiddleware(cfg CORSConfig) func(http.Handler) http.Handler {
	allowedOrigins := cfg.AllowedOrigins
	allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			allowedOrigin := resolveOrigin(origin, allowedOrigins)
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Max-Age", cfg.MaxAge)
			// Expose ETag so clients can use it for caching/conditional requests
			w.Header().Set("Access-Control-Expose-Headers", "ETag")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// resolveOrigin returns the matched allowed origin or "*" if all origins are permitted.
func resolveOrigin(requestOrigin string, allowedOrigins []string) string {
	for _, o := range allowedOrigins {
		if o == "*" {
			return "*"
		}
		if strings.EqualFold(o, requestOrigin) {
			return requestOrigin
		}
	}
	return ""
}
