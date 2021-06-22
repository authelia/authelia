package configuration

import "errors"

const (
	errFmtSecretAlreadyDefined = "error loading secret into key '%s': it's already defined in the config file" //nolint:gosec
	errFmtSecretIOIssue        = "error loading secret file %s into key '%s': %v"                              //nolint:gosec
)

var errSecretOneOrMoreErrors = errors.New("one or more errors occurred during loading secrets")
