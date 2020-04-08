package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// ValidateTOTP validates and update TOTP configuration.
func ValidateTOTP(configuration *schema.TOTPConfiguration, validator *schema.StructValidator) {
	if configuration.Issuer == "" {
		configuration.Issuer = schema.DefaultTOTPConfiguration.Issuer
	}
	if configuration.Period == 0 {
		configuration.Period = schema.DefaultTOTPConfiguration.Period
	} else if configuration.Period < 0 {
		validator.Push(fmt.Errorf("TOTP Period must be 1 or more"))
	}

	if configuration.Skew == nil {
		configuration.Skew = schema.DefaultTOTPConfiguration.Skew
	} else if *configuration.Skew < 0 {
		validator.Push(fmt.Errorf("TOTP Skew must be 0 or more"))
	}
}
