// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateTOTP validates and update TOTP configuration.
func ValidateTOTP(config *schema.Configuration, validator *schema.StructValidator) {
	if config.TOTP.Disable {
		return
	}

	if config.TOTP.Issuer == "" {
		config.TOTP.Issuer = schema.DefaultTOTPConfiguration.Issuer
	}

	if config.TOTP.Algorithm == "" {
		config.TOTP.Algorithm = schema.DefaultTOTPConfiguration.Algorithm
	} else {
		config.TOTP.Algorithm = strings.ToUpper(config.TOTP.Algorithm)

		if !utils.IsStringInSlice(config.TOTP.Algorithm, schema.TOTPPossibleAlgorithms) {
			validator.Push(fmt.Errorf(errFmtTOTPInvalidAlgorithm, strings.Join(schema.TOTPPossibleAlgorithms, "', '"), config.TOTP.Algorithm))
		}
	}

	if config.TOTP.Period == 0 {
		config.TOTP.Period = schema.DefaultTOTPConfiguration.Period
	} else if config.TOTP.Period < 15 {
		validator.Push(fmt.Errorf(errFmtTOTPInvalidPeriod, config.TOTP.Period))
	}

	if config.TOTP.Digits == 0 {
		config.TOTP.Digits = schema.DefaultTOTPConfiguration.Digits
	} else if config.TOTP.Digits != 6 && config.TOTP.Digits != 8 {
		validator.Push(fmt.Errorf(errFmtTOTPInvalidDigits, config.TOTP.Digits))
	}

	if config.TOTP.Skew == nil {
		config.TOTP.Skew = schema.DefaultTOTPConfiguration.Skew
	}

	if config.TOTP.SecretSize == 0 {
		config.TOTP.SecretSize = schema.DefaultTOTPConfiguration.SecretSize
	} else if config.TOTP.SecretSize < schema.TOTPSecretSizeMinimum {
		validator.Push(fmt.Errorf(errFmtTOTPInvalidSecretSize, schema.TOTPSecretSizeMinimum, config.TOTP.SecretSize))
	}
}
