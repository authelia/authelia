package validator

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateWebauthn validates and update Webauthn configuration.
func ValidateWebauthn(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Webauthn.DisplayName == "" {
		config.Webauthn.DisplayName = schema.DefaultWebauthnConfiguration.DisplayName
	}

	if config.Webauthn.Timeout == "" {
		config.Webauthn.Timeout = schema.DefaultWebauthnConfiguration.Timeout
	} else if _, err := utils.ParseDurationString(config.Webauthn.Timeout); err != nil {
		validator.Push(fmt.Errorf(errFmtWebauthnParseTimeout, err))
	}

	switch {
	case config.Webauthn.ConveyancePreference == "":
		config.Webauthn.ConveyancePreference = schema.DefaultWebauthnConfiguration.ConveyancePreference
	case !utils.IsStringInSlice(string(config.Webauthn.ConveyancePreference), validWebauthnConveyancePreferences):
		validator.Push(fmt.Errorf(errFmtWebauthnConveyancePreference, strings.Join(validWebauthnConveyancePreferences, "', '"), config.Webauthn.ConveyancePreference))
	}

	switch {
	case config.Webauthn.UserVerification == "":
		config.Webauthn.UserVerification = schema.DefaultWebauthnConfiguration.UserVerification
	case !utils.IsStringInSlice(string(config.Webauthn.UserVerification), validWebauthnUserVerificationRequirement):
		validator.Push(fmt.Errorf(errFmtWebauthnUserVerification, config.Webauthn.UserVerification))
	}
}
