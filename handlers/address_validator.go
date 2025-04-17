package handlers

import (
	"encoding/json"
	"net/http"

	"address-validator/config"
	"address-validator/services"

	"go.uber.org/zap"
)

// AddressRequest represents the incoming request for address validation
type AddressRequest struct {
	Address string `json:"address"`
}

// AddressHandler handles HTTP requests for address validation
type AddressHandler struct {
	service     *services.AddressService
	rateLimiter *RateLimiter
	logger      *zap.Logger
	config      config.InfraConfig
}

// NewAddressHandler creates a new address handler
func NewAddressHandler(service *services.AddressService, rateLimiter *RateLimiter, config config.InfraConfig, logger *zap.Logger) *AddressHandler {

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

	// Only allow POST requests for edge-cases where a user can add special characters like # for apts
	if r.Method != http.MethodPost {
		h.logger.Warn("method not allowed", zap.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only allow HTTPS
	if h.config.IsHttpSecure && r.TLS == nil {
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
		h.logger.Warn("rate limit exceeded", zap.String("ip", clientIP))
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Parse request body
	var req AddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate address using the service
	result, err := h.service.ValidateAddress(r.Context(), req.Address)

	// Return response with appropriate status code
	if err != nil {
		h.logger.Warn("address validation failed", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
	}
	// Encode response
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
