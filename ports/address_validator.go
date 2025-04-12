package ports

import (
	"context"
)

// AddressValidationResult represents the result of address validation
type AddressValidationResult struct {
	IsValid          bool
	FormattedAddress string
	Latitude         float64
	Longitude        float64
	InRange          bool
	Error            string
}

// AddressValidator defines the interface for address validation
type AddressValidator interface {
	ValidateAddress(ctx context.Context, address string) (AddressValidationResult, error)
}
