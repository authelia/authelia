package validator

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// validateFileExists checks whether a file exist.
func validateFileExists(path string, validator *schema.StructValidator, errTemplate string) {
	exist, err := utils.FileExists(path)
	if err != nil {
		validator.Push(fmt.Errorf("tls: unable to check if file %s exists: %s", path, err))
	}

	if !exist {
		validator.Push(fmt.Errorf(errTemplate, path))
	}
}

// ValidateServerTLS checks a server TLS configuration is correct.
func ValidateServerTLS(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.TLS.Key != "" && config.Server.TLS.Certificate == "" {
		validator.Push(fmt.Errorf(errFmtServerTLSCert))
	} else if config.Server.TLS.Key == "" && config.Server.TLS.Certificate != "" {
		validator.Push(fmt.Errorf(errFmtServerTLSKey))
	}

	if config.Server.TLS.Key != "" {
		validateFileExists(config.Server.TLS.Key, validator, errFmtServerTLSKeyFileDoesNotExist)
	}

	if config.Server.TLS.Certificate != "" {
		validateFileExists(config.Server.TLS.Certificate, validator, errFmtServerTLSCertFileDoesNotExist)
	}

	if config.Server.TLS.Key == "" && config.Server.TLS.Certificate == "" &&
		len(config.Server.TLS.ClientCertificates) > 0 {
		validator.Push(fmt.Errorf(errFmtServerTLSClientAuthNoAuth))
	}

	for _, clientCertPath := range config.Server.TLS.ClientCertificates {
		validateFileExists(clientCertPath, validator, errFmtServerTLSClientAuthCertFileDoesNotExist)
	}
}

// ValidateServer checks a server configuration is correct.
func ValidateServer(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.Host == "" {
		config.Server.Host = schema.DefaultServerConfiguration.Host
	}

	if config.Server.Port == 0 {
		config.Server.Port = schema.DefaultServerConfiguration.Port
	}

	ValidateServerTLS(config, validator)

	switch {
	case strings.Contains(config.Server.Path, "/"):
		validator.Push(fmt.Errorf(errFmtServerPathNoForwardSlashes))
	case !utils.IsStringAlphaNumeric(config.Server.Path):
		validator.Push(fmt.Errorf(errFmtServerPathAlphaNum))
	case config.Server.Path == "": // Don't do anything if it's blank.
		break
	default:
		config.Server.Path = path.Clean("/" + config.Server.Path)
	}

	if config.Server.Buffers.Read <= 0 {
		config.Server.Buffers.Read = schema.DefaultServerConfiguration.Buffers.Read
	}

	if config.Server.Buffers.Write <= 0 {
		config.Server.Buffers.Write = schema.DefaultServerConfiguration.Buffers.Write
	}

	if config.Server.Timeouts.Read <= 0 {
		config.Server.Timeouts.Read = schema.DefaultServerConfiguration.Timeouts.Read
	}

	if config.Server.Timeouts.Write <= 0 {
		config.Server.Timeouts.Write = schema.DefaultServerConfiguration.Timeouts.Write
	}

	if config.Server.Timeouts.Idle <= 0 {
		config.Server.Timeouts.Idle = schema.DefaultServerConfiguration.Timeouts.Idle
	}

	ValidateServerEndpoints(config, validator)
}

// ValidateServerEndpoints configures the default endpoints and checks the configuration of custom endpoints.
//
//nolint:gocyclo
func ValidateServerEndpoints(config *schema.Configuration, validator *schema.StructValidator) {
	// TODO: log pprof/expvars.
	if len(config.Server.Endpoints.Authz) == 0 {
		config.Server.Endpoints.Authz = schema.DefaultServerConfiguration.Endpoints.Authz

		return
	}

	authzs := make([]string, 0, len(config.Server.Endpoints.Authz))

	for name := range config.Server.Endpoints.Authz {
		authzs = append(authzs, name)
	}

	sort.Strings(authzs)

	for _, name := range authzs {
		endpoint := config.Server.Endpoints.Authz[name]

		if name == legacy {
			switch endpoint.Implementation {
			case authzImplementationLegacy:
				break
			case "":
				endpoint.Implementation = authzImplementationLegacy

				config.Server.Endpoints.Authz[name] = endpoint
			default:
				if !utils.IsStringInSlice(endpoint.Implementation, validAuthzImplementations) {
					validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzImplementation, name, strings.Join(validAuthzImplementations, "', '"), endpoint.Implementation))
				} else {
					validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzLegacyInvalidImplementation, name))
				}
			}
		} else if !utils.IsStringInSlice(endpoint.Implementation, validAuthzImplementations) {
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzImplementation, name, strings.Join(validAuthzImplementations, "', '"), endpoint.Implementation))
		}

		if !reAuthzEndpointName.MatchString(name) {
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzInvalidName, name))
		}

		for _, oName := range authzs {
			oEndpoint := config.Server.Endpoints.Authz[oName]

			if oName == name || oName == legacy {
				continue
			}

			switch oEndpoint.Implementation {
			case authzImplementationLegacy, authzImplementationExtAuthz:
				if strings.HasPrefix(name, oName+"/") {
					validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzPrefixDuplicate, name, oName, oEndpoint.Implementation))
				}
			default:
				continue
			}
		}

		var strategies []string

		for _, strategy := range endpoint.AuthnStrategies {
			if utils.IsStringInSlice(strategy.Name, strategies) {
				validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategyDuplicate, name, strategy.Name))
			}

			strategies = append(strategies, strategy.Name)

			if !utils.IsStringInSlice(strategy.Name, validAuthzAuthnStrategies) {
				validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategy, name, strings.Join(validAuthzAuthnStrategies, "', '"), strategy.Name))
			}
		}
	}
}
