package configuration

import (
	"errors"
	"math"
	"time"
)

// DefaultEnvPrefix is the default environment prefix.
const DefaultEnvPrefix = "AUTHELIA_"

// DefaultEnvDelimiter is the default environment delimiter.
const DefaultEnvDelimiter = "_"

const (
	constSecretSuffix = "_FILE"

	constDelimiter = "."

	constWindows = "windows"

	extYML  = ".yml"
	extYAML = ".yaml"
)

const (
	filterField     = "filter"
	filterTemplate  = "template"
	filterExpandEnv = "expand-env"
)

var (
	errNoValidator = errors.New("no validator provided")
	errNoSources   = errors.New("no sources provided")

	errDecodeNonPtrMustHaveValue = errors.New("must have a non-empty value")
)

const (
	errFmtSecretAlreadyDefined = "secrets: error loading secret into key '%s': it's already defined in other " +
		"configuration sources"
	errFmtSecretOSError         = "secrets: error loading secret path %s into key '%s': %w"
	errFmtSecretOSPermission    = "secrets: error loading secret path %s into key '%s': file permission error occurred: %w"
	errFmtSecretOSNotExist      = "secrets: error loading secret path %s into key '%s': file does not exist error occurred: %w"
	errFmtGenerateConfiguration = "error occurred generating configuration: %+v"

	errFmtDecodeHookCouldNotParse           = "could not decode '%s' to a %s%s: %w"
	errFmtDecodeHookCouldNotParseBasic      = "could not decode to a %s%s: %w"
	errFmtDecodeHookCouldNotParseEmptyValue = "could not decode an empty value to a %s%s: %w"

	errFmtSuffixAutoRemappedKey = "you are not required to make any changes as this has been automatically mapped for you, but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in %s"

	errFmtMultiRemappedKeys          = "configuration keys %s are deprecated in %s and has been replaced by '%s' in the format of '%s': you are not required to make any changes as this has been automatically mapped for you to the value '%s', but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in %s"
	errFmtMultiKeyMappingExists      = "error occurred performing deprecation mapping for keys %s to new key %s: the new key already exists with value '%s' but the deprecated keys and the new key can't both be configured"
	errFmtMultiKeyMappingPortConvert = "error occurred performing deprecation mapping for keys %s to new key %s: %w"

	errFmtAutoMapKey           = "configuration key '%s' is deprecated in %s and has been replaced by '%s': " + errFmtSuffixAutoRemappedKey
	errFmtNoAutoMapKeyNoNewKey = "configuration key '%s' is deprecated in %s and has been removed': you are not required to make any configuration changes right now but you may be required to in %s"
	errFmtAutoMapKeyExisting   = "configuration key '%s' is deprecated in %s and has been replaced by '%s': this has not been automatically mapped for you because the replacement key also exists and you will need to adjust your configuration to remove this message"
)

const (
	durationMax = time.Duration(math.MaxInt64)
)

const (
	keyServerHost          = "server.host"
	keyServerPort          = "server.port"
	keyServerPath          = "server.path"
	keyStorageMySQLHost    = "storage.mysql.host"
	keyStorageMySQLPort    = "storage.mysql.port"
	keyStoragePostgresHost = "storage.postgres.host"
	keyStoragePostgresPort = "storage.postgres.port"
)

// IMPORTANT: There is an uppercase copy of this in github.com/authelia/authelia/internal/templates named
// envSecretSuffixes.
// Make sure you update these at the same time.
var (
	secretSuffix          = []string{"key", "secret", "password", "token", "certificate_chain"}
	secretExclusionPrefix = []string{"identity_providers.oidc.lifespans."}
	secretExclusionExact  = []string{"server.tls.key", "authentication_backend.disable_reset_password", "tls_key"}
)

var (
	mapDefaults = map[string]any{
		"webauthn.metadata.validate_trust_anchor":             true,
		"webauthn.metadata.validate_entry":                    true,
		"webauthn.metadata.validate_entry_permit_zero_aaguid": false,
		"webauthn.metadata.validate_status":                   true,
	}
)
