package validator

import (
	"fmt"
	"strings"

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

	if configuration.Algorithm == "" {
		configuration.Algorithm = schema.SHA1
	} else {
		configuration.Algorithm = strings.ToLower(configuration.Algorithm)
		if configuration.Algorithm != schema.MD5 && configuration.Algorithm != schema.SHA1 && configuration.Algorithm != schema.SHA256 && configuration.Algorithm != schema.SHA512 {
			validator.Push(fmt.Errorf("TOTP Algorithm must be one of %s, %s, %s, or %s", schema.MD5, schema.SHA1, schema.SHA256, schema.SHA512))
		}
	}
}
