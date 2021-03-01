package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateSession validates and update session configuration.
func ValidateSession(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Name == "" {
		configuration.Name = schema.DefaultSessionConfiguration.Name
	}

	validateProviderSessionConfiguration(configuration, validator)

	if configuration.Expiration == "" {
		configuration.Expiration = schema.DefaultSessionConfiguration.Expiration // 1 hour
	} else if _, err := utils.ParseDurationString(configuration.Expiration); err != nil {
		validator.Push(fmt.Errorf("Error occurred parsing session expiration string: %s", err))
	}

	if configuration.Inactivity == "" {
		configuration.Inactivity = schema.DefaultSessionConfiguration.Inactivity // 5 min
	} else if _, err := utils.ParseDurationString(configuration.Inactivity); err != nil {
		validator.Push(fmt.Errorf("Error occurred parsing session inactivity string: %s", err))
	}

	if configuration.RememberMeDuration == "" {
		configuration.RememberMeDuration = schema.DefaultSessionConfiguration.RememberMeDuration // 1 month
	} else if _, err := utils.ParseDurationString(configuration.RememberMeDuration); err != nil {
		validator.Push(fmt.Errorf("Error occurred parsing session remember_me_duration string: %s", err))
	}

	if configuration.Domain == "" {
		validator.Push(errors.New("Set domain of the session object"))
	}

	if strings.Contains(configuration.Domain, "*") {
		validator.Push(errors.New("The domain of the session must be the root domain you're protecting instead of a wildcard domain"))
	}
}

func validateProviderSessionConfiguration(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	sessionProviderCounter := 0

	if configuration.Redis != nil {
		sessionProviderCounter++

		validateRedisSessionConfiguration(configuration, validator)
	}

	if configuration.Memcache != nil {
		sessionProviderCounter++

		validateMemcacheSessionConfiguration(configuration, validator)
	}

	if configuration.MySQL != nil {
		sessionProviderCounter++

		validateSQLSessionConfiguration(&configuration.MySQL.SQLSessionConfiguration, validator)
	}

	if configuration.PostgreSQL != nil {
		sessionProviderCounter++

		validateSQLSessionConfiguration(&configuration.PostgreSQL.SQLSessionConfiguration, validator)
	}

	if configuration.Local != nil {
		sessionProviderCounter++

		validateLocalSessionConfiguration(configuration, validator)
	}

	if sessionProviderCounter > 1 {
		validator.Push(errors.New("Not more than one provider must be provided"))
	}
}

func validateRedisSessionConfiguration(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Secret == "" {
		validator.Push(errors.New("Set secret of the session object"))
	}

	if !strings.HasPrefix(configuration.Redis.Host, "/") && configuration.Redis.Port == 0 {
		validator.Push(errors.New("A redis port different than 0 must be provided"))
	}
}

func validateMemcacheSessionConfiguration(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Secret == "" {
		validator.Push(errors.New("Set secret of the session object"))
	}

	for _, server := range configuration.Memcache {
		if server.Host == "" {
			validator.Push(errors.New("A memcache host must be provided"))
		}

		if server.Port == 0 {
			validator.Push(errors.New("A memcache port different than 0 must be provided"))
		}
	}
}

func validateSQLSessionConfiguration(configuration *schema.SQLSessionConfiguration, validator *schema.StructValidator) {
	if configuration.Password == "" || configuration.Username == "" {
		validator.Push(errors.New("Username and password must be provided"))
	}

	if configuration.Database == "" {
		validator.Push(errors.New("A database must be provided"))
	}
}

func validateLocalSessionConfiguration(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Local.Path == "" {
		validator.Push(errors.New("A file path must be provided with key 'path'"))
	}
}
