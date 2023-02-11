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

	cookies := len(config.Cookies)

	switch {
	case cookies == 0 && config.Domain != "":
		// Add legacy configuration to the domains list.
		config.Cookies = append(config.Cookies, schema.SessionCookieConfiguration{
			SessionCookieCommonConfiguration: schema.SessionCookieCommonConfiguration{
				Name:              config.Name,
				Domain:            config.Domain,
				SameSite:          config.SameSite,
				Expiration:        config.Expiration,
				Inactivity:        config.Inactivity,
				RememberMe:        config.RememberMe,
				DisableRememberMe: config.DisableRememberMe,
			},
		})
	case cookies != 0 && config.Domain != "":
		validator.Push(fmt.Errorf(errFmtSessionLegacyAndWarning))
	}

	validateSessionCookieDomains(config, validator)
}

func validateSessionCookieDomains(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if len(config.Cookies) == 0 {
		validator.Push(fmt.Errorf(errFmtSessionOptionRequired, "domain"))
	}

	domains := make([]string, 0)

	for i, d := range config.Cookies {
		validateSessionDomainName(i, config, validator)

		validateSessionUniqueCookieDomain(i, config, domains, validator)

		validateSessionCookieName(i, config)

		validateSessionSafeRedirection(i, config, validator)

		validateSessionExpiration(i, config)

		validateSessionRememberMe(i, config)

		validateSessionSameSite(i, config, validator)

		domains = append(domains, d.Domain)
	}
}

// validateSessionDomainName returns error if the domain name is invalid.
func validateSessionDomainName(i int, config *schema.SessionConfiguration, validator *schema.StructValidator) {
	var d = config.Cookies[i]

	switch {
	case d.Domain == "":
		validator.Push(fmt.Errorf(errFmtSessionDomainRequired, sessionDomainDescriptor(i, d)))
		return
	case strings.HasPrefix(d.Domain, "*."):
		validator.Push(fmt.Errorf(errFmtSessionDomainMustBeRoot, sessionDomainDescriptor(i, d), d.Domain))
		return
	case strings.HasPrefix(d.Domain, "."):
		validator.PushWarning(fmt.Errorf(errFmtSessionDomainHasPeriodPrefix, sessionDomainDescriptor(i, d)))
	case !strings.Contains(d.Domain, "."):
		validator.Push(fmt.Errorf(errFmtSessionDomainInvalidDomainNoDots, sessionDomainDescriptor(i, d)))
		return
	case !reDomainCharacters.MatchString(d.Domain):
		validator.Push(fmt.Errorf(errFmtSessionDomainInvalidDomain, sessionDomainDescriptor(i, d)))
		return
	}

	if isCookieDomainAPublicSuffix(d.Domain) {
		validator.Push(fmt.Errorf(errFmtSessionDomainInvalidDomainPublic, sessionDomainDescriptor(i, d)))
	}
}

func validateSessionCookieName(i int, config *schema.SessionConfiguration) {
	if config.Cookies[i].Name == "" {
		config.Cookies[i].Name = config.Name
	}
}

func validateSessionExpiration(i int, config *schema.SessionConfiguration) {
	if config.Cookies[i].Expiration <= 0 {
		config.Cookies[i].Expiration = config.Expiration
	}

	if config.Cookies[i].Inactivity <= 0 {
		config.Cookies[i].Inactivity = config.Inactivity
	}
}

// validateSessionUniqueCookieDomain Check the current domains do not share a root domain with previous domains.
func validateSessionUniqueCookieDomain(i int, config *schema.SessionConfiguration, domains []string, validator *schema.StructValidator) {
	var d = config.Cookies[i]
	if utils.IsStringInSliceF(d.Domain, domains, utils.HasDomainSuffix) {
		if utils.IsStringInSlice(d.Domain, domains) {
			validator.Push(fmt.Errorf(errFmtSessionDomainDuplicate, sessionDomainDescriptor(i, d)))
		} else {
			validator.Push(fmt.Errorf(errFmtSessionDomainDuplicateCookieScope, sessionDomainDescriptor(i, d)))
		}
	}
}

// validateSessionSafeRedirection validates that AutheliaURL is safe for redirection.
func validateSessionSafeRedirection(index int, config *schema.SessionConfiguration, validator *schema.StructValidator) {
	var d = config.Cookies[index]

	if d.AutheliaURL != nil && d.Domain != "" && !utils.IsURISafeRedirection(d.AutheliaURL, d.Domain) {
		if utils.IsURISecure(d.AutheliaURL) {
			validator.Push(fmt.Errorf(errFmtSessionDomainPortalURLNotInCookieScope, sessionDomainDescriptor(index, d), d.Domain, d.AutheliaURL))
		} else {
			validator.Push(fmt.Errorf(errFmtSessionDomainPortalURLInsecure, sessionDomainDescriptor(index, d), d.AutheliaURL))
		}
	}
}

func validateSessionRememberMe(i int, config *schema.SessionConfiguration) {
	if config.Cookies[i].RememberMe <= 0 && config.Cookies[i].RememberMe != schema.RememberMeDisabled {
		config.Cookies[i].RememberMe = config.RememberMe
	}

	if config.Cookies[i].RememberMe == schema.RememberMeDisabled {
		config.Cookies[i].DisableRememberMe = true
	}
}

func validateSessionSameSite(i int, config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if config.Cookies[i].SameSite == "" {
		if utils.IsStringInSlice(config.SameSite, validSessionSameSiteValues) {
			config.Cookies[i].SameSite = config.SameSite
		} else {
			config.Cookies[i].SameSite = schema.DefaultSessionConfiguration.SameSite
		}
	} else if !utils.IsStringInSlice(config.Cookies[i].SameSite, validSessionSameSiteValues) {
		validator.Push(fmt.Errorf(errFmtSessionDomainSameSite, sessionDomainDescriptor(i, config.Cookies[i]), strings.Join(validSessionSameSiteValues, "', '"), config.Cookies[i].SameSite))
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
