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

	if configuration.Redis != nil && configuration.RedisSentinel != nil {
		validator.Push(errors.New("Must only specify only one session provider (redis or redis_sentinel)"))
	}

	if configuration.Redis != nil {
		validateRedis(configuration, validator)
	}

	if configuration.RedisSentinel != nil {
		validateRedisSentinel(configuration, validator)
	}

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

func validateRedis(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Secret == "" {
		validator.Push(errors.New("Set secret of the session object"))
	}

	if !strings.HasPrefix(configuration.Redis.Host, "/") && configuration.Redis.Port == 0 {
		validator.Push(errors.New("A redis port different than 0 must be provided"))
	}
}

func validateRedisSentinel(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Secret == "" {
		validator.Push(errors.New("Set secret of the session object"))
	}

	if configuration.RedisSentinel.Host == "" || configuration.RedisSentinel.Port == 0 {
		validator.Push(errors.New("The host and port must be specified when using the redis sentinel session provider"))
	}

	for _, node := range configuration.RedisSentinel.Nodes {
		if node.Host == "" {
			validator.Push(fmt.Errorf("The host must be specified for each node when using the redis sentinel session provider, the offending entry is %s:%d", node.Host, node.Port))
		}
	}
}
