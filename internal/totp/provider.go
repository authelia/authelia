package totp

import (
	"github.com/authelia/authelia/v4/internal/models"
)

// Provider for TOTP functionality.
type Provider interface {
	Generate(username string) (config *models.TOTPConfiguration, err error)
	GenerateCustom(username string, algorithm string, digits, period, secretSize uint) (config *models.TOTPConfiguration, err error)
	Validate(token string, config *models.TOTPConfiguration) (valid bool, err error)
}
