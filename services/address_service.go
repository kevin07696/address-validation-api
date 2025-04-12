package services

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"regexp"
	"strings"

	"address-validator/config"
	"address-validator/ports"
)

// Common validation errors
var (
	ErrEmptyAddress      = errors.New("address is empty")
	ErrSuspiciousPattern = errors.New("suspicious address detected")
	ErrOutsideGeofence   = errors.New("address outside allowed geographic area")
)

// earthRadiusKm is the radius of the Earth in kilometers
const earthRadiusKm = 6371.0

// earthRadiusMi is the radius of the Earth in miles
const earthRadiusMi = 3958.8

// AddressService handles address validation business logic
type AddressService struct {
	validator ports.AddressValidator
	logger    *slog.Logger
	config    *config.Config
}

// NewAddressService creates a new address service
func NewAddressService(validator ports.AddressValidator, logger *slog.Logger, config *config.Config) *AddressService {
	return &AddressService{
		validator: validator,
		logger:    logger,
		config:    config,
	}
}

// ValidateAddress validates an address
func (s *AddressService) ValidateAddress(ctx context.Context, address string) (ports.AddressValidationResult, error) {
	s.logger.Info("validating address in service", "address", address)

	// Sanitize the address
	cleanAddress := sanitizeAddress(address)

	// Check if address is empty after sanitization
	if cleanAddress == "" {
		s.logger.Warn("empty address after sanitization")
		return ports.AddressValidationResult{
			IsValid: false,
			Error:   ErrEmptyAddress.Error(),
		}, ErrEmptyAddress
	}

	// Check for suspicious patterns
	if isSuspicious(cleanAddress) {
		s.logger.Warn("suspicious address pattern detected", "address", cleanAddress)
		return ports.AddressValidationResult{
			IsValid: false,
			Error:   ErrSuspiciousPattern.Error(),
		}, ErrSuspiciousPattern
	}

	// If validation passes, delegate to the external validator
	s.logger.Debug("delegating to external validator", "clean_address", cleanAddress)
	result, err := s.validator.ValidateAddress(ctx, cleanAddress)
	if err != nil {
		return result, err
	}

	// Check if the address is within the geofence
	if result.IsValid {
		s.logger.Debug("checking geofence",
			"lat", result.Latitude,
			"lng", result.Longitude,
			"center_lat", s.config.Geofence.CenterLat,
			"center_lng", s.config.Geofence.CenterLng,
			"max_distance", s.config.Geofence.MaxDistance,
			"unit", s.config.Geofence.DistanceUnit)

		inRange := s.isWithinGeofence(result.Latitude, result.Longitude)
		result.InRange = inRange

		if !inRange {
			s.logger.Warn("address outside geofence",
				"lat", result.Latitude,
				"lng", result.Longitude)
		}
	}

	return result, nil
}

// isWithinGeofence checks if a location is within the configured geofence
func (s *AddressService) isWithinGeofence(lat, lng float64) bool {
	// Calculate distance between the point and the center of the geofence
	distance := calculateDistance(
		lat, lng,
		s.config.Geofence.CenterLat, s.config.Geofence.CenterLng,
		s.config.Geofence.DistanceUnit,
	)

	// Check if the distance is less than or equal to the maximum allowed distance
	return distance <= s.config.Geofence.MaxDistance
}

// calculateDistance calculates the distance between two points using the Haversine formula
func calculateDistance(lat1, lng1, lat2, lng2 float64, unit string) float64 {
	// Convert latitude and longitude from degrees to radians
	lat1Rad := lat1 * (math.Pi / 180.0)
	lng1Rad := lng1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)
	lng2Rad := lng2 * (math.Pi / 180.0)

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLng := lng2Rad - lng1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Calculate distance based on unit
	var distance float64
	if strings.ToLower(unit) == "mi" {
		distance = earthRadiusMi * c
	} else {
		// Default to kilometers
		distance = earthRadiusKm * c
	}

	return distance
}

// sanitizeAddress removes special characters, trims whitespace, and limits length
func sanitizeAddress(address string) string {
	// Remove HTML tags and special characters (keep basic address chars)
	specialCharsPattern := regexp.MustCompile(`[<>{}[\];'"\\]`)
	cleaned := specialCharsPattern.ReplaceAllString(address, "")

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Limit length to 200 characters
	if len(cleaned) > 200 {
		cleaned = cleaned[:200]
	}

	return cleaned
}

// isSuspicious checks for suspicious patterns in the address
func isSuspicious(address string) bool {
	// Check for:
	// - Excessive repeating chars (aaaaaa)
	// - Binary/hex patterns
	// - SQL/JS snippets
	suspiciousPattern := regexp.MustCompile(`(?i)((.)\2{5}|0x[0-9a-f]+|select\s.+\sfrom|script|alert\(|function|var\s|let\s|const\s)`)
	return suspiciousPattern.MatchString(address)
}
