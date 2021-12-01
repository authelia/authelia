package validator

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
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

	if configuration.TOTP.Algorithm == "" {
		configuration.TOTP.Algorithm = schema.DefaultTOTPConfiguration.Algorithm
	} else {
		configuration.TOTP.Algorithm = strings.ToUpper(configuration.TOTP.Algorithm)

		if !utils.IsStringInSlice(configuration.TOTP.Algorithm, schema.TOTPPossibleAlgorithms) {
			validator.Push(fmt.Errorf(errFmtTOTPInvalidAlgorithm, configuration.TOTP.Algorithm, strings.Join(schema.TOTPPossibleAlgorithms, ", ")))
		}
	}

	if configuration.TOTP.Period == 0 {
		configuration.TOTP.Period = schema.DefaultTOTPConfiguration.Period
	} else if configuration.TOTP.Period < 15 {
		validator.Push(fmt.Errorf(errFmtTOTPInvalidPeriod, configuration.TOTP.Period))
	}

	if configuration.TOTP.Digits == 0 {
		configuration.TOTP.Digits = schema.DefaultTOTPConfiguration.Digits
	} else if configuration.TOTP.Digits != 6 && configuration.TOTP.Digits != 8 {
		validator.Push(fmt.Errorf(errFmtTOTPInvalidDigits, configuration.TOTP.Digits))
	}

	if configuration.TOTP.Skew == nil {
		configuration.TOTP.Skew = schema.DefaultTOTPConfiguration.Skew
	}
}
