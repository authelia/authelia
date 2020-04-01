package validator

import (
	"errors"
	"github.com/authelia/authelia/internal/utils"

	"github.com/authelia/authelia/internal/configuration/schema"
)

var validRememberMeDurationUnits = []string{"y", "m", "w", "d", "h"}

// ValidateSession validates and update session configuration.
func ValidateSession(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Name == "" {
		configuration.Name = schema.DefaultSessionConfiguration.Name
	}

	if configuration.Redis != nil && configuration.Secret == "" {
		validator.Push(errors.New("Set secret of the session object"))
	}

	if configuration.Expiration == 0 {
		configuration.Expiration = schema.DefaultSessionConfiguration.Expiration // 1 hour
	} else if configuration.Expiration < 1 {
		validator.Push(errors.New("Set expiration of the session above 0"))
	}

	if configuration.Inactivity < 0 {
		validator.Push(errors.New("Set inactivity of the session above 0"))
	}

	if configuration.RememberMe == nil {
		configuration.RememberMe = &schema.DefaultSessionRememberMeConfiguration
	} else if configuration.RememberMe.Duration < 0 {
		validator.Push(errors.New("Set remember me duration of the session 0 or above"))
	} else if !utils.IsStringInSlice(configuration.RememberMe.DurationUnit, validRememberMeDurationUnits) {
		validator.Push(errors.New("Set rememebr me duration unit to one of y, m, w, d, h"))
	}

	if configuration.Domain == "" {
		validator.Push(errors.New("Set domain of the session object"))
	}
}
