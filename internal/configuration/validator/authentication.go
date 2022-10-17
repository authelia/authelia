package validator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/go-crypt/crypt"
)

// ValidateAuthenticationBackend validates and updates the authentication backend configuration.
func ValidateAuthenticationBackend(config *schema.AuthenticationBackend, validator *schema.StructValidator) {
	if config.LDAP == nil && config.File == nil {
		validator.Push(fmt.Errorf(errFmtAuthBackendNotConfigured))
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

	if config.LDAP != nil && config.File != nil {
		validator.Push(fmt.Errorf(errFmtAuthBackendMultipleConfigured))
	}

	if config.File != nil {
		validateFileAuthenticationBackend(config.File, validator)
	}

	if config.LDAP != nil {
		validateLDAPAuthenticationBackend(config, validator)
	}
}

// validateFileAuthenticationBackend validates and updates the file authentication backend configuration.
func validateFileAuthenticationBackend(config *schema.FileAuthenticationBackend, validator *schema.StructValidator) {
	if config.Path == "" {
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPathNotConfigured))
	}

	ValidatePasswordConfiguration(&config.Password, validator)
}

// ValidatePasswordConfiguration validates the file auth backend password configuration.
func ValidatePasswordConfiguration(config *schema.Password, validator *schema.StructValidator) {
	validateFileAuthenticationBackendPasswordConfigLegacy(config)

	switch {
	case config.Algorithm == "":
		config.Algorithm = schema.DefaultPasswordConfig.Algorithm
	case utils.IsStringInSlice(config.Algorithm, validHashAlgorithms):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordUnknownAlg, config.Algorithm, strings.Join(validHashAlgorithms, "', '")))
	}

	validateFileAuthenticationBackendPasswordConfigArgon2(config, validator)
	validateFileAuthenticationBackendPasswordConfigSHA2Crypt(config, validator)
	validateFileAuthenticationBackendPasswordConfigPBKDF2(config, validator)
	validateFileAuthenticationBackendPasswordConfigBCrypt(config, validator)
	validateFileAuthenticationBackendPasswordConfigSCrypt(config, validator)
}

func validateFileAuthenticationBackendPasswordConfigArgon2(config *schema.Password, validator *schema.StructValidator) {
	switch {
	case config.Argon2.Variant == "":
		config.Argon2.Variant = schema.DefaultPasswordConfig.Argon2.Variant
	case utils.IsStringInSlice(config.Argon2.Variant, validArgon2Variants):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidVariant, hashArgon2, config.Argon2.Variant, strings.Join(validArgon2Variants, "', '")))
	}

	switch {
	case config.Argon2.Iterations == 0:
		config.Argon2.Iterations = schema.DefaultPasswordConfig.Argon2.Iterations
	case config.Argon2.Iterations < crypt.Argon2IterationsMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "iterations", config.Argon2.Iterations, crypt.Argon2IterationsMin))
	case config.Argon2.Iterations > crypt.Argon2IterationsMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashArgon2, "iterations", config.Argon2.Iterations, crypt.Argon2IterationsMax))
	}

	switch {
	case config.Argon2.Parallelism == 0:
		config.Argon2.Parallelism = schema.DefaultPasswordConfig.Argon2.Parallelism
	case config.Argon2.Parallelism < crypt.Argon2ParallelismMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "parallelism", config.Argon2.Parallelism, crypt.Argon2ParallelismMin))
	case config.Argon2.Parallelism > crypt.Argon2ParallelismMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashArgon2, "parallelism", config.Argon2.Parallelism, crypt.Argon2ParallelismMax))
	}

	switch {
	case config.Argon2.Memory == 0:
		config.Argon2.Memory = schema.DefaultPasswordConfig.Argon2.Memory
	case config.Argon2.Memory < 0:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "memory", config.Argon2.Parallelism, 1))
	case config.Argon2.Memory < (crypt.Argon2MemoryMinParallelismMultiplier * config.Argon2.Parallelism):
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordArgon2MemoryTooLow, config.Argon2.Memory, config.Argon2.Parallelism*crypt.Argon2MemoryMinParallelismMultiplier, config.Argon2.Parallelism, crypt.Argon2MemoryMinParallelismMultiplier))
	}

	switch {
	case config.Argon2.KeyLength == 0:
		config.Argon2.KeyLength = schema.DefaultPasswordConfig.Argon2.KeyLength
	case config.Argon2.KeyLength < crypt.Argon2KeySizeMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "key_length", config.Argon2.KeyLength, crypt.Argon2KeySizeMin))
	case config.Argon2.KeyLength > crypt.Argon2KeySizeMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashArgon2, "key_length", config.Argon2.KeyLength, crypt.Argon2KeySizeMax))
	}

	switch {
	case config.Argon2.SaltLength == 0:
		config.Argon2.SaltLength = schema.DefaultPasswordConfig.Argon2.SaltLength
	case config.Argon2.SaltLength < crypt.Argon2SaltSizeMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "salt_length", config.Argon2.SaltLength, crypt.Argon2SaltSizeMin))
	case config.Argon2.SaltLength > crypt.Argon2SaltSizeMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashArgon2, "salt_length", config.Argon2.SaltLength, crypt.Argon2SaltSizeMax))
	}
}

func validateFileAuthenticationBackendPasswordConfigSHA2Crypt(config *schema.Password, validator *schema.StructValidator) {
	switch {
	case config.SHA2Crypt.Variant == "":
		config.SHA2Crypt.Variant = schema.DefaultPasswordConfig.SHA2Crypt.Variant
	case utils.IsStringInSlice(config.SHA2Crypt.Variant, validSHA2CryptVariants):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidVariant, hashSHA2Crypt, config.SHA2Crypt.Variant, strings.Join(validSHA2CryptVariants, "', '")))
	}

	switch {
	case config.SHA2Crypt.Iterations == 0:
		config.SHA2Crypt.Iterations = schema.DefaultPasswordConfig.SHA2Crypt.Iterations
	case config.SHA2Crypt.Iterations < crypt.SHA2CryptIterationsMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashSHA2Crypt, "iterations", config.SHA2Crypt.Iterations, crypt.SHA2CryptIterationsMin))
	case config.SHA2Crypt.Iterations > crypt.SHA2CryptIterationsMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashSHA2Crypt, "iterations", config.SHA2Crypt.Iterations, crypt.SHA2CryptIterationsMax))
	}

	switch {
	case config.SHA2Crypt.SaltLength == 0:
		config.SHA2Crypt.SaltLength = schema.DefaultPasswordConfig.SHA2Crypt.SaltLength
	case config.SHA2Crypt.SaltLength < crypt.SHA2CryptSaltSizeMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashSHA2Crypt, "salt_length", config.SHA2Crypt.SaltLength, crypt.SHA2CryptSaltSizeMin))
	case config.SHA2Crypt.SaltLength > crypt.SHA2CryptSaltSizeMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashSHA2Crypt, "salt_length", config.SHA2Crypt.SaltLength, crypt.SHA2CryptSaltSizeMax))
	}
}

func validateFileAuthenticationBackendPasswordConfigPBKDF2(config *schema.Password, validator *schema.StructValidator) {
	switch {
	case config.PBKDF2.Variant == "":
		config.PBKDF2.Variant = schema.DefaultPasswordConfig.PBKDF2.Variant
	case utils.IsStringInSlice(config.PBKDF2.Variant, validPBKDF2Variants):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidVariant, hashPBKDF2, config.PBKDF2.Variant, strings.Join(validPBKDF2Variants, "', '")))
	}

	switch {
	case config.PBKDF2.Iterations == 0:
		config.PBKDF2.Iterations = schema.DefaultPasswordConfig.PBKDF2.Iterations
	case config.PBKDF2.Iterations < crypt.PBKDF2IterationsMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashPBKDF2, "iterations", config.PBKDF2.Iterations, crypt.PBKDF2IterationsMin))
	case config.PBKDF2.Iterations > crypt.PBKDF2IterationsMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashPBKDF2, "iterations", config.PBKDF2.Iterations, crypt.PBKDF2IterationsMax))
	}

	switch {
	case config.PBKDF2.SaltLength == 0:
		config.PBKDF2.SaltLength = schema.DefaultPasswordConfig.PBKDF2.SaltLength
	case config.PBKDF2.SaltLength < crypt.PBKDF2SaltSizeMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashPBKDF2, "salt_length", config.PBKDF2.SaltLength, crypt.PBKDF2SaltSizeMin))
	case config.PBKDF2.SaltLength > crypt.PBKDF2SaltSizeMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashPBKDF2, "salt_length", config.PBKDF2.SaltLength, crypt.PBKDF2SaltSizeMax))
	}
}

func validateFileAuthenticationBackendPasswordConfigBCrypt(config *schema.Password, validator *schema.StructValidator) {
	switch {
	case config.BCrypt.Variant == "":
		config.BCrypt.Variant = schema.DefaultPasswordConfig.BCrypt.Variant
	case utils.IsStringInSlice(config.BCrypt.Variant, validBCryptVariants):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidVariant, hashBCrypt, config.BCrypt.Variant, strings.Join(validBCryptVariants, "', '")))
	}

	switch {
	case config.BCrypt.Cost == 0:
		config.BCrypt.Cost = schema.DefaultPasswordConfig.BCrypt.Cost
	case config.BCrypt.Cost < crypt.BcryptCostMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashBCrypt, "cost", config.BCrypt.Cost, crypt.BcryptCostMin))
	case config.BCrypt.Cost > crypt.BcryptCostMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashBCrypt, "cost", config.BCrypt.Cost, crypt.BcryptCostMax))
	}
}

func validateFileAuthenticationBackendPasswordConfigSCrypt(config *schema.Password, validator *schema.StructValidator) {
	switch {
	case config.SCrypt.Iterations == 0:
		config.SCrypt.Iterations = schema.DefaultPasswordConfig.SCrypt.Iterations
	case config.SCrypt.Iterations < crypt.ScryptIterationsMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashSCrypt, "iterations", config.SCrypt.Iterations, crypt.ScryptIterationsMin))
	}

	switch {
	case config.SCrypt.BlockSize == 0:
		config.SCrypt.BlockSize = schema.DefaultPasswordConfig.SCrypt.BlockSize
	case config.SCrypt.BlockSize < crypt.ScryptBlockSizeMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashSCrypt, "block_size", config.SCrypt.BlockSize, crypt.ScryptBlockSizeMin))
	case config.SCrypt.BlockSize > crypt.ScryptBlockSizeMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashSCrypt, "block_size", config.SCrypt.BlockSize, crypt.ScryptBlockSizeMax))
	}

	switch {
	case config.SCrypt.Parallelism == 0:
		config.SCrypt.Parallelism = schema.DefaultPasswordConfig.SCrypt.Parallelism
	case config.SCrypt.Parallelism < crypt.ScryptParallelismMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashSCrypt, "parallelism", config.SCrypt.Parallelism, crypt.ScryptParallelismMin))
	}

	switch {
	case config.SCrypt.KeyLength == 0:
		config.SCrypt.KeyLength = schema.DefaultPasswordConfig.SCrypt.KeyLength
	case config.SCrypt.KeyLength < crypt.ScryptKeySizeMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashSCrypt, "key_length", config.SCrypt.KeyLength, crypt.ScryptKeySizeMin))
	case config.SCrypt.KeyLength > crypt.ScryptKeySizeMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashSCrypt, "key_length", config.SCrypt.KeyLength, crypt.ScryptKeySizeMax))
	}

	switch {
	case config.SCrypt.SaltLength == 0:
		config.SCrypt.SaltLength = schema.DefaultPasswordConfig.SCrypt.SaltLength
	case config.SCrypt.SaltLength < crypt.ScryptSaltSizeMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashSCrypt, "salt_length", config.SCrypt.SaltLength, crypt.ScryptSaltSizeMin))
	case config.SCrypt.SaltLength > crypt.ScryptSaltSizeMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashSCrypt, "salt_length", config.SCrypt.SaltLength, crypt.ScryptSaltSizeMax))
	}
}

func validateFileAuthenticationBackendPasswordConfigLegacy(config *schema.Password) {
	switch config.Algorithm {
	case hashLegacySHA512:
		config.Algorithm = hashSHA2Crypt

		if config.SHA2Crypt.Variant == "" {
			config.SHA2Crypt.Variant = schema.DefaultPasswordConfig.SHA2Crypt.Variant
		}

		if config.Iterations > 0 && config.SHA2Crypt.Iterations == 0 {
			config.SHA2Crypt.Iterations = config.Iterations
		}

		if config.SaltLength > 0 && config.SHA2Crypt.SaltLength == 0 {
			if config.SaltLength > 16 {
				config.SHA2Crypt.SaltLength = 16
			} else {
				config.SHA2Crypt.SaltLength = config.SaltLength
			}
		}
	case hashLegacyArgon2id:
		config.Algorithm = hashArgon2

		if config.Argon2.Variant == "" {
			config.Argon2.Variant = schema.DefaultPasswordConfig.Argon2.Variant
		}

		if config.Iterations > 0 && config.Argon2.Memory == 0 {
			config.Argon2.Iterations = config.Iterations
		}

		if config.Memory > 0 && config.Argon2.Memory == 0 {
			config.Argon2.Memory = config.Memory * 1024
		}

		if config.Parallelism > 0 && config.Argon2.Parallelism == 0 {
			config.Argon2.Parallelism = config.Parallelism
		}

		if config.KeyLength > 0 && config.Argon2.KeyLength == 0 {
			config.Argon2.KeyLength = config.KeyLength
		}

		if config.SaltLength > 0 && config.Argon2.SaltLength == 0 {
			config.Argon2.SaltLength = config.SaltLength
		}
	}
}

func validateLDAPAuthenticationBackend(config *schema.AuthenticationBackend, validator *schema.StructValidator) {
	if config.LDAP.Implementation == "" {
		config.LDAP.Implementation = schema.LDAPImplementationCustom
	}

	var implementation *schema.LDAPAuthenticationBackend

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

func setDefaultImplementationLDAPAuthenticationBackendProfileMisc(config *schema.LDAPAuthenticationBackend, implementation *schema.LDAPAuthenticationBackend) {
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

func setDefaultImplementationLDAPAuthenticationBackendProfileAttributes(config *schema.LDAPAuthenticationBackend, implementation *schema.LDAPAuthenticationBackend) {
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

func validateLDAPAuthenticationBackendURL(config *schema.LDAPAuthenticationBackend, validator *schema.StructValidator) {
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

func validateLDAPRequiredParameters(config *schema.AuthenticationBackend, validator *schema.StructValidator) {
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
