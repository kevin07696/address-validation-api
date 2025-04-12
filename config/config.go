package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	MaxRequests int
	TimeWindow  time.Duration
}

// GeofenceConfig holds geofencing configuration
type GeofenceConfig struct {
	MaxDistance  float64
	DistanceUnit string
	CenterLat    float64
	CenterLng    float64
}

// Config holds all configuration for the application
type Config struct {
	GoogleMapsAPIKey string
	Port             string
	RequireHTTPS     bool
	RateLimit        RateLimitConfig
	Geofence         GeofenceConfig
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found or could not be loaded: %v\n", err)
	}

	// Get Google Maps API key
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_MAPS_API_KEY environment variable is required")
	}

	// Get port or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Get HTTPS requirement or use default
	requireHTTPS := false
	if httpsStr := os.Getenv("REQUIRE_HTTPS"); httpsStr != "" {
		requireHTTPS = httpsStr == "true" || httpsStr == "1" || httpsStr == "yes"
	}

	// Get rate limit configuration or use defaults
	maxRequests := 10 // Default: 10 requests
	if maxReqStr := os.Getenv("RATE_LIMIT_MAX_REQUESTS"); maxReqStr != "" {
		if val, err := strconv.Atoi(maxReqStr); err == nil && val > 0 {
			maxRequests = val
		}
	}

	timeWindow := 60 * time.Second // Default: 1 minute
	if timeWindowStr := os.Getenv("RATE_LIMIT_TIME_WINDOW_SECONDS"); timeWindowStr != "" {
		if val, err := strconv.Atoi(timeWindowStr); err == nil && val > 0 {
			timeWindow = time.Duration(val) * time.Second
		}
	}

	// Get geofencing configuration or use defaults
	maxDistance := 10.0 // Default: 10 units
	if maxDistStr := os.Getenv("GEOFENCE_MAX_DISTANCE"); maxDistStr != "" {
		if val, err := strconv.ParseFloat(maxDistStr, 64); err == nil && val > 0 {
			maxDistance = val
		}
	}

	distanceUnit := "km" // Default: kilometers
	if unitStr := os.Getenv("GEOFENCE_DISTANCE_UNIT"); unitStr != "" {
		distanceUnit = unitStr
	}

	centerLat := 40.8448 // Default: approximate center of the Bronx
	if latStr := os.Getenv("GEOFENCE_CENTER_LAT"); latStr != "" {
		if val, err := strconv.ParseFloat(latStr, 64); err == nil {
			centerLat = val
		}
	}

	centerLng := -73.8648 // Default: approximate center of the Bronx
	if lngStr := os.Getenv("GEOFENCE_CENTER_LNG"); lngStr != "" {
		if val, err := strconv.ParseFloat(lngStr, 64); err == nil {
			centerLng = val
		}
	}

	return &Config{
		GoogleMapsAPIKey: apiKey,
		Port:             port,
		RequireHTTPS:     requireHTTPS,
		RateLimit: RateLimitConfig{
			MaxRequests: maxRequests,
			TimeWindow:  timeWindow,
		},
		Geofence: GeofenceConfig{
			MaxDistance:  maxDistance,
			DistanceUnit: distanceUnit,
			CenterLat:    centerLat,
			CenterLng:    centerLng,
		},
	}, nil
}
