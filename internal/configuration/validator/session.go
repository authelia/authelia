package validator

import (
	"fmt"
	"path"
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

	switch {
	case config.RememberMe == schema.RememberMeDisabled:
		config.DisableRememberMe = true
	case config.RememberMe <= 0:
		config.RememberMe = schema.DefaultSessionConfiguration.RememberMe // 1 month.
	}

	if config.SameSite == "" {
		config.SameSite = schema.DefaultSessionConfiguration.SameSite
	} else if !utils.IsStringInSlice(config.SameSite, validSessionSameSiteValues) {
		validator.Push(fmt.Errorf(errFmtSessionSameSite, strings.Join(validSessionSameSiteValues, "', '"), config.SameSite))
	}

	// Add legacy configuration to the domains list.
	if config.Domain != "" {
		config.Cookies = append(config.Cookies, schema.SessionCookieConfiguration{
			SessionCookieCommonConfiguration: schema.SessionCookieCommonConfiguration{
				Name:       config.Name,
				Domain:     config.Domain,
				SameSite:   config.SameSite,
				Expiration: config.Expiration,
				Inactivity: config.Inactivity,
				RememberMe: config.RememberMe,
			},
		})
	}

	validateSessionDomains(config, validator)
}

//nolint:gocyclo
func validateSessionDomains(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if len(config.Cookies) == 0 {
		validator.Push(fmt.Errorf(errFmtSessionOptionRequired, "domain"))
	}

	var domains []string

	for i, d := range config.Cookies {
		validateDomainName(i, d, validator)

		// Check the previous domains do not share a root domain.
		if utils.IsStringInSliceF(d.Domain, domains, utils.HasDomainSuffix) {
			if utils.IsStringInSlice(d.Domain, domains) {
				validator.Push(fmt.Errorf(errFmtSessionDomainDuplicate, sessionDomainDescriptor(i, d)))
			} else {
				validator.Push(fmt.Errorf(errFmtSessionDomainDuplicateCookieScope, sessionDomainDescriptor(i, d)))
			}
		}

		if d.PortalURL != nil && d.Domain != "" && !utils.IsURISafeRedirection(d.PortalURL, d.Domain) {
			if utils.IsURISecure(d.PortalURL) {
				validator.Push(fmt.Errorf(errFmtSessionDomainPortalURLNotInCookieScope, sessionDomainDescriptor(i, d), d.Domain, d.PortalURL))
			} else {
				validator.Push(fmt.Errorf(errFmtSessionDomainPortalURLInsecure, sessionDomainDescriptor(i, d), d.PortalURL))
			}
		}

		domains = append(domains, d.Domain)

		if d.Name == "" {
			config.Cookies[i].Name = config.Name
		}

		if d.Expiration <= 0 {
			config.Cookies[i].Expiration = config.Expiration
		}

		if d.Inactivity <= 0 {
			config.Cookies[i].Inactivity = config.Inactivity
		}

		if d.RememberMe <= 0 && !d.DisableRememberMe {
			if !config.DisableRememberMe {
				config.Cookies[i].RememberMe = config.RememberMe
			} else {
				config.Cookies[i].RememberMe = schema.DefaultSessionConfiguration.RememberMe
			}
		}

		if d.SameSite == "" {
			if utils.IsStringInSlice(config.SameSite, validSessionSameSiteValues) {
				config.Cookies[i].SameSite = config.SameSite
			} else {
				config.Cookies[i].SameSite = schema.DefaultSessionConfiguration.SameSite
			}
		} else if !utils.IsStringInSlice(d.SameSite, validSessionSameSiteValues) {
			validator.Push(fmt.Errorf(errFmtSessionDomainSameSite, sessionDomainDescriptor(i, d), strings.Join(validSessionSameSiteValues, "', '"), d.SameSite))
		}
	}
}

// validateDomainName returns error if the domain name is invalid.
func validateDomainName(i int, d schema.SessionCookieConfiguration, validator *schema.StructValidator) {
	switch {
	case d.Domain == "":
		validator.Push(fmt.Errorf(errFmtSessionDomainRequired, sessionDomainDescriptor(i, d)))
	case strings.HasPrefix(d.Domain, "*."):
		validator.Push(fmt.Errorf(errFmtSessionDomainMustBeRoot, sessionDomainDescriptor(i, d), d.Domain))
	case strings.HasPrefix(d.Domain, "."):
		validator.PushWarning(fmt.Errorf(errFmtSessionDomainHasPeriodPrefix, sessionDomainDescriptor(i, d)))
	case !reDomainCharacters.MatchString(d.Domain):
		validator.Push(fmt.Errorf(errFmtSessionDomainInvalidDomain, sessionDomainDescriptor(i, d)))
	}
}

func sessionDomainDescriptor(position int, domain schema.SessionCookieConfiguration) string {
	return fmt.Sprintf("#%d (domain '%s')", position+1, domain.Domain)
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

	if !path.IsAbs(config.Redis.Host) && (config.Redis.Port < 1 || config.Redis.Port > 65535) {
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
