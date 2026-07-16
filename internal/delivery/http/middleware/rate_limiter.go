package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter implementa rate limiting por IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      rate.Limit // requests per second
	burst    int
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(requestsPerMinute, burst int) *RateLimiter {
	rps := rate.Limit(float64(requestsPerMinute) / 60.0)
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

// Limit obtiene o crea un limiter para la IP
func (rl *RateLimiter) Limit(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rps, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// Middleware devuelve el middleware de rate limiting
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetIP(r)
			limiter := rl.Limit(ip)

			if !limiter.Allow() {
				http.Error(w, `{"error": "rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetIP extrae la IP del cliente
func GetIP(r *http.Request) string {
	// Intentar obtener IP del header X-Forwarded-For (proxy)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return forwardedFor
	}

	// Intentar obtener del header X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fallback a RemoteAddr
	return r.RemoteAddr
}
