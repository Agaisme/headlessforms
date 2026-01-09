package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter per IP
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	limit    int           // Max requests
	window   time.Duration // Time window
}

type visitor struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	// Cleanup old entries periodically
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastReset) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:    rl.limit,
			lastReset: time.Now(),
		}
		rl.visitors[ip] = v
		return v
	}

	// Reset tokens if window has passed
	if time.Since(v.lastReset) > rl.window {
		v.tokens = rl.limit
		v.lastReset = time.Now()
	}

	return v
}

func (rl *RateLimiter) allow(ip string) bool {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if v.tokens > 0 {
		v.tokens--
		return true
	}
	return false
}

// Middleware returns HTTP middleware that rate limits requests
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)

			if !rl.allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"status":"fail","message":"Rate limit exceeded. Please try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getIP extracts the client IP from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Default rate limiters for different purposes
var (
	// PublicLimiter for public endpoints (submissions) - 100 req/min
	PublicLimiter = NewRateLimiter(100, time.Minute)

	// AuthLimiter for auth endpoints (login/register) - 10 req/min
	AuthLimiter = NewRateLimiter(10, time.Minute)

	// APILimiter for general API endpoints - 200 req/min
	APILimiter = NewRateLimiter(200, time.Minute)
)
