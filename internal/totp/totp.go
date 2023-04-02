// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package totp

import (
	"encoding/base32"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// NewTimeBasedProvider creates a new totp.TimeBased which implements the totp.Provider.
func NewTimeBasedProvider(config schema.TOTPConfiguration) (provider *TimeBased) {
	provider = &TimeBased{
		config: &config,
	}

	if config.Skew != nil {
		provider.skew = *config.Skew
	} else {
		provider.skew = 1
	}

	return provider
}

// TimeBased totp.Provider for production use.
type TimeBased struct {
	config *schema.TOTPConfiguration
	skew   uint
}

// GenerateCustom generates a TOTP with custom options.
func (p TimeBased) GenerateCustom(username, algorithm, secret string, digits, period, secretSize uint) (config *model.TOTPConfiguration, err error) {
	var key *otp.Key

	var secretData []byte

	if secret != "" {
		if secretData, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret); err != nil {
			return nil, fmt.Errorf("totp generate failed: error decoding base32 string: %w", err)
		}
	}

	opts := totp.GenerateOpts{
		Issuer:      p.config.Issuer,
		AccountName: username,
		Period:      period,
		Secret:      secretData,
		SecretSize:  secretSize,
		Digits:      otp.Digits(digits),
		Algorithm:   otpStringToAlgo(algorithm),
	}

	if key, err = totp.Generate(opts); err != nil {
		return nil, err
	}

	config = &model.TOTPConfiguration{
		CreatedAt: time.Now(),
		Username:  username,
		Issuer:    p.config.Issuer,
		Algorithm: algorithm,
		Digits:    digits,
		Secret:    []byte(key.Secret()),
		Period:    period,
	}

	return config, nil
}

// Generate generates a TOTP with default options.
func (p TimeBased) Generate(username string) (config *model.TOTPConfiguration, err error) {
	return p.GenerateCustom(username, p.config.Algorithm, "", p.config.Digits, p.config.Period, p.config.SecretSize)
}

// Validate the token against the given configuration.
func (p TimeBased) Validate(token string, config *model.TOTPConfiguration) (valid bool, err error) {
	opts := totp.ValidateOpts{
		Period:    config.Period,
		Skew:      p.skew,
		Digits:    otp.Digits(config.Digits),
		Algorithm: otpStringToAlgo(config.Algorithm),
	}

	return totp.ValidateCustom(token, string(config.Secret), time.Now().UTC(), opts)
}
