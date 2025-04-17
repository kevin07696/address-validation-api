package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	MaxRequests uint
	TimeWindow  time.Duration
}

func (c Config) NewRateLimitConfig(logger *zap.Logger) RateLimitConfig {
	// Environment variable constants
	const (
		RATE_LIMIT_MAX_REQUESTS = "RATE_LIMIT_MAX_REQUESTS"
		RATE_LIMIT_TIME_WINDOW  = "RATE_LIMIT_TIME_WINDOW_SECONDS"
		INPUT                   = "input"
	)

	config := RateLimitConfig{
		MaxRequests: 10,
		TimeWindow:  60 * time.Second,
	}

	input := os.Getenv(RATE_LIMIT_MAX_REQUESTS)
	if input == "" {
		logger.Warn(fmt.Sprintf(MissingEnvVarWarning, RATE_LIMIT_MAX_REQUESTS))
	}

	maxRequests, err := strconv.Atoi(input)
	if err == nil && maxRequests > 0 {
		config.MaxRequests = uint(maxRequests)

	}
	if err != nil {
		message := fmt.Sprintf(InvalidEnvVarErr, RATE_LIMIT_MAX_REQUESTS)
		logger.Error(message, zap.String(INPUT, input), zap.Error(err))
	}

	if maxRequests <= 0 {
		err := fmt.Errorf(NegativeValueErr, input)
		message := fmt.Sprintf(InvalidEnvVarErr, RATE_LIMIT_MAX_REQUESTS)
		logger.Error(message, zap.Error(err))
	}

	return config
}
