package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/go-crypt/crypt/algorithm/argon2"
	"github.com/go-crypt/crypt/algorithm/bcrypt"
	"github.com/go-crypt/crypt/algorithm/pbkdf2"
	"github.com/go-crypt/crypt/algorithm/scrypt"
	"github.com/go-crypt/crypt/algorithm/shacrypt"
)

// ValidateAuthenticationBackend validates and updates the authentication backend configuration.
func ValidateAuthenticationBackend(config *schema.AuthenticationBackend, validator *schema.StructValidator) {
	if config.LDAP == nil && config.File == nil {
		validator.Push(errors.New(errFmtAuthBackendNotConfigured))
	}

	if !config.RefreshInterval.Valid() {
		if config.File != nil && config.File.Watch {
			config.RefreshInterval = schema.NewRefreshIntervalDurationAlways()
		} else {
			config.RefreshInterval = schema.NewRefreshIntervalDuration(schema.RefreshIntervalDefault)
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
		validator.Push(errors.New(errFmtAuthBackendMultipleConfigured))
	}

	if config.File != nil {
		validateFileAuthenticationBackend(config.File, validator)
	}

	if config.LDAP != nil {
		validateLDAPAuthenticationBackend(config, validator)
	}
}

// validateFileAuthenticationBackend validates and updates the file authentication backend configuration.
func validateFileAuthenticationBackend(config *schema.AuthenticationBackendFile, validator *schema.StructValidator) {
	if config.Path == "" {
		validator.Push(errors.New(errFmtFileAuthBackendPathNotConfigured))
	}

	for name, attr := range config.ExtraAttributes {
		switch attr.ValueType {
		case authentication.ValueTypeString, authentication.ValueTypeInteger, authentication.ValueTypeBoolean:
			break
		case "":
			validator.Push(fmt.Errorf(errFmtFileAuthBackendExtraAttributeValueTypeMissing, name))
		default:
			validator.Push(fmt.Errorf(errFmtFileAuthBackendExtraAttributeValueType, name, attr.ValueType))
		}

		if expression.IsReservedAttribute(name) {
			validator.Push(fmt.Errorf(errFmtFileAuthBackendExtraAttributeReserved, name, name))
		}
	}

	ValidatePasswordConfiguration(&config.Password, validator)
}

// ValidatePasswordConfiguration validates the file auth backend password configuration.
func ValidatePasswordConfiguration(config *schema.AuthenticationBackendFilePassword, validator *schema.StructValidator) {
	validateFileAuthenticationBackendPasswordConfigLegacy(config)

	switch {
	case config.Algorithm == "":
		config.Algorithm = schema.DefaultPasswordConfig.Algorithm
	case utils.IsStringInSlice(config.Algorithm, validHashAlgorithms):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordUnknownAlg, utils.StringJoinOr(validHashAlgorithms), config.Algorithm))
	}

	validateFileAuthenticationBackendPasswordConfigArgon2(config, validator)
	validateFileAuthenticationBackendPasswordConfigSHA2Crypt(config, validator)
	validateFileAuthenticationBackendPasswordConfigPBKDF2(config, validator)
	validateFileAuthenticationBackendPasswordConfigBcrypt(config, validator)
	validateFileAuthenticationBackendPasswordConfigScrypt(config, validator)
}

//nolint:gocyclo // Function is well formed.
func validateFileAuthenticationBackendPasswordConfigArgon2(config *schema.AuthenticationBackendFilePassword, validator *schema.StructValidator) {
	switch {
	case config.Argon2.Variant == "":
		config.Argon2.Variant = schema.DefaultPasswordConfig.Argon2.Variant
	case utils.IsStringInSlice(config.Argon2.Variant, validArgon2Variants):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidVariant, hashArgon2, utils.StringJoinOr(validArgon2Variants), config.Argon2.Variant))
	}

	switch {
	case config.Argon2.Iterations == 0:
		config.Argon2.Iterations = schema.DefaultPasswordConfig.Argon2.Iterations
	case config.Argon2.Iterations < argon2.IterationsMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "iterations", config.Argon2.Iterations, argon2.IterationsMin))
	case config.Argon2.Iterations > argon2.IterationsMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashArgon2, "iterations", config.Argon2.Iterations, argon2.IterationsMax))
	}

	switch {
	case config.Argon2.Parallelism == 0:
		config.Argon2.Parallelism = schema.DefaultPasswordConfig.Argon2.Parallelism
	case config.Argon2.Parallelism < argon2.ParallelismMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "parallelism", config.Argon2.Parallelism, argon2.ParallelismMin))
	case config.Argon2.Parallelism > argon2.ParallelismMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashArgon2, "parallelism", config.Argon2.Parallelism, argon2.ParallelismMax))
	}

	switch {
	case config.Argon2.Memory == 0:
		config.Argon2.Memory = schema.DefaultPasswordConfig.Argon2.Memory
	case config.Argon2.Memory < argon2.MemoryMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "memory", config.Argon2.Memory, argon2.MemoryMin))
	case uint64(config.Argon2.Memory) > uint64(argon2.MemoryMax): //nolint:gosec // Validated at runtime.
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashArgon2, "memory", config.Argon2.Memory, argon2.MemoryMax))
	case config.Argon2.Memory < (config.Argon2.Parallelism * argon2.MemoryMinParallelismMultiplier):
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordArgon2MemoryTooLow, config.Argon2.Memory, config.Argon2.Parallelism*argon2.MemoryMinParallelismMultiplier, config.Argon2.Parallelism, argon2.MemoryMinParallelismMultiplier))
	}

	switch {
	case config.Argon2.KeyLength == 0:
		config.Argon2.KeyLength = schema.DefaultPasswordConfig.Argon2.KeyLength
	case config.Argon2.KeyLength < argon2.KeyLengthMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "key_length", config.Argon2.KeyLength, argon2.KeyLengthMin))
	case config.Argon2.KeyLength > argon2.KeyLengthMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashArgon2, "key_length", config.Argon2.KeyLength, argon2.KeyLengthMax))
	}

	switch {
	case config.Argon2.SaltLength == 0:
		config.Argon2.SaltLength = schema.DefaultPasswordConfig.Argon2.SaltLength
	case config.Argon2.SaltLength < argon2.SaltLengthMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashArgon2, "salt_length", config.Argon2.SaltLength, argon2.SaltLengthMin))
	case config.Argon2.SaltLength > argon2.SaltLengthMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashArgon2, "salt_length", config.Argon2.SaltLength, argon2.SaltLengthMax))
	}
}

func validateFileAuthenticationBackendPasswordConfigSHA2Crypt(config *schema.AuthenticationBackendFilePassword, validator *schema.StructValidator) {
	switch {
	case config.SHA2Crypt.Variant == "":
		config.SHA2Crypt.Variant = schema.DefaultPasswordConfig.SHA2Crypt.Variant
	case utils.IsStringInSlice(config.SHA2Crypt.Variant, validSHA2CryptVariants):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidVariant, hashSHA2Crypt, utils.StringJoinOr(validSHA2CryptVariants), config.SHA2Crypt.Variant))
	}

	switch {
	case config.SHA2Crypt.Iterations == 0:
		config.SHA2Crypt.Iterations = schema.DefaultPasswordConfig.SHA2Crypt.Iterations
	case config.SHA2Crypt.Iterations < shacrypt.IterationsMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashSHA2Crypt, "iterations", config.SHA2Crypt.Iterations, shacrypt.IterationsMin))
	case config.SHA2Crypt.Iterations > shacrypt.IterationsMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashSHA2Crypt, "iterations", config.SHA2Crypt.Iterations, shacrypt.IterationsMax))
	}

	switch {
	case config.SHA2Crypt.SaltLength == 0:
		config.SHA2Crypt.SaltLength = schema.DefaultPasswordConfig.SHA2Crypt.SaltLength
	case config.SHA2Crypt.SaltLength < shacrypt.SaltLengthMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashSHA2Crypt, "salt_length", config.SHA2Crypt.SaltLength, shacrypt.SaltLengthMin))
	case config.SHA2Crypt.SaltLength > shacrypt.SaltLengthMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashSHA2Crypt, "salt_length", config.SHA2Crypt.SaltLength, shacrypt.SaltLengthMax))
	}
}

func validateFileAuthenticationBackendPasswordConfigPBKDF2(config *schema.AuthenticationBackendFilePassword, validator *schema.StructValidator) {
	switch {
	case config.PBKDF2.Variant == "":
		config.PBKDF2.Variant = schema.DefaultPasswordConfig.PBKDF2.Variant
	case utils.IsStringInSlice(config.PBKDF2.Variant, validPBKDF2Variants):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidVariant, hashPBKDF2, utils.StringJoinOr(validPBKDF2Variants), config.PBKDF2.Variant))
	}

	switch {
	case config.PBKDF2.Iterations == 0:
		config.PBKDF2.Iterations = schema.PBKDF2VariantDefaultIterations(config.PBKDF2.Variant)
	case config.PBKDF2.Iterations < pbkdf2.IterationsMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashPBKDF2, "iterations", config.PBKDF2.Iterations, pbkdf2.IterationsMin))
	case config.PBKDF2.Iterations > pbkdf2.IterationsMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashPBKDF2, "iterations", config.PBKDF2.Iterations, pbkdf2.IterationsMax))
	}

	switch {
	case config.PBKDF2.SaltLength == 0:
		config.PBKDF2.SaltLength = schema.DefaultPasswordConfig.PBKDF2.SaltLength
	case config.PBKDF2.SaltLength < pbkdf2.SaltLengthMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashPBKDF2, "salt_length", config.PBKDF2.SaltLength, pbkdf2.SaltLengthMin))
	case config.PBKDF2.SaltLength > pbkdf2.SaltLengthMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashPBKDF2, "salt_length", config.PBKDF2.SaltLength, pbkdf2.SaltLengthMax))
	}
}

func validateFileAuthenticationBackendPasswordConfigBcrypt(config *schema.AuthenticationBackendFilePassword, validator *schema.StructValidator) {
	switch {
	case config.Bcrypt.Variant == "":
		config.Bcrypt.Variant = schema.DefaultPasswordConfig.Bcrypt.Variant
	case utils.IsStringInSlice(config.Bcrypt.Variant, validBcryptVariants):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidVariant, hashBcrypt, utils.StringJoinOr(validBcryptVariants), config.Bcrypt.Variant))
	}

	switch {
	case config.Bcrypt.Cost == 0:
		config.Bcrypt.Cost = schema.DefaultPasswordConfig.Bcrypt.Cost
	case config.Bcrypt.Cost < bcrypt.IterationsMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashBcrypt, "cost", config.Bcrypt.Cost, bcrypt.IterationsMin))
	case config.Bcrypt.Cost > bcrypt.IterationsMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashBcrypt, "cost", config.Bcrypt.Cost, bcrypt.IterationsMax))
	}
}

//nolint:gocyclo
func validateFileAuthenticationBackendPasswordConfigScrypt(config *schema.AuthenticationBackendFilePassword, validator *schema.StructValidator) {
	switch {
	case config.Scrypt.Variant == "":
		config.Scrypt.Variant = schema.DefaultPasswordConfig.Scrypt.Variant
	case utils.IsStringInSlice(config.Scrypt.Variant, validScryptVariants):
		break
	default:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordInvalidVariant, hashScrypt, utils.StringJoinOr(validScryptVariants), config.Scrypt.Variant))
	}

	switch {
	case config.Scrypt.Iterations == 0:
		config.Scrypt.Iterations = schema.DefaultPasswordConfig.Scrypt.Iterations
	case config.Scrypt.Iterations < scrypt.IterationsMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashScrypt, "iterations", config.Scrypt.Iterations, scrypt.IterationsMin))
	case config.Scrypt.Iterations > scrypt.IterationsMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashScrypt, "iterations", config.Scrypt.Iterations, scrypt.IterationsMax))
	}

	switch {
	case config.Scrypt.BlockSize == 0:
		config.Scrypt.BlockSize = schema.DefaultPasswordConfig.Scrypt.BlockSize
	case config.Scrypt.BlockSize < scrypt.BlockSizeMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashScrypt, "block_size", config.Scrypt.BlockSize, scrypt.BlockSizeMin))
	case config.Scrypt.BlockSize > scrypt.BlockSizeMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashScrypt, "block_size", config.Scrypt.BlockSize, scrypt.BlockSizeMax))
	}

	switch {
	case config.Scrypt.Parallelism == 0:
		config.Scrypt.Parallelism = schema.DefaultPasswordConfig.Scrypt.Parallelism
	case config.Scrypt.Parallelism < scrypt.ParallelismMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashScrypt, "parallelism", config.Scrypt.Parallelism, scrypt.ParallelismMin))
	case config.Scrypt.Parallelism > scrypt.ParallelismMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashScrypt, "parallelism", config.Scrypt.Parallelism, scrypt.ParallelismMax))
	case config.Scrypt.Variant == "yescrypt" && config.Scrypt.Parallelism != 1:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionInvalid, hashScrypt, "parallelism", config.Scrypt.Parallelism, 1, "variant", config.Scrypt.Variant))
	}

	switch {
	case config.Scrypt.KeyLength == 0:
		config.Scrypt.KeyLength = schema.DefaultPasswordConfig.Scrypt.KeyLength
	case config.Scrypt.KeyLength < scrypt.KeyLengthMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashScrypt, "key_length", config.Scrypt.KeyLength, scrypt.KeyLengthMin))
	case config.Scrypt.KeyLength > scrypt.KeyLengthMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashScrypt, "key_length", config.Scrypt.KeyLength, scrypt.KeyLengthMax))
	}

	switch {
	case config.Scrypt.SaltLength == 0:
		config.Scrypt.SaltLength = schema.DefaultPasswordConfig.Scrypt.SaltLength
	case config.Scrypt.SaltLength < scrypt.SaltLengthMin:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooSmall, hashScrypt, "salt_length", config.Scrypt.SaltLength, scrypt.SaltLengthMin))
	case config.Scrypt.SaltLength > scrypt.SaltLengthMax:
		validator.Push(fmt.Errorf(errFmtFileAuthBackendPasswordOptionTooLarge, hashScrypt, "salt_length", config.Scrypt.SaltLength, scrypt.SaltLengthMax))
	}
}

//nolint:gocyclo,staticcheck // Function is clear enough and being used for deprecated functionality mapping.
func validateFileAuthenticationBackendPasswordConfigLegacy(config *schema.AuthenticationBackendFilePassword) {
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
			config.SHA2Crypt.SaltLength = min(config.SaltLength, 16)
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

	defaultTLS := validateLDAPAuthenticationBackendImplementation(config, validator)

	defaultTLS.ServerName = validateLDAPAuthenticationAddress(config.LDAP, validator)

	if config.LDAP.TLS == nil {
		config.LDAP.TLS = &schema.TLS{}
	}

	if err := ValidateTLSConfig(config.LDAP.TLS, defaultTLS); err != nil {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendTLSConfigInvalid, err))
	}

	if config.LDAP.Pooling.Enable {
		if config.LDAP.Pooling.Count < 1 {
			config.LDAP.Pooling.Count = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Count
		}

		if config.LDAP.Pooling.Retries < 1 {
			config.LDAP.Pooling.Retries = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Retries
		}

		if config.LDAP.Pooling.Timeout < 1 {
			config.LDAP.Pooling.Timeout = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Timeout
		}
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

	validateLDAPExtraAttributes(config, validator)
	validateLDAPRequiredParameters(config, validator)
}

func validateLDAPAuthenticationBackendImplementation(config *schema.AuthenticationBackend, validator *schema.StructValidator) *schema.TLS {
	var implementation *schema.AuthenticationBackendLDAP

	switch config.LDAP.Implementation {
	case schema.LDAPImplementationCustom:
		implementation = &schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom
	case schema.LDAPImplementationActiveDirectory:
		implementation = &schema.DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory
	case schema.LDAPImplementationRFC2307bis:
		implementation = &schema.DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis
	case schema.LDAPImplementationFreeIPA:
		implementation = &schema.DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA
	case schema.LDAPImplementationLLDAP:
		implementation = &schema.DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP
	case schema.LDAPImplementationGLAuth:
		implementation = &schema.DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth
	default:
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendOptionMustBeOneOf, "implementation", utils.StringJoinOr(validLDAPImplementations), config.LDAP.Implementation))
	}

	tlsconfig := &schema.TLS{}

	if implementation != nil {
		if config.LDAP.Timeout == 0 {
			config.LDAP.Timeout = implementation.Timeout
		}

		tlsconfig = &schema.TLS{
			MinimumVersion: implementation.TLS.MinimumVersion,
			MaximumVersion: implementation.TLS.MaximumVersion,
		}

		setDefaultImplementationLDAPAuthenticationBackendProfileAttributes(config.LDAP, implementation)
	}

	return tlsconfig
}

func ldapImplementationShouldSetStr(config, implementation string) bool {
	return config == "" && implementation != ""
}

func setDefaultImplementationLDAPAuthenticationBackendProfileAttributes(config *schema.AuthenticationBackendLDAP, implementation *schema.AuthenticationBackendLDAP) {
	if ldapImplementationShouldSetStr(config.AdditionalUsersDN, implementation.AdditionalUsersDN) {
		config.AdditionalUsersDN = implementation.AdditionalUsersDN
	}

	if ldapImplementationShouldSetStr(config.UsersFilter, implementation.UsersFilter) {
		config.UsersFilter = implementation.UsersFilter
	}

	if ldapImplementationShouldSetStr(config.AdditionalGroupsDN, implementation.AdditionalGroupsDN) {
		config.AdditionalGroupsDN = implementation.AdditionalGroupsDN
	}

	if ldapImplementationShouldSetStr(config.GroupsFilter, implementation.GroupsFilter) {
		config.GroupsFilter = implementation.GroupsFilter
	}

	if ldapImplementationShouldSetStr(config.GroupSearchMode, implementation.GroupSearchMode) {
		config.GroupSearchMode = implementation.GroupSearchMode
	}

	if ldapImplementationShouldSetStr(config.Attributes.DistinguishedName, implementation.Attributes.DistinguishedName) {
		config.Attributes.DistinguishedName = implementation.Attributes.DistinguishedName
	}

	if ldapImplementationShouldSetStr(config.Attributes.Username, implementation.Attributes.Username) {
		config.Attributes.Username = implementation.Attributes.Username
	}

	if ldapImplementationShouldSetStr(config.Attributes.DisplayName, implementation.Attributes.DisplayName) {
		config.Attributes.DisplayName = implementation.Attributes.DisplayName
	}

	if ldapImplementationShouldSetStr(config.Attributes.Mail, implementation.Attributes.Mail) {
		config.Attributes.Mail = implementation.Attributes.Mail
	}

	if ldapImplementationShouldSetStr(config.Attributes.MemberOf, implementation.Attributes.MemberOf) {
		config.Attributes.MemberOf = implementation.Attributes.MemberOf
	}

	if ldapImplementationShouldSetStr(config.Attributes.GroupName, implementation.Attributes.GroupName) {
		config.Attributes.GroupName = implementation.Attributes.GroupName
	}
}

func validateLDAPAuthenticationAddress(config *schema.AuthenticationBackendLDAP, validator *schema.StructValidator) (hostname string) {
	if config.Address == nil {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "address"))

		return
	}

	var (
		err error
	)
	if err = config.Address.ValidateLDAP(); err != nil {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendAddress, config.Address.String(), err))
	}

	return config.Address.Hostname()
}

func validateLDAPRequiredParameters(config *schema.AuthenticationBackend, validator *schema.StructValidator) {
	if config.LDAP.PermitUnauthenticatedBind {
		if config.LDAP.Password != "" {
			validator.Push(errors.New(errFmtLDAPAuthBackendUnauthenticatedBindWithPassword))
		}

		if !config.PasswordReset.Disable {
			validator.Push(errors.New(errFmtLDAPAuthBackendUnauthenticatedBindWithResetEnabled))
		}
	} else {
		if config.LDAP.User == "" {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "user"))
		}

		if config.LDAP.Password == "" {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendMissingOption, "password"))
		}
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

	validateLDAPGroupFilter(config, validator)
}

func validateLDAPExtraAttributes(config *schema.AuthenticationBackend, validator *schema.StructValidator) {
	for name, attr := range config.LDAP.Attributes.Extra {
		switch attr.ValueType {
		case authentication.ValueTypeString, authentication.ValueTypeInteger, authentication.ValueTypeBoolean:
			break
		case "":
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendExtraAttributeValueTypeMissing, name))
		default:
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendExtraAttributeValueType, name, attr.ValueType))
		}

		attribute := name

		if attr.Name != "" {
			attribute = attr.Name
		}

		if expression.IsReservedAttribute(attribute) {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendExtraAttributeReserved, name, attribute))
		}
	}
}

func validateLDAPGroupFilter(config *schema.AuthenticationBackend, validator *schema.StructValidator) {
	if config.LDAP.GroupSearchMode == "" {
		config.LDAP.GroupSearchMode = schema.LDAPGroupSearchModeFilter
	}

	if !utils.IsStringInSlice(config.LDAP.GroupSearchMode, validLDAPGroupSearchModes) {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendOptionMustBeOneOf, "group_search_mode", utils.StringJoinOr(validLDAPGroupSearchModes), config.LDAP.GroupSearchMode))
	}

	pMemberOfDN, pMemberOfRDN := strings.Contains(config.LDAP.GroupsFilter, "{memberof:dn}"), strings.Contains(config.LDAP.GroupsFilter, "{memberof:rdn}")

	if config.LDAP.GroupSearchMode == schema.LDAPGroupSearchModeMemberOf {
		if !pMemberOfDN && !pMemberOfRDN {
			validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterMissingPlaceholderGroupSearchMode, "groups_filter", utils.StringJoinOr([]string{"{memberof:rdn}", "{memberof:dn}"}), config.LDAP.GroupSearchMode))
		}
	}

	if pMemberOfDN && config.LDAP.Attributes.DistinguishedName == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterMissingAttribute, "distinguished_name", utils.StringJoinOr([]string{"{memberof:dn}"})))
	}

	if (pMemberOfDN || pMemberOfRDN) && config.LDAP.Attributes.MemberOf == "" {
		validator.Push(fmt.Errorf(errFmtLDAPAuthBackendFilterMissingAttribute, "member_of", utils.StringJoinOr([]string{"{memberof:rdn}", "{memberof:dn}"})))
	}
}
