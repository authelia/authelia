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
func validateFileAuthenticationBackend(config *schema.FileAuthenticationBackendConfig, validator *schema.StructValidator) {
	if config.Path == "" {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPathNotConfigured))
	}

	if config.Password == nil {
		config.Password = &schema.DefaultPasswordConfig
	} else {
		ValidatePasswordConfiguration(config.Password, validator)
	}
}

// ValidatePasswordConfiguration validates the file auth backend password configuration.
func ValidatePasswordConfiguration(config *schema.PasswordConfig, validator *schema.StructValidator) {
	validateFileAuthenticationBackendPasswordConfigLegacy(config)

	if config.Algorithm == "" {
		config.Algorithm = schema.DefaultPasswordConfig.Algorithm
	}

	validateFileAuthenticationBackendPasswordConfigArgon2(config, validator)
	validateFileAuthenticationBackendPasswordConfigSHA2Crypt(config, validator)
}

func validateFileAuthenticationBackendPasswordConfigArgon2(config *schema.PasswordConfig, validator *schema.StructValidator) {
	switch config.Argon2.Variant {
	case "":
		config.Argon2.Variant = schema.DefaultPasswordConfig.Argon2.Variant
	case "argon2id", "argon2i", "argon2d":
		break
	default:
		validator.Push(fmt.Errorf("TODO"))
	}

	if config.Argon2.Iterations == 0 {
		config.Argon2.Iterations = schema.DefaultPasswordConfig.Argon2.Iterations
	}

	if config.Argon2.Iterations == 0 {
		config.Argon2.Iterations = schema.DefaultPasswordConfig.Argon2.Iterations
	}

	if config.Argon2.Memory == 0 {
		config.Argon2.Memory = schema.DefaultPasswordConfig.Argon2.Memory
	}

	if config.Argon2.Parallelism == 0 {
		config.Argon2.Parallelism = schema.DefaultPasswordConfig.Argon2.Parallelism
	}

	if config.Argon2.KeyLength == 0 {
		config.Argon2.KeyLength = schema.DefaultPasswordConfig.Argon2.KeyLength
	}

	if config.Argon2.SaltLength == 0 {
		config.Argon2.SaltLength = schema.DefaultPasswordConfig.Argon2.SaltLength
	}
}

func validateFileAuthenticationBackendPasswordConfigSHA2Crypt(config *schema.PasswordConfig, validator *schema.StructValidator) {
	switch config.SHA2Crypt.Variant {
	case "":
		config.SHA2Crypt.Variant = schema.DefaultPasswordConfig.SHA2Crypt.Variant
	case "sha256", "sha512":
		break
	default:
		validator.Push(fmt.Errorf("TODO"))
	}

	if config.SHA2Crypt.Rounds == 0 {
		config.SHA2Crypt.Rounds = schema.DefaultPasswordConfig.SHA2Crypt.Rounds
	}

	if config.SHA2Crypt.SaltLength == 0 {
		config.SHA2Crypt.SaltLength = schema.DefaultPasswordConfig.SHA2Crypt.SaltLength
	}
}

func validateFileAuthenticationBackendPasswordConfigLegacy(config *schema.PasswordConfig) {
	switch config.Algorithm {
	case hashSHA512:
		config.Algorithm = "sha2crypt"
		config.SHA2Crypt.Variant = hashSHA512

		if config.Iterations > 0 {
			config.SHA2Crypt.Rounds = uint32(config.Iterations)
		}

		if config.SaltLength > 0 {
			config.SHA2Crypt.SaltLength = uint32(config.SaltLength)
		}
	case hashArgon2id:
		config.Algorithm = "argon2"
		config.Argon2.Variant = hashArgon2id

		if config.Iterations > 0 {
			config.Argon2.Iterations = uint32(config.Iterations)
		}

		if config.Memory > 0 {
			config.Argon2.Memory = uint32(config.Memory) * 1024
		}

		if config.Parallelism > 0 {
			config.Argon2.Parallelism = uint32(config.Parallelism)
		}

		if config.KeyLength > 0 {
			config.Argon2.KeyLength = uint32(config.KeyLength)
		}

		if config.SaltLength > 0 {
			config.Argon2.SaltLength = uint32(config.SaltLength)
		}
	}
}

func validateLDAPAuthenticationBackend(config *schema.AuthenticationBackendConfiguration, validator *schema.StructValidator) {
	if config.LDAP.Implementation == "" {
		config.LDAP.Implementation = schema.LDAPImplementationCustom
	}

	var implementation *schema.LDAPAuthenticationBackendConfiguration

	switch config.LDAP.Implementation {
	case schema.LDAPImplementationCustom:
		implementation = &schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom
	case schema.LDAPImplementationActiveDirectory:
		implementation = &schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory
	default:
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendImplementation, config.LDAP.Implementation, strings.Join([]string{schema.LDAPImplementationCustom, schema.LDAPImplementationActiveDirectory}, "', '")))
	}

	if implementation != nil {
		setDefaultImplementationLDAPAuthenticationBackendProfileMisc(config.LDAP, implementation)
		setDefaultImplementationLDAPAuthenticationBackendProfileAttributes(config.LDAP, implementation)
	}

	if config.LDAP.TLS != nil {
		if _, err := utils.TLSStringToTLSConfigVersion(config.LDAP.TLS.MinimumVersion); err != nil {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendTLSMinVersion, config.LDAP.TLS.MinimumVersion, err))
		}
	} else {
		config.LDAP.TLS = &schema.TLSConfig{}
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

func setDefaultImplementationLDAPAuthenticationBackendProfileMisc(config *schema.LDAPAuthenticationBackendConfiguration, implementation *schema.LDAPAuthenticationBackendConfiguration) {
	if config.Timeout == 0 {
		config.Timeout = implementation.Timeout
	}

	if implementation.TLS == nil {
		return
	}

	if config.TLS == nil {
		config.TLS = implementation.TLS
	} else if config.TLS.MinimumVersion == "" {
		config.TLS.MinimumVersion = implementation.TLS.MinimumVersion
	}
}

func ldapImplementationShouldSetStr(config, implementation string) bool {
	return config == "" && implementation != ""
}

func setDefaultImplementationLDAPAuthenticationBackendProfileAttributes(config *schema.LDAPAuthenticationBackendConfiguration, implementation *schema.LDAPAuthenticationBackendConfiguration) {
	if ldapImplementationShouldSetStr(config.UsersFilter, implementation.UsersFilter) {
		config.UsersFilter = implementation.UsersFilter
	}

	if ldapImplementationShouldSetStr(config.UsernameAttribute, implementation.UsernameAttribute) {
		config.UsernameAttribute = implementation.UsernameAttribute
	}

	if ldapImplementationShouldSetStr(config.DisplayNameAttribute, implementation.DisplayNameAttribute) {
		config.DisplayNameAttribute = implementation.DisplayNameAttribute
	}

	if ldapImplementationShouldSetStr(config.MailAttribute, implementation.MailAttribute) {
		config.MailAttribute = implementation.MailAttribute
	}

	if ldapImplementationShouldSetStr(config.GroupsFilter, implementation.GroupsFilter) {
		config.GroupsFilter = implementation.GroupsFilter
	}

	if ldapImplementationShouldSetStr(config.GroupNameAttribute, implementation.GroupNameAttribute) {
		config.GroupNameAttribute = implementation.GroupNameAttribute
	}
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
