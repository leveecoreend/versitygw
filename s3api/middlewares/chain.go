package middlewares

import (
	"net/http"
)

// Chain applies a series of middleware functions to a base handler,
// executing them in the order provided (outermost first).
func Chain(base http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		base = middlewares[i](base)
	}
	return base
}

// DefaultMiddlewareChain builds a handler with the standard set of middlewares
// applied in the recommended order for versitygw:
//  1. RecoveryMiddleware  — catch panics
//  2. RequestIDMiddleware — attach/propagate request IDs
//  3. LoggingMiddleware   — structured access logging
//  4. TimeoutMiddleware   — enforce request deadline
//  5. CORSMiddleware      — handle CORS preflight and headers
func DefaultMiddlewareChain(base http.Handler) http.Handler {
	return Chain(
		base,
		RecoveryMiddleware,
		RequestIDMiddleware,
		LoggingMiddleware,
		TimeoutMiddleware(DefaultTimeoutConfig()),
		CORSMiddleware(DefaultCORSConfig()),
	)
}
