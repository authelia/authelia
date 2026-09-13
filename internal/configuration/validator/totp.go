// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateTOTP validates and updates TOTP configuration.
func ValidateTOTP(config *schema.Configuration, validator *schema.StructValidator) {
	if config.TOTP.Disable {
		return
	}

	if config.TOTP.Issuer == "" {
		config.TOTP.Issuer = schema.DefaultTOTPConfiguration.Issuer
	}

	validateTOTPValueSetAlgorithm(config, validator)
	validateTOTPValueSetPeriod(config, validator)
	validateTOTPValueSetDigits(config, validator)

	if config.TOTP.Skew == nil {
		config.TOTP.Skew = schema.DefaultTOTPConfiguration.Skew
	}

	if config.TOTP.SecretSize == 0 {
		config.TOTP.SecretSize = schema.DefaultTOTPConfiguration.SecretSize
	} else if config.TOTP.SecretSize < schema.TOTPSecretSizeMinimum {
		validator.Push(fmt.Errorf(errFmtTOTPInvalidSecretSize, schema.TOTPSecretSizeMinimum, config.TOTP.SecretSize))
	}

	validateTOTPApps(config, validator)
}

func validateTOTPApps(config *schema.Configuration, validator *schema.StructValidator) {
	stores := []struct {
		name     string
		value    *schema.TOTPAppsStore
		fallback url.URL
	}{
		{"apple_store", &config.TOTP.Apps.AppleStore, schema.DefaultTOTPApps.AppleStore.URL},
		{"google_play", &config.TOTP.Apps.GooglePlay, schema.DefaultTOTPApps.GooglePlay.URL},
	}

	for _, store := range stores {
		if store.value.Disable {
			continue
		}

		if store.value.URL.String() == "" {
			store.value.URL = store.fallback

			continue
		}

		if store.value.URL.Scheme != schemeHTTPS {
			validator.Push(fmt.Errorf(errFmtTOTPAppsInvalidScheme, store.name, store.value.URL.String(), store.value.URL.Scheme))
		}
	}
}

func validateTOTPValueSetAlgorithm(config *schema.Configuration, validator *schema.StructValidator) {
	if config.TOTP.DefaultAlgorithm == "" {
		config.TOTP.DefaultAlgorithm = schema.DefaultTOTPConfiguration.DefaultAlgorithm
	} else {
		config.TOTP.DefaultAlgorithm = strings.ToUpper(config.TOTP.DefaultAlgorithm)

		if !utils.IsStringInSlice(config.TOTP.DefaultAlgorithm, schema.TOTPPossibleAlgorithms) {
			validator.Push(fmt.Errorf(errFmtTOTPInvalidAlgorithm, utils.StringJoinOr(schema.TOTPPossibleAlgorithms), config.TOTP.DefaultAlgorithm))
		}
	}

	for i, algorithm := range config.TOTP.AllowedAlgorithms {
		config.TOTP.AllowedAlgorithms[i] = strings.ToUpper(algorithm)

		if !utils.IsStringInSlice(config.TOTP.AllowedAlgorithms[i], schema.TOTPPossibleAlgorithms) {
			validator.Push(fmt.Errorf(errFmtTOTPInvalidAllowedAlgorithm, utils.StringJoinOr(schema.TOTPPossibleAlgorithms), config.TOTP.AllowedAlgorithms[i]))
		}
	}

	if !utils.IsStringInSlice(config.TOTP.DefaultAlgorithm, config.TOTP.AllowedAlgorithms) {
		config.TOTP.AllowedAlgorithms = append(config.TOTP.AllowedAlgorithms, config.TOTP.DefaultAlgorithm)
	}
}

func validateTOTPValueSetPeriod(config *schema.Configuration, validator *schema.StructValidator) {
	if config.TOTP.DefaultPeriod == 0 {
		config.TOTP.DefaultPeriod = schema.DefaultTOTPConfiguration.DefaultPeriod
	} else if config.TOTP.DefaultPeriod < 15 {
		validator.Push(fmt.Errorf(errFmtTOTPInvalidPeriod, config.TOTP.DefaultPeriod))
	}

	var hasDefaultPeriod bool

	for _, period := range config.TOTP.AllowedPeriods {
		if period < 15 {
			validator.Push(fmt.Errorf(errFmtTOTPInvalidAllowedPeriod, period))
		}

		if period == config.TOTP.DefaultPeriod {
			hasDefaultPeriod = true
		}
	}

	if !hasDefaultPeriod {
		config.TOTP.AllowedPeriods = append(config.TOTP.AllowedPeriods, config.TOTP.DefaultPeriod)
	}
}

func validateTOTPValueSetDigits(config *schema.Configuration, validator *schema.StructValidator) {
	if config.TOTP.DefaultDigits == 0 {
		config.TOTP.DefaultDigits = schema.DefaultTOTPConfiguration.DefaultDigits
	} else if config.TOTP.DefaultDigits != 6 && config.TOTP.DefaultDigits != 8 {
		validator.Push(fmt.Errorf(errFmtTOTPInvalidDigits, config.TOTP.DefaultDigits))
	}

	var hasDefaultDigits bool

	for _, digits := range config.TOTP.AllowedDigits {
		if digits != 6 && digits != 8 {
			validator.Push(fmt.Errorf(errFmtTOTPInvalidAllowedDigit, digits))
		}

		if digits == config.TOTP.DefaultDigits {
			hasDefaultDigits = true
		}
	}

	if !hasDefaultDigits {
		config.TOTP.AllowedDigits = append(config.TOTP.AllowedDigits, config.TOTP.DefaultDigits)
	}
}
