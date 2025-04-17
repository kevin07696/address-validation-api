package ports

import (
	"context"
)

// AddressValidationResult represents the result of address validation
type AddressValidationResult struct {
	IsValid          bool    `json:"isValid"`
	FormattedAddress string  `json:"formattedAddress"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	InRange          bool    `json:"inRange"`
	Error            string  `json:"error"`
}

const (
	DISTANCE_KILOMETER = "km"
	DISTANCE_MILES     = "mi"
)

// AddressValidator defines the interface for address validation
type AddressValidator interface {
	ValidateAddress(ctx context.Context, address string) (AddressValidationResult, error)
}
