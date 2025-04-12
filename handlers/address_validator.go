package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"address-validator/config"
	"address-validator/services"
)

// AddressRequest represents the incoming request for address validation
type AddressRequest struct {
	Address string `json:"address"`
}

// RateLimiter provides a simple rate limiting mechanism
type RateLimiter struct {
	requests    map[string][]time.Time
	maxRequests int
	timeWindow  time.Duration
	mu          sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, timeWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		timeWindow:  timeWindow,
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
	if len(validRequests) >= rl.maxRequests {
		return false
	}

	// Add current request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

// AddressHandler handles HTTP requests for address validation
type AddressHandler struct {
	service     *services.AddressService
	rateLimiter *RateLimiter
	logger      *slog.Logger
	config      *config.Config
}

// NewAddressHandler creates a new address handler
func NewAddressHandler(service *services.AddressService, config *config.Config, logger *slog.Logger) *AddressHandler {
	// Create a rate limiter with values from config
	rateLimiter := NewRateLimiter(config.RateLimit.MaxRequests, config.RateLimit.TimeWindow)

	return &AddressHandler{
		service:     service,
		rateLimiter: rateLimiter,
		logger:      logger,
		config:      config,
	}
}

// ValidateAddress handles the address validation endpoint
func (h *AddressHandler) ValidateAddress(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST requests
	if r.Method != http.MethodPost {
		h.logger.Warn("method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only allow HTTPS
	if h.config.RequireHTTPS && r.TLS == nil {
		h.logger.Warn("HTTPS required")
		http.Error(w, "HTTPS required", http.StatusBadRequest)
		return
	}

	// Get client IP for rate limiting
	clientIP := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}

	// Check rate limit
	if !h.rateLimiter.Allow(clientIP) {
		h.logger.Warn("rate limit exceeded", "ip", clientIP)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Parse request body
	var req AddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("received address validation request", "address", req.Address, "ip", clientIP)

	// Validate address using the service
	result, err := h.service.ValidateAddress(r.Context(), req.Address)

	// Return response with appropriate status code
	if err != nil {
		h.logger.Warn("address validation failed", "error", err)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		h.logger.Info("address validation successful", "formatted_address", result.FormattedAddress)
	}

	// Encode response
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
