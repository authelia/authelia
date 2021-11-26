package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateTOTP validates and update TOTP configuration.
func ValidateTOTP(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.TOTP == nil {
		configuration.TOTP = &schema.DefaultTOTPConfiguration

		return
	}

	if configuration.TOTP.Issuer == "" {
		configuration.TOTP.Issuer = schema.DefaultTOTPConfiguration.Issuer
	}

	if configuration.TOTP.Period == 0 {
		configuration.TOTP.Period = schema.DefaultTOTPConfiguration.Period
	} else if configuration.TOTP.Period < 0 {
		validator.Push(fmt.Errorf("TOTP Period must be 1 or more"))
	}

	if configuration.TOTP.Skew == nil {
		configuration.TOTP.Skew = schema.DefaultTOTPConfiguration.Skew
	} else if *configuration.TOTP.Skew < 0 {
		validator.Push(fmt.Errorf("TOTP Skew must be 0 or more"))
	}
}
