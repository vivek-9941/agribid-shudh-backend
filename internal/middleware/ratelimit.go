package middleware

import (
	"net/http"
	"sync"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"golang.org/x/time/rate"
)

// RateLimiter provides per-key rate limiting using token bucket algorithm.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rpm      int
}

// NewRateLimiter creates a new rate limiter with the given requests-per-minute.
func NewRateLimiter(rpm int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rpm:      rpm,
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	l, ok := rl.limiters[key]
	rl.mu.RUnlock()
	if ok {
		return l
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock.
	if l, ok := rl.limiters[key]; ok {
		return l
	}

	// rate.Every converts per-minute to per-second interval.
	l = rate.NewLimiter(rate.Limit(float64(rl.rpm)/60.0), rl.rpm)
	rl.limiters[key] = l
	return l
}

// RateLimit returns middleware that rate-limits requests by IP address.
func RateLimit(rpm int) func(http.Handler) http.Handler {
	rl := NewRateLimiter(rpm)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP for auth routes, user ID for authenticated routes.
			key := r.RemoteAddr
			if userCtx := GetUserContext(r.Context()); userCtx != nil {
				key = userCtx.UserID.String()
			}

			if !rl.getLimiter(key).Allow() {
				response.Error(w, apperrors.UnprocessableEntity(
					"RATE_LIMIT_EXCEEDED",
					"too many requests, please try again later",
					map[string]interface{}{"limit_rpm": rpm},
				))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
