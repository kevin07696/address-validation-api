package adapters

import (
	"address-validator/config"
	"address-validator/ports"
	"context"
	"fmt"
	"strings"

	// Using standard log for simplicity, replace with zap if needed
	"go.uber.org/zap" // Assuming you use zap for logging
	addressvalidation "google.golang.org/api/addressvalidation/v1"
	"google.golang.org/api/option"
)

type GoogleAddressValidationAdapter struct {
	client *addressvalidation.Service
	logger *zap.Logger      // Using zap as in your example
	config config.MapConfig // Keeping your config type for consistency
}

// NewGoogleAddressValidationAdapter creates a new Google Address Validation adapter
func NewGoogleAddressValidationAdapter(config config.MapConfig, logger *zap.Logger) (*GoogleAddressValidationAdapter, error) {
	ctx := context.Background()
	client, err := addressvalidation.NewService(ctx, option.WithAPIKey(config.GoogleMapsAPIKey)) // Using API Key as in your example
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Address Validation service: %w", err)
	}

	return &GoogleAddressValidationAdapter{
		client: client,
		logger: logger,
		config: config,
	}, nil
}

// ValidateAddress validates an address using Google Address Validation API
func (gava *GoogleAddressValidationAdapter) ValidateAddress(ctx context.Context, address string) (ports.AddressValidationResult, error) {
	// Create result object
	result := ports.AddressValidationResult{
		IsValid: false,
	}

	// Call Google Address Validation API
	req := &addressvalidation.GoogleMapsAddressvalidationV1ValidateAddressRequest{
		Address: &addressvalidation.GoogleTypePostalAddress{
			AddressLines: []string{address},
			RegionCode:   gava.config.Country,
			Locality:     gava.config.Locality,
		},
	}

	gava.logger.Debug("calling Google Address Validation API", zap.Any("request", req))
	resp, err := gava.client.V1.ValidateAddress(req).Do()
	if err != nil {
		gava.logger.Error("address validation error", zap.Error(err))
		result.Error = "Failed to validate address: " + err.Error()
		return result, fmt.Errorf("address validation error: %w", err)
	}

	// Check the validation results
	if resp != nil && resp.Result != nil && resp.Result.Verdict != nil {
		verdict := resp.Result.Verdict

		// Consider an address valid if it's at least Premises level and complete
		if verdict.ValidationGranularity >= "PREMISE" && verdict.AddressComplete {
			result.IsValid = true
		}

		if resp.Result.Address != nil && resp.Result.Address.FormattedAddress != "" {
			result.FormattedAddress = resp.Result.Address.FormattedAddress
		}

		if resp.Result.Geocode != nil && resp.Result.Geocode.Location != nil {
			result.Latitude = resp.Result.Geocode.Location.Latitude
			result.Longitude = resp.Result.Geocode.Location.Longitude
		}

		// You might want to add more detailed error information based on the verdict
		if !result.IsValid {
			var errors []string
			if verdict.InputGranularity == "OTHER" {
				errors = append(errors, "Input address was not recognized.")
			}
			if !verdict.AddressComplete {
				errors = append(errors, "Address is incomplete.")
			}
			// Add more checks based on your requirements
			if len(errors) > 0 {
				result.Error = strings.Join(errors, " ")
			} else if result.Error == "" {
				result.Error = "Address validation failed based on granularity."
			}
		}
	} else {
		gava.logger.Warn("no validation result found for address")
		result.Error = "No validation result found."
		return result, fmt.Errorf("no validation result found")
	}

	return result, nil
}
