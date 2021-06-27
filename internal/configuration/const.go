package configuration

import "errors"

const (
	windows = "windows"

	envPrefixAlt = "AUTHELIA_"
	envPrefix    = "AUTHELIA__"
	secretSuffix = "_FILE"

	delimiter    = "."
	delimiterEnv = "_"
)

const (
	errFmtSecretAlreadyDefined  = "error loading secret into key '%s': it's already defined in the config files" //nolint:gosec
	errFmtSecretIOIssue         = "error loading secret file %s into key '%s': %v"                               //nolint:gosec
	errFmtGenerateConfiguration = "error occurred generating configuration: %+v"
)

var secretSuffixes = []string{"key", "secret", "password", "token"}
var errInvalidPrefix = errors.New("invalid prefix")
