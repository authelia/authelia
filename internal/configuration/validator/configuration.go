package validator

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/configuration/schema"
)

var defaultPort = 8080
var defaultLogLevel = "info"

// ValidateConfiguration and adapt the configuration read from file.
//nolint:gocyclo // This function is likely to always have lots of if/else statements, as long as we keep the flow clean it should be understandable.
func ValidateConfiguration(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.Host == "" {
		configuration.Host = "0.0.0.0"
	}

	if configuration.Port == 0 {
		configuration.Port = defaultPort
	}

	if configuration.TLSKey != "" && configuration.TLSCert == "" {
		validator.Push(fmt.Errorf("No TLS certificate provided, please check the \"tls_cert\" which has been configured"))
	} else if configuration.TLSKey == "" && configuration.TLSCert != "" {
		validator.Push(fmt.Errorf("No TLS key provided, please check the \"tls_key\" which has been configured"))
	}

	if configuration.LogLevel == "" {
		configuration.LogLevel = defaultLogLevel
	}

	if configuration.JWTSecret == "" {
		validator.Push(fmt.Errorf("Provide a JWT secret using \"jwt_secret\" key"))
	}

	if configuration.DefaultRedirectionURL != "" {
		_, err := url.ParseRequestURI(configuration.DefaultRedirectionURL)
		if err != nil {
			validator.Push(fmt.Errorf("Unable to parse default redirection url"))
		}
	}

	if configuration.TOTP == nil {
		configuration.TOTP = &schema.DefaultTOTPConfiguration
	}

	ValidateTOTP(configuration.TOTP, validator)

	ValidateAuthenticationBackend(&configuration.AuthenticationBackend, validator)

	if configuration.AccessControl.DefaultPolicy == "" {
		configuration.AccessControl.DefaultPolicy = "deny"
	}

	ValidateSession(&configuration.Session, validator)

	if configuration.Regulation == nil {
		configuration.Regulation = &schema.DefaultRegulationConfiguration
	}

	ValidateRegulation(configuration.Regulation, validator)

	ValidateServer(&configuration.Server, validator)

	ValidateStorage(configuration.Storage, validator)

	if configuration.Notifier == nil {
		validator.Push(fmt.Errorf("A notifier configuration must be provided"))
	} else {
		ValidateNotifier(configuration.Notifier, validator)
	}
}
