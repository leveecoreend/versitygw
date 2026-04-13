package middlewares

import "net/http"

// MiddlewareFunc is a function that wraps an http.Handler.
type MiddlewareFunc func(http.Handler) http.Handler

// Chain composes multiple middleware functions into a single http.Handler.
// Middleware is applied in the order provided — the first middleware in the
// slice is the outermost (executes first on request, last on response).
//
// Example:
//
//	handler := Chain(
//	    myHandler,
//	    RecoveryMiddleware,
//	    RequestIDMiddleware,
//	    LoggingMiddleware,
//	    CORSMiddleware(DefaultCORSConfig()),
//	)
func Chain(h http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// DefaultMiddlewareChain returns the standard middleware stack used by versitygw.
// Middleware order (outermost to innermost):
//  1. RecoveryMiddleware  — catches panics before anything else can fail
//  2. RequestIDMiddleware — attaches a request ID for tracing
//  3. LoggingMiddleware   — logs with the request ID already set
//  4. CORSMiddleware      — handles CORS headers and preflight
func DefaultMiddlewareChain(h http.Handler) http.Handler {
	return Chain(
		h,
		RecoveryMiddleware,
		RequestIDMiddleware,
		LoggingMiddleware,
		CORSMiddleware(DefaultCORSConfig()),
	)
}
