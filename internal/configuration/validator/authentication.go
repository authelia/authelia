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
		validateLDAPAuthenticationBackend(config, validator)
	}

	if config.RefreshInterval == "" {
		config.RefreshInterval = schema.RefreshIntervalDefault
	} else {
		_, err := utils.ParseDurationString(config.RefreshInterval)
		if err != nil && config.RefreshInterval != schema.ProfileRefreshDisabled && config.RefreshInterval != schema.ProfileRefreshAlways {
			validator.Push(fmt.Errorf(errFmtAuthBackendRefreshInterval, config.RefreshInterval, err))
		}
	}

	if config.PasswordReset.CustomURL.String() != "" {
		switch config.PasswordReset.CustomURL.Scheme {
		case schemeHTTP, schemeHTTPS:
			config.PasswordReset.Disable = false
		default:
			validator.Push(fmt.Errorf(errFmtAuthBackendPasswordResetCustomURLScheme, config.PasswordReset.CustomURL.String(), config.PasswordReset.CustomURL.Scheme))
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
		ValidatePasswordConfiguration(config.Password, validator)
	}
}

// ValidatePasswordConfiguration validates the file auth backend password configuration.
func ValidatePasswordConfiguration(config *schema.PasswordConfiguration, validator *schema.StructValidator) {
	// Salt Length.
	switch {
	case config.SaltLength == 0:
		config.SaltLength = schema.DefaultPasswordConfiguration.SaltLength
	case config.SaltLength < 8:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordSaltLength, config.SaltLength))
	}

	switch config.Algorithm {
	case "":
		config.Algorithm = schema.DefaultPasswordConfiguration.Algorithm
		fallthrough
	case hashArgon2id:
		validateFileAuthenticationBackendArgon2id(config, validator)
	case hashSHA512:
		validateFileAuthenticationBackendSHA512(config)
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordUnknownAlg, config.Algorithm))
	}

	if config.Iterations < 1 {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidIterations, config.Iterations))
	}
}

func validateFileAuthenticationBackendSHA512(config *schema.PasswordConfiguration) {
	// Iterations (time).
	if config.Iterations == 0 {
		config.Iterations = schema.DefaultPasswordSHA512Configuration.Iterations
	}
}
func validateFileAuthenticationBackendArgon2id(config *schema.PasswordConfiguration, validator *schema.StructValidator) {
	// Iterations (time).
	if config.Iterations == 0 {
		config.Iterations = schema.DefaultPasswordConfiguration.Iterations
	}

	// Parallelism.
	if config.Parallelism == 0 {
		config.Parallelism = schema.DefaultPasswordConfiguration.Parallelism
	} else if config.Parallelism < 1 {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordArgon2idInvalidParallelism, config.Parallelism))
	}

	// Memory.
	if config.Memory == 0 {
		config.Memory = schema.DefaultPasswordConfiguration.Memory
	} else if config.Memory < config.Parallelism*8 {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordArgon2idInvalidMemory, config.Parallelism, config.Parallelism*8, config.Memory))
	}

	// Key Length.
	if config.KeyLength == 0 {
		config.KeyLength = schema.DefaultPasswordConfiguration.KeyLength
	} else if config.KeyLength < 16 {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordArgon2idInvalidKeyLength, config.KeyLength))
	}
}

func validateLDAPAuthenticationBackend(config *schema.AuthenticationBackendConfiguration, validator *schema.StructValidator) {
	if config.LDAP.Timeout == 0 {
		config.LDAP.Timeout = schema.DefaultLDAPAuthenticationBackendConfiguration.Timeout
	}

	if config.LDAP.Implementation == "" {
		config.LDAP.Implementation = schema.DefaultLDAPAuthenticationBackendConfiguration.Implementation
	}

	if config.LDAP.TLS == nil {
		config.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS
	} else if config.LDAP.TLS.MinimumVersion == "" {
		config.LDAP.TLS.MinimumVersion = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS.MinimumVersion
	}

	if _, err := utils.TLSStringToTLSConfigVersion(config.LDAP.TLS.MinimumVersion); err != nil {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendTLSMinVersion, config.LDAP.TLS.MinimumVersion, err))
	}

	switch config.LDAP.Implementation {
	case schema.LDAPImplementationCustom:
		setDefaultImplementationCustomLDAPAuthenticationBackend(config.LDAP)
	case schema.LDAPImplementationActiveDirectory:
		setDefaultImplementationActiveDirectoryLDAPAuthenticationBackend(config.LDAP)
	default:
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendImplementation, config.LDAP.Implementation, strings.Join([]string{schema.LDAPImplementationCustom, schema.LDAPImplementationActiveDirectory}, "', '")))
	}

	if strings.Contains(config.LDAP.UsersFilter, "{0}") {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterReplacedPlaceholders, "users_filter", "{0}", "{input}"))
	}

	if strings.Contains(config.LDAP.GroupsFilter, "{0}") {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterReplacedPlaceholders, "groups_filter", "{0}", "{input}"))
	}

	if strings.Contains(config.LDAP.GroupsFilter, "{1}") {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterReplacedPlaceholders, "groups_filter", "{1}", "{username}"))
	}

	if config.LDAP.URL == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "url"))
	} else {
		validateLDAPAuthenticationBackendURL(config.LDAP, validator)
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

func validateLDAPRequiredParameters(config *schema.AuthenticationBackendConfiguration, validator *schema.StructValidator) {
	if config.LDAP.PermitUnauthenticatedBind {
		if config.LDAP.Password != "" {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendUnauthenticatedBindWithPassword))
		}

		if !config.PasswordReset.Disable {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendUnauthenticatedBindWithResetEnabled))
		}
	} else {
		if config.LDAP.User == "" {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "user"))
		}

		if config.LDAP.Password == "" {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "password"))
		}
	}

	if config.LDAP.BaseDN == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "base_dn"))
	}

	if config.LDAP.UsersFilter == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "users_filter"))
	} else {
		if !strings.HasPrefix(config.LDAP.UsersFilter, "(") || !strings.HasSuffix(config.LDAP.UsersFilter, ")") {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterEnclosingParenthesis, "users_filter", config.LDAP.UsersFilter, config.LDAP.UsersFilter))
		}

		if !strings.Contains(config.LDAP.UsersFilter, "{username_attribute}") {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterMissingPlaceholder, "users_filter", "username_attribute"))
		}

		// This test helps the user know that users_filter is broken after the breaking change induced by this commit.
		if !strings.Contains(config.LDAP.UsersFilter, "{input}") {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterMissingPlaceholder, "users_filter", "input"))
		}
	}

	if config.LDAP.GroupsFilter == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "groups_filter"))
	} else if !strings.HasPrefix(config.LDAP.GroupsFilter, "(") || !strings.HasSuffix(config.LDAP.GroupsFilter, ")") {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterEnclosingParenthesis, "groups_filter", config.LDAP.GroupsFilter, config.LDAP.GroupsFilter))
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
