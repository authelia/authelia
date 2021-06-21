package configuration

import "errors"

const (
	errFmtSecretAlreadyDefined = "error loading secret into key '%s': it's already defined in the config file"
	errFmtSecretIOIssue        = "error loading secret file %s into key '%s': %v"
)

var errSecretOneOrMoreErrors = errors.New("one or more errors occurred during loading secrets")
