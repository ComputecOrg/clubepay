package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

type rateLimitEntry struct {
	count    int
	expireAt time.Time
}

type rateLimiter struct {
	mu          sync.Mutex
	entries     map[string]*rateLimitEntry
	maxRequests int
	window      time.Duration
}

func newRateLimiter(maxRequests int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		entries:     make(map[string]*rateLimitEntry),
		maxRequests: maxRequests,
		window:      window,
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.entries {
			if now.After(entry.expireAt) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// allow checks the rate limit for the given IP.
// Returns (allowed bool, retryAfterSeconds int).
func (rl *rateLimiter) allow(ip string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[ip]

	if !exists || now.After(entry.expireAt) {
		// New window
		rl.entries[ip] = &rateLimitEntry{
			count:    1,
			expireAt: now.Add(rl.window),
		}
		return true, 0
	}

	if entry.count >= rl.maxRequests {
		retryAfter := int(entry.expireAt.Sub(now).Seconds()) + 1
		return false, retryAfter
	}

	entry.count++
	return true, 0
}

// RateLimit returns a middleware that limits requests per IP.
// maxRequests is the maximum number of requests allowed per window.
// window is the duration of the rate limit window.
func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(maxRequests, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				// Fallback: use RemoteAddr as-is if parsing fails
				ip = r.RemoteAddr
			}

			allowed, retryAfter := rl.allow(ip)
			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				writeErr(w, http.StatusTooManyRequests, "too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
