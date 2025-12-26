package totp

import (
	"encoding/base32"
	"fmt"

	"github.com/authelia/otp"
	"github.com/authelia/otp/totp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// NewTimeBasedProvider creates a new totp.TimeBased which implements the totp.Provider.
func NewTimeBasedProvider(config schema.TOTP) (provider *TimeBased) {
	provider = &TimeBased{
		opts:      NewTOTPOptionsFromSchema(config),
		issuer:    config.Issuer,
		algorithm: config.DefaultAlgorithm,
		digits:    uint32(config.DefaultDigits),
		period:    uint(config.DefaultPeriod),
		size:      uint(config.SecretSize),
	}

	if config.Skew != nil && *config.Skew >= 0 {
		provider.skew = uint(*config.Skew)
	} else {
		provider.skew = 1
	}

	return provider
}

func NewTOTPOptionsFromSchema(config schema.TOTP) *model.TOTPOptions {
	return &model.TOTPOptions{
		Algorithm:  config.DefaultAlgorithm,
		Algorithms: config.AllowedAlgorithms,
		Period:     config.DefaultPeriod,
		Periods:    config.AllowedPeriods,
		Length:     config.DefaultDigits,
		Lengths:    config.AllowedDigits,
	}
}

// TimeBased totp.Provider for production use.
type TimeBased struct {
	opts *model.TOTPOptions

	issuer    string
	algorithm string
	digits    uint32
	period    uint
	skew      uint
	size      uint
}

// GenerateCustom generates a TOTP with custom options.
func (p TimeBased) GenerateCustom(ctx Context, username, algorithm, secret string, digits uint32, period, secretSize uint) (config *model.TOTPConfiguration, err error) {
	var key *otp.Key

	var secretData []byte

	if secret != "" {
		if secretData, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret); err != nil {
			return nil, fmt.Errorf("totp generate failed: error decoding base32 string: %w", err)
		}
	}

	if secretSize == 0 {
		secretSize = p.size
	}

	opts := totp.GenerateOpts{
		Issuer:      p.issuer,
		AccountName: username,
		Period:      period,
		Secret:      secretData,
		SecretSize:  secretSize,
		Digits:      otp.Digits(digits),
		Algorithm:   otpStringToAlgo(algorithm),
		Rand:        ctx.GetRandom(),
	}

	if key, err = totp.Generate(opts); err != nil {
		return nil, fmt.Errorf("error generating totp: %w", err)
	}

	config = &model.TOTPConfiguration{
		CreatedAt: ctx.GetClock().Now(),
		Username:  username,
		Issuer:    p.issuer,
		Algorithm: algorithm,
		Digits:    digits,
		Secret:    []byte(key.Secret()),
		Period:    period,
	}

	return config, nil
}

// Generate generates a TOTP with default options.
func (p TimeBased) Generate(ctx Context, username string) (config *model.TOTPConfiguration, err error) {
	return p.GenerateCustom(ctx, username, p.algorithm, "", p.digits, p.period, p.size)
}

// Validate the token against the given configuration.
func (p TimeBased) Validate(ctx Context, token string, config *model.TOTPConfiguration) (valid bool, step uint64, err error) {
	opts := totp.ValidateOpts{
		Period:    config.Period,
		Skew:      p.skew,
		Digits:    otp.Digits(config.Digits),
		Algorithm: otpStringToAlgo(config.Algorithm),
	}

	return totp.ValidateCustomStep(token, string(config.Secret), ctx.GetClock().Now().UTC(), opts)
}

// Options returns the configured options for this provider.
func (p TimeBased) Options() model.TOTPOptions {
	return *p.opts
}

var (
	_ Provider = (*TimeBased)(nil)
)
