package validator

import (
	"errors"
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateSession validates and update session configuration.
func ValidateSession(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Name == "" {
		configuration.Name = schema.DefaultSessionConfiguration.Name
	}

	if configuration.Redis != nil && configuration.Secret == "" {
		validator.Push(errors.New("Set secret of the session object"))
	}

	if configuration.Expiration == "" {
		configuration.Expiration = schema.DefaultSessionConfiguration.Expiration // 1 hour
	} else if _, err := utils.ParseDurationString(configuration.Expiration); err != nil {
		validator.Push(errors.New(fmt.Sprintf("Error occurred parsing session expiration string: %s", err)))
	}

	if configuration.Inactivity == "" {
		configuration.Inactivity = schema.DefaultSessionConfiguration.Inactivity // 5 min
	} else if _, err := utils.ParseDurationString(configuration.Inactivity); err != nil {
		validator.Push(errors.New(fmt.Sprintf("Error occurred parsing session inactivity string: %s", err)))
	}

	if configuration.RememberMeDuration == "" {
		configuration.RememberMeDuration = schema.DefaultSessionConfiguration.RememberMeDuration // 1 month
	} else if _, err := utils.ParseDurationString(configuration.RememberMeDuration); err != nil {
		validator.Push(errors.New(fmt.Sprintf("Error occurred parsing session remember_me_duration string: %s", err)))
	}

	if configuration.Domain == "" {
		validator.Push(errors.New("Set domain of the session object"))
	}
}
