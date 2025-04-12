package adapters

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"address-validator/config"
	"address-validator/ports"

	"googlemaps.github.io/maps"
)

// Constants for Earth's radius in different units
const (
	earthRadiusKm = 6371.0 // Earth's radius in kilometers
	earthRadiusMi = 3958.8 // Earth's radius in miles
)

// GoogleMapsClient defines the interface for a Google Maps client
type GoogleMapsClient interface {
	Geocode(ctx context.Context, r *maps.GeocodingRequest) ([]maps.GeocodingResult, error)
}

// GoogleMapsAdapter implements the ports.AddressValidator interface using Google Maps API
type GoogleMapsAdapter struct {
	client GoogleMapsClient
	logger *slog.Logger
	config *config.Config
}

// NewGoogleMapsAdapter creates a new Google Maps adapter
func NewGoogleMapsAdapter(cfg *config.Config, logger *slog.Logger) (*GoogleMapsAdapter, error) {
	client, err := maps.NewClient(maps.WithAPIKey(cfg.GoogleMapsAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Maps client: %w", err)
	}
	return &GoogleMapsAdapter{
		client: client,
		logger: logger,
		config: cfg,
	}, nil
}

// calculateDistance calculates the distance between two points using the Haversine formula
func (gma *GoogleMapsAdapter) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Convert latitude and longitude from degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLng := lng2Rad - lng1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Calculate distance based on the unit
	var distance float64
	if gma.config.Geofence.DistanceUnit == "km" {
		distance = earthRadiusKm * c
	} else {
		distance = earthRadiusMi * c
	}

	return distance
}

// ValidateAddress validates an address using Google Maps API
func (gma *GoogleMapsAdapter) ValidateAddress(ctx context.Context, address string) (ports.AddressValidationResult, error) {
	// Create result object
	result := ports.AddressValidationResult{
		IsValid: false,
	}

	gma.logger.Debug("validating address with Google Maps API", "address", address)

	// Check if address is empty
	if address == "" {
		gma.logger.Warn("empty address provided")
		result.Error = "Address is required"
		return result, fmt.Errorf("address is required")
	}

	// Call Google Maps Geocoding API
	r := &maps.GeocodingRequest{
		Address: address,
	}

	gma.logger.Debug("calling Google Maps API", "request", r)
	geocodeResults, err := gma.client.Geocode(ctx, r)
	if err != nil {
		gma.logger.Error("geocoding error", "error", err)
		result.Error = "Failed to validate address: " + err.Error()
		return result, fmt.Errorf("geocoding error: %w", err)
	}

	// Check if any results were returned
	if len(geocodeResults) == 0 {
		gma.logger.Warn("no results found for address")
		result.Error = "Address not found"
		return result, fmt.Errorf("address not found")
	}

	// Get the first result
	geocodeResult := geocodeResults[0]

	// Set result fields
	result.IsValid = true
	result.FormattedAddress = geocodeResult.FormattedAddress
	result.Latitude = geocodeResult.Geometry.Location.Lat
	result.Longitude = geocodeResult.Geometry.Location.Lng

	gma.logger.Info("address validated successfully",
		"formatted_address", result.FormattedAddress,
		"latitude", result.Latitude,
		"longitude", result.Longitude)

	// Check if the address is within the geofence
	distance := gma.calculateDistance(
		result.Latitude,
		result.Longitude,
		gma.config.Geofence.CenterLat,
		gma.config.Geofence.CenterLng,
	)

	// Set InRange field based on the distance
	result.InRange = distance <= gma.config.Geofence.MaxDistance

	gma.logger.Info("geofencing check completed",
		"distance", distance,
		"max_distance", gma.config.Geofence.MaxDistance,
		"unit", gma.config.Geofence.DistanceUnit,
		"in_range", result.InRange)

	return result, nil
}
