package config

import (
	"log"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct{}

// LoadConfig loads the configuration from environment variables
func LoadConfig() Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Warning: .env file not found or could not be loaded: %v\n", err)
	}

	return Config{}
}
