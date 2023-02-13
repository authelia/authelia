package totp

import (
	"github.com/authelia/authelia/v4/internal/model"
)

// Provider for TOTP functionality.
type Provider interface {
	Generate(username string) (config *model.TOTPConfiguration, err error)
	GenerateCustom(username string, algorithm, secret string, digits, period, secretSize uint) (config *model.TOTPConfiguration, err error)
	Validate(token string, config *model.TOTPConfiguration) (valid bool, err error)
	Options() model.TOTPOptions
}
