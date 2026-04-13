package middlewares

import "errors"

// Authentication-related sentinel errors.
var (
	ErrMissingAuthHeader    = errors.New("missing Authorization header")
	ErrUnsupportedAuthMethod = errors.New("unsupported authentication method")
	ErrMalformedAuthHeader  = errors.New("malformed Authorization header")
)
