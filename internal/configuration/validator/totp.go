package validator

import (
	"github.com/authelia/authelia/internal/configuration/schema"
)

const defaultTOTPIssuer = "Authelia"

// ValidateTOTP validates and update TOTP configuration.
func ValidateTOTP(configuration *schema.TOTPConfiguration, validator *schema.StructValidator) {
	if configuration.Issuer == "" {
		configuration.Issuer = defaultTOTPIssuer
	}
}
