package totp

import (
	"github.com/authelia/authelia/v4/internal/model"
)

// Provider for TOTP functionality.
type Provider interface {
	Generate(ctx Context, username string) (config *model.TOTPConfiguration, err error)
	GenerateCustom(ctx Context, username string, algorithm, secret string, digits, period, secretSize uint) (config *model.TOTPConfiguration, err error)
	Validate(ctx Context, token string, config *model.TOTPConfiguration) (valid bool, step uint64, err error)
	Options() model.TOTPOptions
}
