package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateSession validates and update session configuration.
func ValidateSession(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if config.Name == "" {
		config.Name = schema.DefaultSessionConfiguration.Name
	}

	if config.Redis != nil {
		if config.Redis.HighAvailability != nil {
			validateRedisSentinel(config, validator)
		} else {
			validateRedis(config, validator)
		}
	}

	validateSession(config, validator)
}

func validateSession(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if config.Expiration <= 0 {
		config.Expiration = schema.DefaultSessionConfiguration.Expiration // 1 hour.
	}

	if config.Inactivity <= 0 {
		config.Inactivity = schema.DefaultSessionConfiguration.Inactivity // 5 min.
	}

	if config.RememberMeDuration <= 0 && config.RememberMeDuration != schema.RememberMeDisabled {
		config.RememberMeDuration = schema.DefaultSessionConfiguration.RememberMeDuration // 1 month.
	}

	if config.Domain == "" {
		validator.Push(fmt.Errorf(errFmtSessionOptionRequired, "domain"))
	}

	if strings.HasPrefix(config.Domain, "*.") {
		validator.Push(fmt.Errorf(errFmtSessionDomainMustBeRoot, config.Domain))
	}

	if config.SameSite == "" {
		config.SameSite = schema.DefaultSessionConfiguration.SameSite
	} else if !utils.IsStringInSlice(config.SameSite, validSessionSameSiteValues) {
		validator.Push(fmt.Errorf(errFmtSessionSameSite, strings.Join(validSessionSameSiteValues, "', '"), config.SameSite))
	}
}

func validateRedisCommon(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if config.Secret == "" {
		validator.Push(fmt.Errorf(errFmtSessionSecretRequired, "redis"))
	}

	if config.Redis.TLS != nil {
		configDefaultTLS := &schema.TLSConfig{
			ServerName:     config.Redis.Host,
			MinimumVersion: schema.DefaultRedisConfiguration.TLS.MinimumVersion,
			MaximumVersion: schema.DefaultRedisConfiguration.TLS.MaximumVersion,
		}

		if err := ValidateTLSConfig(config.Redis.TLS, configDefaultTLS); err != nil {
			validator.Push(fmt.Errorf(errFmtSessionRedisTLSConfigInvalid, err))
		}
	}
}

func validateRedis(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if config.Redis.Host == "" {
		validator.Push(fmt.Errorf(errFmtSessionRedisHostRequired))
	}

	validateRedisCommon(config, validator)

	if !strings.HasPrefix(config.Redis.Host, "/") && config.Redis.Port == 0 {
		validator.Push(errors.New("A redis port different than 0 must be provided"))
	} else if config.Redis.Port < 0 || config.Redis.Port > 65535 {
		validator.Push(fmt.Errorf(errFmtSessionRedisPortRange, config.Redis.Port))
	}

	if config.Redis.MaximumActiveConnections <= 0 {
		config.Redis.MaximumActiveConnections = 8
	}
}

func validateRedisSentinel(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if config.Redis.HighAvailability.SentinelName == "" {
		validator.Push(fmt.Errorf(errFmtSessionRedisSentinelMissingName))
	}

	if config.Redis.Port == 0 {
		config.Redis.Port = 26379
	} else if config.Redis.Port < 0 || config.Redis.Port > 65535 {
		validator.Push(fmt.Errorf(errFmtSessionRedisPortRange, config.Redis.Port))
	}

	if config.Redis.Host == "" && len(config.Redis.HighAvailability.Nodes) == 0 {
		validator.Push(fmt.Errorf(errFmtSessionRedisHostOrNodesRequired))
	}

	validateRedisCommon(config, validator)

	hostMissing := false

	for i, node := range config.Redis.HighAvailability.Nodes {
		if node.Host == "" {
			hostMissing = true
		}

		if node.Port == 0 {
			config.Redis.HighAvailability.Nodes[i].Port = 26379
		}
	}

	if hostMissing {
		validator.Push(fmt.Errorf(errFmtSessionRedisSentinelNodeHostMissing))
	}
}
