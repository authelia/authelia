package configuration

import "errors"

const (
	windows = "windows"
)

const (
	errFmtSecretAlreadyDefined = "error loading secret into key '%s': it's already defined in the config file" //nolint:gosec
	errFmtSecretIOIssue        = "error loading secret file %s into key '%s': %v"                              //nolint:gosec
)

var errSecretOneOrMoreErrors = errors.New("one or more errors occurred during loading secrets")

// AUTHELIA_PORT_ and AUTHELIA_SERVICE_ are added to k8s pods in some situations.
var ignoredEnvPrefixes = []string{"AUTHELIA_PORT_", "AUTHELIA_SERVICE_"}
var ignoredKeys = []string{"AUTHELIA_PORT"}
