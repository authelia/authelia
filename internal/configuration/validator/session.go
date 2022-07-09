package validator

import (
	"errors"
	"fmt"
	"sort"
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
	validateSessionDomains(config, validator)
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

	if config.Domain == "" && len(config.Domains) == 0 {
		validator.Push(fmt.Errorf(errFmtSessionOptionRequired, "domain"))
	}

	// Add default domain to protected domains list
	// Refactor: this can be refatored as `func (s * SessionConfiguration) ProtectedDomains() []string` in SessionConfiguration structure.
	if config.Domain != "" {
		config.ProtectedDomains = append(config.ProtectedDomains, config.Domain)
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

// validateSessionDomains validates domain list.
func validateSessionDomains(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	cookieDomainList := make([]string, 0, len(config.Domains))
	for index := range config.Domains {
		// ensure there's not duplicated domain_cookie.
		if sliceContainsString(cookieDomainList, config.Domains[index].CookieDomain) {
			validator.Push(fmt.Errorf(errFmtSessionDupplicatedDomainCookie, config.Domains[index].CookieDomain, index))
		}

		cookieDomainList = append(cookieDomainList, config.Domains[index].CookieDomain)
		sort.Strings(cookieDomainList)

		if len(config.Domains[index].Domains) == 0 {
			validator.Push(fmt.Errorf(errFmtSessionDomainListRequired, index))
		}

		config.ProtectedDomains = append(config.ProtectedDomains, config.Domains[index].Domains...)

		if config.Domains[index].Expiration <= 0 {
			config.Domains[index].Expiration = config.Expiration
		}

		if config.Domains[index].Inactivity <= 0 {
			config.Domains[index].Inactivity = config.Inactivity
		}

		if config.Domains[index].RememberMeDuration <= 0 && config.Domains[index].RememberMeDuration != schema.RememberMeDisabled {
			config.Domains[index].RememberMeDuration = config.RememberMeDuration
		}
	}
}

// sliceContainsString returns true if str is found in slice.
func sliceContainsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}

func validateRedisCommon(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if config.Secret == "" {
		validator.Push(fmt.Errorf(errFmtSessionSecretRequired, "redis"))
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
