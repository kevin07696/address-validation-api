package services

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strings"

	"address-validator/config"
	"address-validator/ports"

	"go.uber.org/zap"
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
	logger    *zap.Logger
	config    config.MapConfig
}

// NewAddressService creates a new address service
func NewAddressService(validator ports.AddressValidator, logger *zap.Logger, config config.MapConfig) *AddressService {
	return &AddressService{
		validator: validator,
		logger:    logger,
		config:    config,
	}
}

// ValidateAddress validates an address
func (s *AddressService) ValidateAddress(ctx context.Context, address string) (ports.AddressValidationResult, error) {

	// Sanitize the address
	cleanAddress := sanitizeAddress(address)

	// Check if address is empty after sanitization
	if cleanAddress == "" || cleanAddress == " " {
		s.logger.Warn("empty address after sanitization")
		return ports.AddressValidationResult{
			IsValid: false,
			Error:   ErrEmptyAddress.Error(),
		}, ErrEmptyAddress
	}

	// If validation passes, delegate to the external validator
	result, err := s.validator.ValidateAddress(ctx, cleanAddress)
	if err != nil {
		return result, err
	}

	s.logger.Debug("Request Completed", zap.Any("result", result))

	// Check if the address is within the geofence
	if result.IsValid {
		distance := calculateDistance(
			result.Latitude, result.Longitude,
			s.config.CenterLat, s.config.CenterLng,
			s.config.DistanceUnit,
		)
		s.logger.Debug("Checking Distance", zap.Float64("distance", distance))

		// Check if the distance is less than or equal to the maximum allowed distance
		result.InRange = distance <= s.config.MaxDistance
		s.logger.Debug("Checking Distance", zap.Bool("inRange", result.InRange))

	}

	return result, nil
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
	if strings.ToLower(unit) == ports.DISTANCE_MILES {
		distance = earthRadiusMi * c
	} else {
		// Default to kilometers
		distance = earthRadiusKm * c
	}

	return distance
}

// cleaning up spaces and only allowing words, spaces, period, comma, and dash
func sanitizeAddress(address string) string {
	// 1. Trim leading/trailing whitespace
	address = strings.TrimSpace(address)

	// 2. Collapse multiple spaces into one
	address = regexp.MustCompile(`\s+`).ReplaceAllString(address, " ")

	// 3. Remove potentially dangerous characters
	//    (keeps alphanumeric, spaces, basic punctuation)
	address = regexp.MustCompile(`[^\w\s,.-]`).ReplaceAllString(address, "")

	return address
}
