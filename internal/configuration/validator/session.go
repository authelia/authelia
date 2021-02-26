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

	if configuration.Redis != nil {
		if configuration.Redis.Sentinel != "" {
			validateRedisSentinel(configuration, validator)
		} else {
			validateRedis(configuration, validator)

			if len(configuration.Redis.Nodes) > 0 {
				validator.Push(errors.New("Session redis provider does not have the sentinel option specified but has nodes defined which is invalid"))
			}

			if configuration.Redis.SentinelPassword != "" {
				validator.Push(errors.New("Session redis provider does not have the sentinel option specified but has a password defined which is invalid"))
			}

			if configuration.Redis.RouteRandomly || configuration.Redis.RouteByLatency {
				validator.Push(errors.New("Session redis provider does not have the sentinel option specified but has routing options defined which is invalid"))
			}
		}
	}

	validateSession(configuration, validator)
}

func validateSession(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
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
	if configuration.Redis.Host == "" {
		validator.Push(fmt.Errorf(errFmtSessionRedisHostRequired, "redis"))
	}

	if configuration.Secret == "" {
		validator.Push(fmt.Errorf(errFmtSessionSecretRedisProvider, "redis"))
	}

	if !strings.HasPrefix(configuration.Redis.Host, "/") && configuration.Redis.Port == 0 {
		validator.Push(errors.New("A redis port different than 0 must be provided"))
	} else if configuration.Redis.Port < 0 || configuration.Redis.Port > 65535 {
		validator.Push(fmt.Errorf(errFmtSessionRedisPortRange, "redis"))
	}
}

func validateRedisSentinel(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Redis.Host == "" {
		validator.Push(fmt.Errorf(errFmtSessionRedisHostRequired, "redis sentinel"))
	}

	if configuration.Secret == "" {
		validator.Push(fmt.Errorf(errFmtSessionSecretRedisProvider, "redis sentinel"))
	}

	if configuration.Redis.Port == 0 {
		configuration.Redis.Port = 26379
	} else if configuration.Redis.Port < 0 || configuration.Redis.Port > 65535 {
		validator.Push(fmt.Errorf(errFmtSessionRedisPortRange, "redis sentinel"))
	}

	for i, node := range configuration.Redis.Nodes {
		if node.Host == "" {
			validator.Push(fmt.Errorf("The redis sentinel nodes require a host set but you have not set the host for one or more nodes"))
			break
		}

		if node.Port == 0 {
			configuration.Redis.Nodes[i].Port = 26379
		}
	}
}
