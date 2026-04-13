package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and returns a 500 Internal Server Error.
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID := GetRequestID(r.Context())
					stack := debug.Stack()

					logger.Error("panic recovered",
						"request_id", requestID,
						"panic", fmt.Sprintf("%v", rec),
						"stack", string(stack),
					)

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
