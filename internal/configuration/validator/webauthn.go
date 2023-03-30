package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateWebauthn validates and update Webauthn configuration.
func ValidateWebauthn(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Webauthn.DisplayName == "" {
		config.Webauthn.DisplayName = schema.DefaultWebauthnConfiguration.DisplayName
	}

	if config.Webauthn.Timeout <= 0 {
		config.Webauthn.Timeout = schema.DefaultWebauthnConfiguration.Timeout
	}

	switch {
	case config.Webauthn.ConveyancePreference == "":
		config.Webauthn.ConveyancePreference = schema.DefaultWebauthnConfiguration.ConveyancePreference
	case !utils.IsStringInSlice(string(config.Webauthn.ConveyancePreference), validWebauthnConveyancePreferences):
		validator.Push(fmt.Errorf(errFmtWebauthnConveyancePreference, strJoinOr(validWebauthnConveyancePreferences), config.Webauthn.ConveyancePreference))
	}

	switch {
	case config.Webauthn.UserVerification == "":
		config.Webauthn.UserVerification = schema.DefaultWebauthnConfiguration.UserVerification
	case !utils.IsStringInSlice(string(config.Webauthn.UserVerification), validWebauthnUserVerificationRequirement):
		validator.Push(fmt.Errorf(errFmtWebauthnUserVerification, strJoinOr(validWebauthnConveyancePreferences), config.Webauthn.UserVerification))
	}
}
