package middlewares

import (
	"net/http"
	"time"
)

// Chain applies a list of middleware functions to a handler, in order.
// The first middleware in the list is the outermost wrapper.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// DefaultMiddlewareChain returns a handler wrapped with the standard set of
// middlewares used by versitygw: request ID, logging, recovery, CORS, timeout,
// and rate limiting.
func DefaultMiddlewareChain(h http.Handler) http.Handler {
	timeoutCfg := DefaultTimeoutConfig()
	timeoutCfg.Timeout = 30 * time.Second

	return Chain(
		h,
		RequestIDMiddleware,
		LoggingMiddleware,
		RecoveryMiddleware,
		CORSMiddleware(DefaultCORSConfig()),
		TimeoutMiddleware(timeoutCfg),
		RateLimitMiddleware(DefaultRateLimitConfig()),
	)
}
