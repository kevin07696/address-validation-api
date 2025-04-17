package config

import (
	"log"
	"os"
	"regexp"
	"strings"
)

func (c Config) NewLoggerConfig(environment Environment) LoggerConfig {
	const (
		LEVEL       = "LEVEL"
		ENCODING    = "ENCODING"
		OUTPUT_PATH = "OUTPUT_PATH"
		ERROR_PATH  = "ERROR_PATH"
	)

	config := LoggerConfig{
		Level:         "info",
		Encoding:      "json",
		OutputPath:    "stdout",
		ErrorPath:     "stderr",
		IsDevelopment: false,
	}

	input := os.Getenv(LEVEL)
	if input != "" {
		switch input {
		case "info", "INFO", "debug", "DEBUG", "warn", "WARN", "error", "ERROR", "dpanic", "DPANIC", "panic", "PANIC", "fatal", "FATAL":
			config.Level = input
		default:
			log.Printf(InvalidEnvVarErr, LEVEL)
		}
	} else {
		log.Printf(MissingEnvVarWarning, LEVEL)
	}

	input = os.Getenv(ENCODING)
	if input != "" {
		switch input {
		case "json", "console":
			config.Encoding = input
		default:
			log.Printf(InvalidEnvVarErr, LEVEL)
		}
	} else {
		log.Printf(MissingEnvVarWarning, ENCODING)
	}

	setPath := func(path *string, ENV_VAR string) {
		pathPatterns := regexp.MustCompile(`^(?i)((/[^\0\r\n]+)|([a-zA-Z]:[\\/][^\0\r\n]*)|stdout|stderr|([a-z]+://[\w\-.:/]+))$`)
		input := os.Getenv(ENV_VAR)
		if input == "" {
			log.Printf(MissingEnvVarWarning, ENV_VAR)
			return
		}

		if !pathPatterns.MatchString(input) {
			log.Printf(InvalidEnvVarErr, ENV_VAR)
			return
		}

		if strings.Contains(input, "..") {
			log.Printf(InvalidEnvVarErr, ENV_VAR)
			return
		}

		*path = input
	}

	setPath(&config.OutputPath, OUTPUT_PATH)
	setPath(&config.ErrorPath, ERROR_PATH)

	if environment != ENV_PRODUCTION {
		config.IsDevelopment = true
	}

	return config
}
