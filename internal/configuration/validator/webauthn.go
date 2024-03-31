package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateWebAuthn validates and update WebAuthn configuration.
func ValidateWebAuthn(config *schema.Configuration, validator *schema.StructValidator) {
	if config.WebAuthn.DisplayName == "" {
		config.WebAuthn.DisplayName = schema.DefaultWebAuthnConfiguration.DisplayName
	}

	if config.WebAuthn.Timeout <= 0 {
		config.WebAuthn.Timeout = schema.DefaultWebAuthnConfiguration.Timeout
	}

	switch {
	case config.WebAuthn.ConveyancePreference == "":
		config.WebAuthn.ConveyancePreference = schema.DefaultWebAuthnConfiguration.ConveyancePreference
	case !utils.IsStringInSlice(string(config.WebAuthn.ConveyancePreference), validWebAuthnConveyancePreferences):
		validator.Push(fmt.Errorf(errFmtWebAuthnConveyancePreference, utils.StringJoinOr(validWebAuthnConveyancePreferences), config.WebAuthn.ConveyancePreference))
	}

	switch {
	case config.WebAuthn.UserVerification == "":
		config.WebAuthn.UserVerification = schema.DefaultWebAuthnConfiguration.UserVerification
	case !utils.IsStringInSlice(string(config.WebAuthn.UserVerification), validWebAuthnUserVerificationRequirement):
		validator.Push(fmt.Errorf(errFmtWebAuthnUserVerification, utils.StringJoinOr(validWebAuthnConveyancePreferences), config.WebAuthn.UserVerification))
	}
}
