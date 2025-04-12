package adapters

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"address-validator/config"
	"address-validator/ports"

	"googlemaps.github.io/maps"
)

// MockGoogleMapsClient is a mock implementation of the Google Maps client for testing
type MockGoogleMapsClient struct {
	geocodeResults []maps.GeocodingResult
	geocodeError   error
}

// Geocode is a mock implementation of the Geocode method
func (m *MockGoogleMapsClient) Geocode(_ context.Context, _ *maps.GeocodingRequest) ([]maps.GeocodingResult, error) {
	return m.geocodeResults, m.geocodeError
}

// TestCalculateDistance tests the calculateDistance method
func TestCalculateDistance(t *testing.T) {
	// Create a logger for testing
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a config for testing
	cfg := &config.Config{
		Geofence: config.GeofenceConfig{
			DistanceUnit: "km",
		},
	}

	// Create a Google Maps adapter for testing
	adapter := &GoogleMapsAdapter{
		logger: logger,
		config: cfg,
	}

	// Test cases
	testCases := []struct {
		name           string
		lat1           float64
		lng1           float64
		lat2           float64
		lng2           float64
		expectedResult float64
		distanceUnit   string
	}{
		{
			name:           "Same point",
			lat1:           40.7128,
			lng1:           -74.0060,
			lat2:           40.7128,
			lng2:           -74.0060,
			expectedResult: 0,
			distanceUnit:   "km",
		},
		{
			name:           "New York to Los Angeles (km)",
			lat1:           40.7128,
			lng1:           -74.0060,
			lat2:           34.0522,
			lng2:           -118.2437,
			expectedResult: 3935.9,
			distanceUnit:   "km",
		},
		{
			name:           "New York to Los Angeles (mi)",
			lat1:           40.7128,
			lng1:           -74.0060,
			lat2:           34.0522,
			lng2:           -118.2437,
			expectedResult: 2445.5,
			distanceUnit:   "mi",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the distance unit
			adapter.config.Geofence.DistanceUnit = tc.distanceUnit

			// Calculate the distance
			result := adapter.calculateDistance(tc.lat1, tc.lng1, tc.lat2, tc.lng2)

			// Check if the result is within 0.5 of the expected result
			if result < tc.expectedResult-0.5 || result > tc.expectedResult+0.5 {
				t.Errorf("Expected distance to be around %.1f %s, but got %.1f %s",
					tc.expectedResult, tc.distanceUnit, result, tc.distanceUnit)
			}
		})
	}
}

// TestValidateAddress tests the ValidateAddress method
func TestValidateAddress(t *testing.T) {
	// Create a logger for testing
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Test cases
	testCases := []struct {
		name           string
		address        string
		geocodeResults []maps.GeocodingResult
		geocodeError   error
		geofenceConfig config.GeofenceConfig
		expectedResult ports.AddressValidationResult
		expectedError  bool
	}{
		{
			name:    "Empty address",
			address: "",
			expectedResult: ports.AddressValidationResult{
				IsValid: false,
				Error:   "Address is required",
			},
			expectedError: true,
		},
		{
			name:           "No results found",
			address:        "Invalid Address",
			geocodeResults: []maps.GeocodingResult{},
			expectedResult: ports.AddressValidationResult{
				IsValid: false,
				Error:   "Address not found",
			},
			expectedError: true,
		},
		{
			name:    "Address within geofence",
			address: "123 Main St, Bronx, NY",
			geocodeResults: []maps.GeocodingResult{
				{
					FormattedAddress: "123 Main St, Bronx, NY 10456, USA",
					Geometry: maps.AddressGeometry{
						Location: maps.LatLng{
							Lat: 40.8448, // Same as geofence center
							Lng: -73.8648,
						},
					},
				},
			},
			geofenceConfig: config.GeofenceConfig{
				MaxDistance:  10,
				DistanceUnit: "km",
				CenterLat:    40.8448, // Center of the Bronx
				CenterLng:    -73.8648,
			},
			expectedResult: ports.AddressValidationResult{
				IsValid:          true,
				FormattedAddress: "123 Main St, Bronx, NY 10456, USA",
				Latitude:         40.8448,
				Longitude:        -73.8648,
				InRange:          true,
			},
			expectedError: false,
		},
		{
			name:    "Address outside geofence",
			address: "123 Main St, Manhattan, NY",
			geocodeResults: []maps.GeocodingResult{
				{
					FormattedAddress: "123 Main St, Manhattan, NY 10001, USA",
					Geometry: maps.AddressGeometry{
						Location: maps.LatLng{
							Lat: 40.7128, // Manhattan (far from the Bronx)
							Lng: -74.0060,
						},
					},
				},
			},
			geofenceConfig: config.GeofenceConfig{
				MaxDistance:  10,
				DistanceUnit: "km",
				CenterLat:    40.8448, // Center of the Bronx
				CenterLng:    -73.8648,
			},
			expectedResult: ports.AddressValidationResult{
				IsValid:          true,
				FormattedAddress: "123 Main St, Manhattan, NY 10001, USA",
				Latitude:         40.7128,
				Longitude:        -74.0060,
				InRange:          false,
			},
			expectedError: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a config for testing
			cfg := &config.Config{
				Geofence: tc.geofenceConfig,
			}

			// Create a mock client
			mockClient := &MockGoogleMapsClient{
				geocodeResults: tc.geocodeResults,
				geocodeError:   tc.geocodeError,
			}

			// Create a Google Maps adapter for testing
			adapter := &GoogleMapsAdapter{
				client: mockClient,
				logger: logger,
				config: cfg,
			}

			// Validate the address
			result, err := adapter.ValidateAddress(context.Background(), tc.address)

			// Check if an error was expected
			if tc.expectedError && err == nil {
				t.Errorf("Expected an error, but got nil")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			// Check the result
			if result.IsValid != tc.expectedResult.IsValid {
				t.Errorf("Expected IsValid to be %v, but got %v", tc.expectedResult.IsValid, result.IsValid)
			}
			if result.FormattedAddress != tc.expectedResult.FormattedAddress {
				t.Errorf("Expected FormattedAddress to be %s, but got %s", tc.expectedResult.FormattedAddress, result.FormattedAddress)
			}
			if result.Latitude != tc.expectedResult.Latitude {
				t.Errorf("Expected Latitude to be %f, but got %f", tc.expectedResult.Latitude, result.Latitude)
			}
			if result.Longitude != tc.expectedResult.Longitude {
				t.Errorf("Expected Longitude to be %f, but got %f", tc.expectedResult.Longitude, result.Longitude)
			}
			if result.InRange != tc.expectedResult.InRange {
				t.Errorf("Expected InRange to be %v, but got %v", tc.expectedResult.InRange, result.InRange)
			}
			if result.Error != tc.expectedResult.Error {
				t.Errorf("Expected Error to be %s, but got %s", tc.expectedResult.Error, result.Error)
			}
		})
	}
}
