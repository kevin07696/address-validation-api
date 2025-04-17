package handlers

import (
	"address-validator/config"
	"sync"
	"time"
)

// RateLimiter provides a simple rate limiting mechanism
type RateLimiter struct {
	requests    map[string][]time.Time
	maxRequests uint
	timeWindow  time.Duration
	mu          sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string][]time.Time),
		maxRequests: config.MaxRequests,
		timeWindow:  config.TimeWindow,
	}
}

// Allow checks if a request is allowed based on the rate limit
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Remove old requests outside the time window
	var validRequests []time.Time
	for _, t := range rl.requests[ip] {
		if now.Sub(t) <= rl.timeWindow {
			validRequests = append(validRequests, t)
		}
	}

	// Update requests for this IP
	rl.requests[ip] = validRequests

	// Check if rate limit is exceeded
	if len(validRequests) >= int(rl.maxRequests) {
		return false
	}

	// Add current request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}
