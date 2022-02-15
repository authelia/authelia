package validator

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateWebauthn validates and update Webauthn configuration.
func ValidateWebauthn(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.Webauthn.DisplayName == "" {
		configuration.Webauthn.DisplayName = schema.DefaultWebauthnConfiguration.DisplayName
	}

	if configuration.Webauthn.Timeout == 0 {
		configuration.Webauthn.Timeout = schema.DefaultWebauthnConfiguration.Timeout
	}

	if configuration.Webauthn.ConveyancePreference == "" {
		configuration.Webauthn.ConveyancePreference = schema.DefaultWebauthnConfiguration.ConveyancePreference
	}

	if configuration.Webauthn.UserVerification == "" {
		configuration.Webauthn.UserVerification = schema.DefaultWebauthnConfiguration.UserVerification
	}
}
