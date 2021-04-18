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
		if configuration.Redis.HighAvailability != nil {
			if configuration.Redis.HighAvailability.SentinelName != "" {
				validateRedisSentinel(configuration, validator)
			} else {
				validator.Push(fmt.Errorf("Session provider redis is configured for high availability but doesn't have a sentinel_name which is required"))
			}
		} else {
			validateRedis(configuration, validator)
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

	if configuration.SameSite == "" {
		configuration.SameSite = schema.DefaultSessionConfiguration.SameSite
	} else if configuration.SameSite != "none" && configuration.SameSite != "lax" && configuration.SameSite != "strict" {
		validator.Push(errors.New("session same_site is configured incorrectly, must be one of 'none', 'lax', or 'strict'"))
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

	if configuration.Redis.MaximumActiveConnections <= 0 {
		configuration.Redis.MaximumActiveConnections = 8
	}
}

func validateRedisSentinel(configuration *schema.SessionConfiguration, validator *schema.StructValidator) {
	if configuration.Redis.Port == 0 {
		configuration.Redis.Port = 26379
	} else if configuration.Redis.Port < 0 || configuration.Redis.Port > 65535 {
		validator.Push(fmt.Errorf(errFmtSessionRedisPortRange, "redis sentinel"))
	}

	validateHighAvailability(configuration, validator, "redis sentinel")
}

func validateHighAvailability(configuration *schema.SessionConfiguration, validator *schema.StructValidator, provider string) {
	if configuration.Redis.Host == "" && len(configuration.Redis.HighAvailability.Nodes) == 0 {
		validator.Push(fmt.Errorf(errFmtSessionRedisHostOrNodesRequired, provider))
	}

	if configuration.Secret == "" {
		validator.Push(fmt.Errorf(errFmtSessionSecretRedisProvider, provider))
	}

	for i, node := range configuration.Redis.HighAvailability.Nodes {
		if node.Host == "" {
			validator.Push(fmt.Errorf("The %s nodes require a host set but you have not set the host for one or more nodes", provider))
			break
		}

		if node.Port == 0 {
			if provider == "redis sentinel" {
				configuration.Redis.HighAvailability.Nodes[i].Port = 26379
			}
		}
	}
}
