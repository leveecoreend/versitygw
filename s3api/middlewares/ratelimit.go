package middlewares

import (
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig holds configuration for the rate limiting middleware.
type RateLimitConfig struct {
	// RequestsPerSecond is the maximum number of requests allowed per second per IP.
	RequestsPerSecond int
	// BurstSize is the maximum burst size allowed.
	BurstSize int
}

// DefaultRateLimitConfig returns a sensible default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 100,
		BurstSize:         200,
	}
}

type tokenBucket struct {
	tokens    float64
	lastRefil time.Time
	mu        sync.Mutex
}

func (b *tokenBucket) allow(rate float64, burst float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefil).Seconds()
	b.tokens += elapsed * rate
	if b.tokens > burst {
		b.tokens = burst
	}
	b.lastRefil = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

type rateLimiter struct {
	buckets map[string]*tokenBucket
	mu      sync.Mutex
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{buckets: make(map[string]*tokenBucket)}
}

func (rl *rateLimiter) getBucket(ip string) *tokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if b, ok := rl.buckets[ip]; ok {
		return b
	}
	b := &tokenBucket{lastRefil: time.Now()}
	rl.buckets[ip] = b
	return b
}

// RateLimitMiddleware returns an HTTP middleware that limits requests per IP.
func RateLimitMiddleware(cfg RateLimitConfig) func(http.Handler) http.Handler {
	rl := newRateLimiter()
	rate := float64(cfg.RequestsPerSecond)
	burst := float64(cfg.BurstSize)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			bucket := rl.getBucket(ip)
			if !bucket.allow(rate, burst) {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
