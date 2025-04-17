package config

import (
	"log"
	"os"
)

type Environment uint8

func (e Environment) ToString() string {
	return environmentStrings[e]
}

const (
	ENV_PRODUCTION Environment = iota
	ENV_DEVELOPMENT
)

var environmentStrings = []string{"PRODUCTION", "DEVELOPMENT"}

type InfraConfig struct {
	Environment  Environment
	Port         uint16
	IsHttpSecure bool
}

func (c Config) NewInfraConfig() InfraConfig {
	config := InfraConfig{
		Port:         8080,
		IsHttpSecure: true,
		Environment:  ENV_PRODUCTION,
	}

	const (
		PORT          = "PORT"
		ENVIRONMENT   = "ENVIRONMENT"
		REQUIRE_HTTPS = "REQUIRE_HTTPS"
	)

	// =====================
	// Port Configuration Section
	// =====================
	input := os.Getenv(PORT)
	if input == "" {
		log.Printf(MissingEnvVarWarning, PORT)
	} else {
		port, err := ParseStringToUint16(input)
		if err != nil {
			log.Printf("Invalid PORT value: %v", err)
		} else {
			// Port validation checks
			switch {
			case port == 0:
				log.Println("Port 0 is reserved")
			case port <= 1023:
				log.Println("Privileged port (1-1023) may require root access")
			case port == 65535:
				log.Println("Port 65535 often blocked by firewalls")
			default:
				config.Port = port
			}
		}
	}

	// =====================
	// HTTPS Configuration Section
	// =====================
	input = os.Getenv(REQUIRE_HTTPS)
	if input == "" {
		log.Printf(MissingEnvVarWarning, ENVIRONMENT)
	}
	config.IsHttpSecure = os.Getenv(REQUIRE_HTTPS) != "false"

	// =====================
	// Environment Configuration Section
	// =====================
	input = os.Getenv(ENVIRONMENT)
	if input == "" {
		log.Printf(MissingEnvVarWarning, ENVIRONMENT)
	} else {
		switch input {
		case ENV_DEVELOPMENT.ToString():
			config.Environment = ENV_DEVELOPMENT
		}
	}

	return config
}
