package middlewares

import (
	"context"
	"net/http"
	"time"
)

// TimeoutConfig holds configuration for the timeout middleware.
type TimeoutConfig struct {
	// Duration is the maximum time allowed for a request to complete.
	Duration time.Duration
}

// DefaultTimeoutConfig returns a TimeoutConfig with sensible defaults.
// Using 60s instead of 30s since large object uploads were timing out in practice.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Duration: 60 * time.Second,
	}
}

// TimeoutMiddleware returns a middleware that cancels requests exceeding the
// configured duration, responding with 503 Service Unavailable.
func TimeoutMiddleware(cfg TimeoutConfig) func(http.Handler) http.Handler {
	if cfg.Duration <= 0 {
		cfg = DefaultTimeoutConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), cfg.Duration)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{})
			go func() {
				defer close(done)
				next.ServeHTTP(w, r)
			}()

			select {
			case <-done:
				// request completed normally
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					w.WriteHeader(http.StatusServiceUnavailable)
				}
			}
		})
	}
}
