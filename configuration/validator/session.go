package validator

import (
	"errors"

	"github.com/clems4ever/authelia/configuration/schema"
)

// ValidateSession validates and update session configuration.
func ValidateSession(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Name == "" {
		configuration.Name = schema.DefaultSessionConfiguration.Name
	}

	if configuration.Secret == "" {
		validator.Push(errors.New("Set secret of the session object"))
	}

	if configuration.Expiration == 0 {
		configuration.Expiration = schema.DefaultSessionConfiguration.Expiration // 1 hour
	}

	if configuration.Domain == "" {
		validator.Push(errors.New("Set domain of the session object"))
	}
}
