package config

import (
	"address-validator/ports"
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
)

type MapConfig struct {
	GoogleMapsAPIKey string
	MaxDistance      float64
	DistanceUnit     string
	CenterLat        float64
	CenterLng        float64
	Country          string
	Locality         string
}

func (c Config) NewMapConfig(logger *zap.Logger) MapConfig {
	const (
		GOOGLE_MAPS_API_KEY = "GOOGLE_MAPS_API_KEY"
		MAPS_MAX_DISTANCE   = "MAP_MAX_DISTANCE"
		MAPS_DISTANCE_UNIT  = "MAP_DISTANCE_UNIT"
		MAPS_CENTER_LAT     = "MAP_CENTER_LAT"
		MAPS_CENTER_LNG     = "MAP_CENTER_LNG"
		MAPS_COUNTRY        = "MAP_COUNTRY"
		MAPS_LOCALITY       = "MAP_LOCALITY"
	)

	config := MapConfig{
		MaxDistance:  2,
		DistanceUnit: ports.DISTANCE_MILES,
		Country:      "us",
		Locality:     "Bronx",
	}

	// =====================
	// Google Maps API Key Section
	// =====================
	config.GoogleMapsAPIKey = os.Getenv(GOOGLE_MAPS_API_KEY)
	if config.GoogleMapsAPIKey == "" {
		message := fmt.Sprintf(MissingRequiredEnvVarErr, GOOGLE_MAPS_API_KEY)
		logger.Fatal(message)
	}

	// Get geofencing configuration or use defaults
	input := os.Getenv(MAPS_MAX_DISTANCE)
	if input == "" {
		message := fmt.Sprintf(MissingEnvVarWarning, MAPS_MAX_DISTANCE)
		logger.Error(message)
	} else if maxDistance, err := strconv.ParseFloat(input, 64); err == nil && maxDistance > 0 {
		config.MaxDistance = maxDistance
	}

	input = os.Getenv(MAPS_DISTANCE_UNIT)
	if input == "" {
		message := fmt.Sprintf(MissingEnvVarWarning, MAPS_DISTANCE_UNIT)
		logger.Warn(message)
	} else {
		switch input {
		case ports.DISTANCE_KILOMETER:
			config.DistanceUnit = input
		case ports.DISTANCE_MILES:
			config.DistanceUnit = input
		default:
			message := fmt.Sprintf(InvalidEnvVarErr, MAPS_DISTANCE_UNIT)
			logger.Warn(message)
		}
	}

	input = os.Getenv(MAPS_CENTER_LAT)
	if input == "" {
		message := fmt.Sprintf(MissingRequiredEnvVarErr, MAPS_CENTER_LAT)
		logger.Fatal(message)
	}

	if val, err := strconv.ParseFloat(input, 64); err == nil {
		config.CenterLat = val
	} else {
		message := fmt.Sprintf(InvalidEnvVarErr, MAPS_CENTER_LAT)
		logger.Fatal(message, zap.Error(err))
	}

	input = os.Getenv(MAPS_CENTER_LNG)
	if input == "" {
		message := fmt.Sprintf(MissingRequiredEnvVarErr, MAPS_CENTER_LNG)
		logger.Fatal(message)
	}

	if val, err := strconv.ParseFloat(input, 64); err == nil {
		config.CenterLng = val
	} else {
		message := fmt.Sprintf(InvalidEnvVarErr, MAPS_CENTER_LNG)
		logger.Fatal(message, zap.Error(err))
	}

	logger.Debug("Defined Map Configuration", zap.Any("config", config))

	return config
}
