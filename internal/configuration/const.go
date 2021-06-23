package configuration

import "errors"

const (
	windows = "windows"

	envPrefixAlt = "AUTHELIA_"
	envPrefix    = "AUTHELIA__"

	delimiter    = "."
	delimiterEnv = "_"

	secretPrefix    = "secret."
	secretSuffix    = ".file"
	secretSuffixEnv = "_FILE"
)

const (
	errFmtSecretAlreadyDefined = "error loading secret into key '%s': it's already defined in the config files" //nolint:gosec
	errFmtSecretIOIssue        = "error loading secret file %s into key '%s': %v"                               //nolint:gosec
)

var secretSuffixes = []string{"key", "secret", "password", "token"}
var errInvalidPrefix = errors.New("invalid prefix")
