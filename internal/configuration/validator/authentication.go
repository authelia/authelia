package validator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateAuthenticationBackend validates and updates the authentication backend configuration.
func ValidateAuthenticationBackend(config *schema.AuthenticationBackendConfiguration, validator *schema.StructValidator) {
	if config.LDAP == nil && config.File == nil {
		validator.Push(fmt.Errorf(errFmtAuthBackendNotConfigured))
	}

	if config.LDAP != nil && config.File != nil {
		validator.Push(fmt.Errorf(errFmtAuthBackendMultipleConfigured))
	}

	if config.File != nil {
		validateFileAuthenticationBackend(config.File, validator)
	} else if config.LDAP != nil {
		validateLDAPAuthenticationBackend(config.LDAP, validator)
	}

	if config.RefreshInterval == "" {
		config.RefreshInterval = schema.RefreshIntervalDefault
	} else {
		_, err := utils.ParseDurationString(config.RefreshInterval)
		if err != nil && config.RefreshInterval != schema.ProfileRefreshDisabled && config.RefreshInterval != schema.ProfileRefreshAlways {
			validator.Push(fmt.Errorf(errFmtAuthBackendRefreshInterval, config.RefreshInterval, err))
		}
	}
}

// validateFileAuthenticationBackend validates and updates the file authentication backend configuration.
func validateFileAuthenticationBackend(config *schema.FileAuthenticationBackendConfiguration, validator *schema.StructValidator) {
	if config.Path == "" {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPathNotConfigured))
	}

	if config.Password == nil {
		config.Password = &schema.DefaultPasswordConfiguration
	} else {
		// Salt Length.
		switch {
		case config.Password.SaltLength == 0:
			config.Password.SaltLength = schema.DefaultPasswordConfiguration.SaltLength
		case config.Password.SaltLength < 8:
			validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordSaltLength, config.Password.SaltLength))
		}

		switch config.Password.Algorithm {
		case "":
			config.Password.Algorithm = schema.DefaultPasswordConfiguration.Algorithm
			fallthrough
		case hashArgon2id:
			validateFileAuthenticationBackendArgon2id(config, validator)
		case hashSHA512:
			validateFileAuthenticationBackendSHA512(config)
		default:
			validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordUnknownAlg, config.Password.Algorithm))
		}

		if config.Password.Iterations < 1 {
			validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidIterations, config.Password.Iterations))
		}
	}
}

func validateFileAuthenticationBackendSHA512(config *schema.FileAuthenticationBackendConfiguration) {
	// Iterations (time).
	if config.Password.Iterations == 0 {
		config.Password.Iterations = schema.DefaultPasswordSHA512Configuration.Iterations
	}
}
func validateFileAuthenticationBackendArgon2id(config *schema.FileAuthenticationBackendConfiguration, validator *schema.StructValidator) {
	// Iterations (time).
	if config.Password.Iterations == 0 {
		config.Password.Iterations = schema.DefaultPasswordConfiguration.Iterations
	}

	// Parallelism.
	if config.Password.Parallelism == 0 {
		config.Password.Parallelism = schema.DefaultPasswordConfiguration.Parallelism
	} else if config.Password.Parallelism < 1 {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordArgon2idInvalidParallelism, config.Password.Parallelism))
	}

	// Memory.
	if config.Password.Memory == 0 {
		config.Password.Memory = schema.DefaultPasswordConfiguration.Memory
	} else if config.Password.Memory < config.Password.Parallelism*8 {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordArgon2idInvalidMemory, config.Password.Parallelism, config.Password.Parallelism*8, config.Password.Memory))
	}

	// Key Length.
	if config.Password.KeyLength == 0 {
		config.Password.KeyLength = schema.DefaultPasswordConfiguration.KeyLength
	} else if config.Password.KeyLength < 16 {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordArgon2idInvalidKeyLength, config.Password.KeyLength))
	}
}

func validateLDAPAuthenticationBackend(config *schema.LDAPAuthenticationBackendConfiguration, validator *schema.StructValidator) {
	if config.Timeout == 0 {
		config.Timeout = schema.DefaultLDAPAuthenticationBackendConfiguration.Timeout
	}

	if config.Implementation == "" {
		config.Implementation = schema.DefaultLDAPAuthenticationBackendConfiguration.Implementation
	}

	if config.TLS == nil {
		config.TLS = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS
	}

	if config.TLS.MinimumVersion == "" {
		config.TLS.MinimumVersion = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS.MinimumVersion
	}

	if _, err := utils.TLSStringToTLSConfigVersion(config.TLS.MinimumVersion); err != nil {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendTLSMinVersion, config.TLS.MinimumVersion, err))
	}

	switch config.Implementation {
	case schema.LDAPImplementationCustom:
		setDefaultImplementationCustomLDAPAuthenticationBackend(config)
	case schema.LDAPImplementationActiveDirectory:
		setDefaultImplementationActiveDirectoryLDAPAuthenticationBackend(config)
	default:
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendImplementation, config.Implementation, strings.Join([]string{schema.LDAPImplementationCustom, schema.LDAPImplementationActiveDirectory}, "', '")))
	}

	if strings.Contains(config.UsersFilter, "{0}") {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterReplacedPlaceholders, "users_filter", "{0}", "{input}"))
	}

	if strings.Contains(config.GroupsFilter, "{0}") {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterReplacedPlaceholders, "groups_filter", "{0}", "{input}"))
	}

	if strings.Contains(config.GroupsFilter, "{1}") {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterReplacedPlaceholders, "groups_filter", "{1}", "{username}"))
	}

	if config.URL == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "url"))
	} else {
		validateLDAPAuthenticationBackendURL(config, validator)
	}

	validateLDAPRequiredParameters(config, validator)
}

func validateLDAPAuthenticationBackendURL(config *schema.LDAPAuthenticationBackendConfiguration, validator *schema.StructValidator) {
	var (
		parsedURL *url.URL
		err       error
	)

	if parsedURL, err = url.Parse(config.URL); err != nil {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendURLNotParsable, err))

		return
	}

	if parsedURL.Scheme != schemeLDAP && parsedURL.Scheme != schemeLDAPS {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendURLInvalidScheme, parsedURL.Scheme))

		return
	}

	config.URL = parsedURL.String()
	if config.TLS.ServerName == "" {
		config.TLS.ServerName = parsedURL.Hostname()
	}
}

func validateLDAPRequiredParameters(config *schema.LDAPAuthenticationBackendConfiguration, validator *schema.StructValidator) {
	// TODO: see if it's possible to disable this check if disable_reset_password is set and when anonymous/user binding is supported (#101 and #387).
	if config.User == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "user"))
	}

	// TODO: see if it's possible to disable this check if disable_reset_password is set and when anonymous/user binding is supported (#101 and #387).
	if config.Password == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "password"))
	}

	if config.BaseDN == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "base_dn"))
	}

	if config.UsersFilter == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "users_filter"))
	} else {
		if !strings.HasPrefix(config.UsersFilter, "(") || !strings.HasSuffix(config.UsersFilter, ")") {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterEnclosingParenthesis, "users_filter", config.UsersFilter, config.UsersFilter))
		}

		if !strings.Contains(config.UsersFilter, "{username_attribute}") {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterMissingPlaceholder, "users_filter", "username_attribute"))
		}

		// This test helps the user know that users_filter is broken after the breaking change induced by this commit.
		if !strings.Contains(config.UsersFilter, "{input}") {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterMissingPlaceholder, "users_filter", "input"))
		}
	}

	if config.GroupsFilter == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "groups_filter"))
	} else if !strings.HasPrefix(config.GroupsFilter, "(") || !strings.HasSuffix(config.GroupsFilter, ")") {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterEnclosingParenthesis, "groups_filter", config.GroupsFilter, config.GroupsFilter))
	}
}

func setDefaultImplementationActiveDirectoryLDAPAuthenticationBackend(config *schema.LDAPAuthenticationBackendConfiguration) {
	if config.UsersFilter == "" {
		config.UsersFilter = schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsersFilter
	}

	if config.UsernameAttribute == "" {
		config.UsernameAttribute = schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsernameAttribute
	}

	if config.DisplayNameAttribute == "" {
		config.DisplayNameAttribute = schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.DisplayNameAttribute
	}

	if config.MailAttribute == "" {
		config.MailAttribute = schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.MailAttribute
	}

	if config.GroupsFilter == "" {
		config.GroupsFilter = schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupsFilter
	}

	if config.GroupNameAttribute == "" {
		config.GroupNameAttribute = schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupNameAttribute
	}
}

func setDefaultImplementationCustomLDAPAuthenticationBackend(config *schema.LDAPAuthenticationBackendConfiguration) {
	if config.UsernameAttribute == "" {
		config.UsernameAttribute = schema.DefaultLDAPAuthenticationBackendConfiguration.UsernameAttribute
	}

	if config.GroupNameAttribute == "" {
		config.GroupNameAttribute = schema.DefaultLDAPAuthenticationBackendConfiguration.GroupNameAttribute
	}

	if config.MailAttribute == "" {
		config.MailAttribute = schema.DefaultLDAPAuthenticationBackendConfiguration.MailAttribute
	}

	if config.DisplayNameAttribute == "" {
		config.DisplayNameAttribute = schema.DefaultLDAPAuthenticationBackendConfiguration.DisplayNameAttribute
	}
}
