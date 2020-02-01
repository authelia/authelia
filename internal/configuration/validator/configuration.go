package validator

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/configuration/schema"
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

	if configuration.DefaultRedirectionURL != "" {
		_, err := url.ParseRequestURI(configuration.DefaultRedirectionURL)
		if err != nil {
			validator.Push(fmt.Errorf("Unable to parse default redirection url"))
		}
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

	if configuration.Notifier == nil {
		validator.Push(fmt.Errorf("A notifier configuration must be provided"))
	} else {
		ValidateNotifier(configuration.Notifier, validator)
	}

	ValidateSQLStorage(configuration.Storage, validator)
}
