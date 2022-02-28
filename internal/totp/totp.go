package totp

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/models"
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
func (p TimeBased) GenerateCustom(username, algorithm string, digits, period, secretSize uint) (config *models.TOTPConfiguration, err error) {
	var key *otp.Key

	opts := totp.GenerateOpts{
		Issuer:      p.config.Issuer,
		AccountName: username,
		Period:      period,
		SecretSize:  secretSize,
		Digits:      otp.Digits(digits),
		Algorithm:   otpStringToAlgo(algorithm),
	}

	if key, err = totp.Generate(opts); err != nil {
		return nil, err
	}

	config = &models.TOTPConfiguration{
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
func (p TimeBased) Generate(username string) (config *models.TOTPConfiguration, err error) {
	return p.GenerateCustom(username, p.config.Algorithm, p.config.Digits, p.config.Period, 32)
}

// Validate the token against the given configuration.
func (p TimeBased) Validate(token string, config *models.TOTPConfiguration) (valid bool, err error) {
	opts := totp.ValidateOpts{
		Period:    config.Period,
		Skew:      p.skew,
		Digits:    otp.Digits(config.Digits),
		Algorithm: otpStringToAlgo(config.Algorithm),
	}

	return totp.ValidateCustom(token, string(config.Secret), time.Now().UTC(), opts)
}
