package validator

import (
	"errors"
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateSession validates and update session configuration.
func ValidateSession(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Session.Name == "" {
		config.Session.Name = schema.DefaultSessionConfiguration.Name
	}

	if config.Session.Redis != nil {
		if config.Session.Redis.HighAvailability != nil {
			validateRedisSentinel(&config.Session, validator)
		} else {
			validateRedis(&config.Session, validator)
		}
	}

	validateSession(config, validator)
}

func validateSession(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Session.Expiration <= 0 {
		config.Session.Expiration = schema.DefaultSessionConfiguration.Expiration // 1 hour.
	}

	if config.Session.Inactivity <= 0 {
		config.Session.Inactivity = schema.DefaultSessionConfiguration.Inactivity // 5 min.
	}

	switch {
	case config.Session.RememberMe == schema.RememberMeDisabled:
		config.Session.DisableRememberMe = true
	case config.Session.RememberMe <= 0:
		config.Session.RememberMe = schema.DefaultSessionConfiguration.RememberMe // 1 month.
	}

	if config.Session.SameSite == "" {
		config.Session.SameSite = schema.DefaultSessionConfiguration.SameSite
	} else if !utils.IsStringInSlice(config.Session.SameSite, validSessionSameSiteValues) {
		validator.Push(fmt.Errorf(errFmtSessionSameSite, utils.StringJoinOr(validSessionSameSiteValues), config.Session.SameSite))
	}

	cookies := len(config.Session.Cookies)
	n := len(config.Session.Domain) //nolint:staticcheck

	if cookies != 0 && config.DefaultRedirectionURL != nil { //nolint:staticcheck
		validator.Push(errors.New(errFmtSessionLegacyRedirectionURL))
	}

	switch {
	case cookies == 0 && n != 0:
		validator.PushWarning(errors.New(errFmtSessionDomainLegacy))
		// Add legacy configuration to the domains list.
		config.Session.Cookies = append(config.Session.Cookies, schema.SessionCookie{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:              config.Session.Name,
				SameSite:          config.Session.SameSite,
				Expiration:        config.Session.Expiration,
				Inactivity:        config.Session.Inactivity,
				RememberMe:        config.Session.RememberMe,
				DisableRememberMe: config.Session.DisableRememberMe,
			},
			Domain:                config.Session.Domain,        //nolint:staticcheck
			DefaultRedirectionURL: config.DefaultRedirectionURL, //nolint:staticcheck
			Legacy:                true,
		})
	case cookies != 0 && n != 0:
		validator.Push(errors.New(errFmtSessionLegacyAndWarning))
	}

	validateSessionCookieDomains(&config.Session, validator)
}

func validateSessionCookieDomains(config *schema.Session, validator *schema.StructValidator) {
	if len(config.Cookies) == 0 {
		validator.Push(fmt.Errorf(errFmtSessionOptionRequired, "cookies"))
	}

	domains := make([]string, 0, len(config.Cookies))

	for i, d := range config.Cookies {
		validateSessionDomainName(i, config, validator)

		validateSessionUniqueCookieDomain(i, config, domains, validator)

		validateSessionCookieName(i, config)

		validateSessionCookiesURLs(i, config, validator)

		validateSessionExpiration(i, config)

		validateSessionRememberMe(i, config)

		validateSessionSameSite(i, config, validator)

		domains = append(domains, d.Domain)
	}
}

// validateSessionDomainName returns error if the domain name is invalid.
func validateSessionDomainName(i int, config *schema.Session, validator *schema.StructValidator) {
	var d = config.Cookies[i]

	switch {
	case d.Domain == "":
		validator.Push(fmt.Errorf(errFmtSessionDomainOptionRequired, sessionDomainDescriptor(i, d), attrSessionDomain))
		return
	case strings.HasPrefix(d.Domain, "*."):
		validator.Push(fmt.Errorf(errFmtSessionDomainMustBeRoot, sessionDomainDescriptor(i, d), d.Domain))
		return
	case strings.HasPrefix(d.Domain, "."):
		validator.PushWarning(fmt.Errorf(errFmtSessionDomainHasPeriodPrefix, sessionDomainDescriptor(i, d)))
	case net.ParseIP(d.Domain) != nil:
		return
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

func validateSessionCookieName(i int, config *schema.Session) {
	if config.Cookies[i].Name == "" {
		config.Cookies[i].Name = config.Name
	}
}

func validateSessionExpiration(i int, config *schema.Session) {
	if config.Cookies[i].Expiration <= 0 {
		config.Cookies[i].Expiration = config.Expiration
	}

	if config.Cookies[i].Inactivity <= 0 {
		config.Cookies[i].Inactivity = config.Inactivity
	}
}

// validateSessionUniqueCookieDomain Check the current domains do not share a root domain with previous domains.
func validateSessionUniqueCookieDomain(i int, config *schema.Session, domains []string, validator *schema.StructValidator) {
	var d = config.Cookies[i]
	if utils.IsStringInSliceF(d.Domain, domains, utils.HasDomainSuffix) {
		if utils.IsStringInSlice(d.Domain, domains) {
			validator.Push(fmt.Errorf(errFmtSessionDomainDuplicate, sessionDomainDescriptor(i, d)))
		} else {
			validator.Push(fmt.Errorf(errFmtSessionDomainDuplicateCookieScope, sessionDomainDescriptor(i, d)))
		}
	}
}

// validateSessionCookiesURLs validates the AutheliaURL and DefaultRedirectionURL.
//
//nolint:gocyclo
func validateSessionCookiesURLs(i int, config *schema.Session, validator *schema.StructValidator) {
	var d = config.Cookies[i]

	if d.AutheliaURL == nil {
		if !d.Legacy && d.Domain != "" {
			validator.Push(fmt.Errorf(errFmtSessionDomainOptionRequired, sessionDomainDescriptor(i, d), attrSessionAutheliaURL))
		}
	} else {
		switch d.AutheliaURL.Path {
		case "", "/":
			break
		default:
			if strings.HasSuffix(d.AutheliaURL.Path, "/") || !d.AutheliaURL.IsAbs() {
				break
			}

			d.AutheliaURL.Path += "/"
		}

		if !d.AutheliaURL.IsAbs() {
			validator.Push(fmt.Errorf(errFmtSessionDomainURLNotAbsolute, sessionDomainDescriptor(i, d), attrSessionAutheliaURL, d.AutheliaURL))
		} else if !utils.IsURISecure(d.AutheliaURL) {
			validator.Push(fmt.Errorf(errFmtSessionDomainURLInsecure, sessionDomainDescriptor(i, d), attrSessionAutheliaURL, d.AutheliaURL))
		}

		if d.Domain != "" && !utils.HasURIDomainSuffix(d.AutheliaURL, d.Domain) {
			validator.Push(fmt.Errorf(errFmtSessionDomainURLNotInCookieScope, sessionDomainDescriptor(i, d), attrSessionAutheliaURL, d.Domain, d.AutheliaURL))
		}
	}

	if d.DefaultRedirectionURL != nil {
		if !d.DefaultRedirectionURL.IsAbs() {
			validator.Push(fmt.Errorf(errFmtSessionDomainURLNotAbsolute, sessionDomainDescriptor(i, d), attrDefaultRedirectionURL, d.DefaultRedirectionURL))
		} else if !utils.IsURISecure(d.DefaultRedirectionURL) {
			validator.Push(fmt.Errorf(errFmtSessionDomainURLInsecure, sessionDomainDescriptor(i, d), attrDefaultRedirectionURL, d.DefaultRedirectionURL))
		}

		if d.Domain != "" && !utils.HasURIDomainSuffix(d.DefaultRedirectionURL, d.Domain) {
			if d.Legacy {
				validator.PushWarning(fmt.Errorf(errFmtSessionDomainURLNotInCookieScope, sessionDomainDescriptor(i, d), attrDefaultRedirectionURL, d.Domain, d.DefaultRedirectionURL))
				d.DefaultRedirectionURL = nil
			} else {
				validator.Push(fmt.Errorf(errFmtSessionDomainURLNotInCookieScope, sessionDomainDescriptor(i, d), attrDefaultRedirectionURL, d.Domain, d.DefaultRedirectionURL))
			}
		}

		if d.AutheliaURL != nil && utils.EqualURLs(d.AutheliaURL, d.DefaultRedirectionURL) {
			validator.Push(fmt.Errorf(errFmtSessionDomainAutheliaURLAndRedirectionURLEqual, sessionDomainDescriptor(i, d), d.DefaultRedirectionURL, d.AutheliaURL))
		}
	}

	config.Cookies[i] = d
}

func validateSessionRememberMe(i int, config *schema.Session) {
	if config.Cookies[i].RememberMe <= 0 && config.Cookies[i].RememberMe != schema.RememberMeDisabled {
		config.Cookies[i].RememberMe = config.RememberMe
	}

	if config.Cookies[i].RememberMe == schema.RememberMeDisabled {
		config.Cookies[i].DisableRememberMe = true
	}
}

func validateSessionSameSite(i int, config *schema.Session, validator *schema.StructValidator) {
	if config.Cookies[i].SameSite == "" {
		if utils.IsStringInSlice(config.SameSite, validSessionSameSiteValues) {
			config.Cookies[i].SameSite = config.SameSite
		} else {
			config.Cookies[i].SameSite = schema.DefaultSessionConfiguration.SameSite
		}
	} else if !utils.IsStringInSlice(config.Cookies[i].SameSite, validSessionSameSiteValues) {
		validator.Push(fmt.Errorf(errFmtSessionDomainSameSite, sessionDomainDescriptor(i, config.Cookies[i]), utils.StringJoinOr(validSessionSameSiteValues), config.Cookies[i].SameSite))
	}
}

func sessionDomainDescriptor(position int, domain schema.SessionCookie) string {
	return fmt.Sprintf("#%d (domain '%s')", position+1, domain.Domain)
}

func validateRedisCommon(config *schema.Session, validator *schema.StructValidator) {
	if config.Secret == "" {
		validator.Push(fmt.Errorf(errFmtSessionSecretRequired, "redis"))
	}

	if config.Redis.TLS != nil {
		configDefaultTLS := &schema.TLS{
			ServerName:     config.Redis.Host,
			MinimumVersion: schema.DefaultRedisConfiguration.TLS.MinimumVersion,
			MaximumVersion: schema.DefaultRedisConfiguration.TLS.MaximumVersion,
		}

		if err := ValidateTLSConfig(config.Redis.TLS, configDefaultTLS); err != nil {
			validator.Push(fmt.Errorf(errFmtSessionRedisTLSConfigInvalid, err))
		}
	}
}

func validateRedis(config *schema.Session, validator *schema.StructValidator) {
	if config.Redis.Host == "" {
		validator.Push(errors.New(errFmtSessionRedisHostRequired))
	}

	validateRedisCommon(config, validator)

	abs := path.IsAbs(config.Redis.Host)

	if !abs && config.Redis.Port == 0 {
		config.Redis.Port = schema.DefaultRedisConfiguration.Port
	} else if !abs && (config.Redis.Port < 1 || config.Redis.Port > 65535) {
		validator.Push(fmt.Errorf(errFmtSessionRedisPortRange, config.Redis.Port))
	}

	if config.Redis.MaximumActiveConnections <= 0 {
		config.Redis.MaximumActiveConnections = schema.DefaultRedisConfiguration.MaximumActiveConnections
	}
}

func validateRedisSentinel(config *schema.Session, validator *schema.StructValidator) {
	if config.Redis.HighAvailability.SentinelName == "" {
		validator.Push(errors.New(errFmtSessionRedisSentinelMissingName))
	}

	if config.Redis.Port == 0 {
		config.Redis.Port = 26379
	} else if config.Redis.Port < 1 || config.Redis.Port > 65535 {
		validator.Push(fmt.Errorf(errFmtSessionRedisPortRange, config.Redis.Port))
	}

	if config.Redis.Host == "" && len(config.Redis.HighAvailability.Nodes) == 0 {
		validator.Push(errors.New(errFmtSessionRedisHostOrNodesRequired))
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
		validator.Push(errors.New(errFmtSessionRedisSentinelNodeHostMissing))
	}
}
