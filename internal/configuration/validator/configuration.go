package validator

import (
	"fmt"

	"github.com/clems4ever/authelia/internal/configuration/schema"
)

var defaultPort = 8080
var defaultLogsLevel = "info"

// Validate and adapt the configuration read from file.
func Validate(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.Host == "" {
		configuration.Host = "0.0.0.0"
	}

	if configuration.Port == 0 {
		configuration.Port = defaultPort
	}

	if configuration.LogsLevel == "" {
		configuration.LogsLevel = defaultLogsLevel
	}

	if configuration.JWTSecret == "" {
		validator.Push(fmt.Errorf("Provide a JWT secret using `jwt_secret` key"))
	}

	ValidateAuthenticationBackend(&configuration.AuthenticationBackend, validator)
	ValidateSession(&configuration.Session, validator)

	if configuration.TOTP == nil {
		configuration.TOTP = &schema.TOTPConfiguration{}
		ValidateTOTP(configuration.TOTP, validator)
	}

	ValidateSQLStorage(configuration.Storage, validator)
}
