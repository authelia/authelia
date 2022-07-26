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

	// adding legacy session.domain to domains list.
	if config.Domain != "" {
		last := len(config.Domains) - 1
		// workaround for failing tests when validator is called twice.
		if last < 0 || config.Domains[last].Domain != config.Domain {
			config.Domains = append(config.Domains, schema.SessionDomainConfiguration{
				Domain:    config.Domain,
				PortalURL: config.PortalURL,
			})
		}
	}

	if config.SameSite == "" {
		config.SameSite = schema.DefaultSessionConfiguration.SameSite
	} else if !utils.IsStringInSlice(config.SameSite, validSessionSameSiteValues) {
		validator.Push(fmt.Errorf(errFmtSessionSameSite, strings.Join(validSessionSameSiteValues, "', '"), config.SameSite))
	}
}

func validateSessionDomains(config *schema.SessionConfiguration, validator *schema.StructValidator) {
	if len(config.Domains) == 0 {
		validator.Push(fmt.Errorf(errFmtSessionOptionRequired, "domain"))
	}

	cookieDomainList := make([]string, 0, len(config.Domains))

	for index := range config.Domains {
		if err := validateDomainName(config.Domains[index].Domain); err != nil {
			validator.Push(fmt.Errorf(errFmtSessionDomainMustBeRoot, config.Domains[index].Domain))
		}

		// ensure there's not duplicated domain_cookie.
		if sliceContainsString(cookieDomainList, config.Domains[index].Domain) {
			validator.Push(fmt.Errorf(errFmtSessionDuplicatedDomainCookie, config.Domains[index].Domain, index))
		}

		// subdomains are not allowed.
		if sliceHasSuffix(cookieDomainList, config.Domains[index].Domain) {
			validator.Push(fmt.Errorf(errFmtSessionSubdomainConflict, config.Domains[index].Domain))
		}

		if err := validatePortalURL(config.Domains[index].PortalURL, config.Domains[index].Domain); err != nil {
			validator.PushWarning(fmt.Errorf(errFmtSessionPortalURLUndefined, config.Domains[index].Domain))
			config.Domains[index].PortalURL = ""
		}

		cookieDomainList = append(cookieDomainList, config.Domains[index].Domain)

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

// validateDomainName returns error if the domain name is invalid.
func validateDomainName(domain string) error {
	if domain == "" {
		return fmt.Errorf(errFmtSessionOptionRequired, "domain")
	}

	if strings.HasPrefix(domain, "*.") {
		return fmt.Errorf(errFmtSessionDomainMustBeRoot, domain)
	}

	// TODO: create a validation regexp.

	return nil
}

func validatePortalURL(url string, domain string) error {
	if url == "" {
		return fmt.Errorf(errFmtSessionPortalURLUndefined, domain)
	}

	// TODO: if domain is not part of url should return error.

	return nil
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

// sliceContainsString returns true if an element of slice has specified suffix(str) or str has a slice element as suffix.
func sliceHasSuffix(slice []string, str string) bool {
	for _, s := range slice {
		if strings.HasSuffix(s, str) {
			return true
		}

		if strings.HasSuffix(str, s) {
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
