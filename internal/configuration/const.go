package configuration

import (
	"errors"
)

const (
	constWindows = "windows"

	constEnvPrefix    = "AUTHELIA__"
	constEnvPrefixAlt = "AUTHELIA_"
	constSecretSuffix = "_FILE"

	constDelimiter    = "."
	constDelimiterEnv = "_"
)

const (
	errFmtSecretAlreadyDefined = "secrets: error loading secret into key '%s': it's already defined in other " +
		"configuration sources"
	errFmtSecretIOIssue         = "secrets: error loading secret path %s into key '%s': %v"
	errFmtGenerateConfiguration = "error occurred generating Configuration: %+v"
)

var secretSuffixes = []string{"key", "secret", "password", "token"}
var errInvalidPrefix = errors.New("invalid prefix")
var errNoSources = errors.New("no sources provided")
