package validator

import (
	"fmt"
	"github.com/authelia/authelia/internal/configuration/schema"
)

const defaultTOTPIssuer = "Authelia"
const DefaultTOTPPeriod = 30
const DefaultTOTPSkew = 1

// ValidateTOTP validates and update TOTP configuration.
func ValidateTOTP(configuration *schema.TOTPConfiguration, validator *schema.StructValidator) {
	if configuration.Issuer == "" {
		configuration.Issuer = defaultTOTPIssuer
	}
	if configuration.Period == 0 {
		configuration.Period = DefaultTOTPPeriod
	} else if configuration.Period < 0 {
		validator.Push(fmt.Errorf("TOTP Period must be 1 or more"))
	}

	if configuration.Skew == nil {
		var skew = DefaultTOTPSkew
		configuration.Skew = &skew
	} else if *configuration.Skew < 0 {
		validator.Push(fmt.Errorf("TOTP Skew must be 0 or more"))
	}
}
