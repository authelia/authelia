package validator

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateServerTLS checks a server TLS configuration is correct.
func ValidateServerTLS(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.TLS.Key != "" && config.Server.TLS.Certificate == "" {
		validator.Push(fmt.Errorf(errFmtServerTLSCert))
	} else if config.Server.TLS.Key == "" && config.Server.TLS.Certificate != "" {
		validator.Push(fmt.Errorf(errFmtServerTLSKey))
	}

	if config.Server.TLS.Key != "" {
		validateServerTLSFileExists("key", config.Server.TLS.Key, validator)
	}

	if config.Server.TLS.Certificate != "" {
		validateServerTLSFileExists("certificate", config.Server.TLS.Certificate, validator)
	}

	if config.Server.TLS.Key == "" && config.Server.TLS.Certificate == "" &&
		len(config.Server.TLS.ClientCertificates) > 0 {
		validator.Push(fmt.Errorf(errFmtServerTLSClientAuthNoAuth))
	}

	for _, clientCertPath := range config.Server.TLS.ClientCertificates {
		validateServerTLSFileExists("client_certificates", clientCertPath, validator)
	}
}

// validateServerTLSFileExists checks whether a file exist.
func validateServerTLSFileExists(name, path string, validator *schema.StructValidator) {
	var (
		info os.FileInfo
		err  error
	)

	switch info, err = os.Stat(path); {
	case os.IsNotExist(err):
		validator.Push(fmt.Errorf("server: tls: option '%s' with path '%s' refers to a file that doesn't exist", name, path))
	case err != nil:
		validator.Push(fmt.Errorf("server: tls: option '%s' with path '%s' could not be verified due to a file system error: %w", name, path, err))
	case info.IsDir():
		validator.Push(fmt.Errorf("server: tls: option '%s' with path '%s' refers to a directory but it should refer to a file", name, path))
	}
}

// ValidateServer checks the server configuration is correct.
func ValidateServer(config *schema.Configuration, validator *schema.StructValidator) {
	ValidateServerAddress(config, validator)
	ValidateServerTLS(config, validator)

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

// ValidateServerAddress checks the configured server address is correct.
//

func ValidateServerAddress(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.Address == nil {
		config.Server.Address = schema.DefaultServerConfiguration.Address
	} else {
		var err error

		if err = config.Server.Address.ValidateHTTP(); err != nil {
			validator.Push(fmt.Errorf(errFmtServerAddress, config.Server.Address.String(), err))
		}
	}

	switch subpath := config.Server.Address.RouterPath(); {
	case subpath == "":
		config.Server.Address.SetPath("/")
	case subpath != "/":
		if p := strings.TrimPrefix(subpath, "/"); strings.Contains(p, "/") {
			validator.Push(fmt.Errorf(errFmtServerPathNotEndForwardSlash, subpath))
		} else if !utils.IsStringAlphaNumeric(p) {
			validator.Push(fmt.Errorf(errFmtServerPathAlphaNumeric, subpath))
		}
	}
}

// ValidateServerEndpoints configures the default endpoints and checks the configuration of custom endpoints.
func ValidateServerEndpoints(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.Endpoints.EnableExpvars {
		validator.PushWarning(fmt.Errorf("server: endpoints: option 'enable_expvars' should not be enabled in production"))
	}

	if config.Server.Endpoints.EnablePprof {
		validator.PushWarning(fmt.Errorf("server: endpoints: option 'enable_pprof' should not be enabled in production"))
	}

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

		validateServerEndpointsAuthzEndpoint(config, name, endpoint, validator)

		for _, oName := range authzs {
			oEndpoint := config.Server.Endpoints.Authz[oName]

			if oName == name || oName == legacy {
				continue
			}

			switch oEndpoint.Implementation {
			case schema.AuthzImplementationLegacy, schema.AuthzImplementationExtAuthz:
				if strings.HasPrefix(name, oName+"/") {
					validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzPrefixDuplicate, name, oName, oEndpoint.Implementation))
				}
			default:
				continue
			}
		}

		validateServerEndpointsAuthzStrategies(name, endpoint.Implementation, endpoint.AuthnStrategies, validator)
	}
}

func validateServerEndpointsAuthzEndpoint(config *schema.Configuration, name string, endpoint schema.ServerEndpointsAuthz, validator *schema.StructValidator) {
	if name == legacy {
		switch endpoint.Implementation {
		case schema.AuthzImplementationLegacy:
			break
		case "":
			endpoint.Implementation = schema.AuthzImplementationLegacy

			config.Server.Endpoints.Authz[name] = endpoint
		default:
			if !utils.IsStringInSlice(endpoint.Implementation, validAuthzImplementations) {
				validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzImplementation, name, strJoinOr(validAuthzImplementations), endpoint.Implementation))
			} else {
				validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzLegacyInvalidImplementation, name))
			}
		}
	} else if !utils.IsStringInSlice(endpoint.Implementation, validAuthzImplementations) {
		validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzImplementation, name, strJoinOr(validAuthzImplementations), endpoint.Implementation))
	}

	if !reAuthzEndpointName.MatchString(name) {
		validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzInvalidName, name))
	}
}

//nolint:gocyclo
func validateServerEndpointsAuthzStrategies(name, implementation string, strategies []schema.ServerEndpointsAuthzAuthnStrategy, validator *schema.StructValidator) {
	var defaults []schema.ServerEndpointsAuthzAuthnStrategy

	switch implementation {
	case schema.AuthzImplementationLegacy:
		defaults = schema.DefaultServerConfiguration.Endpoints.Authz[schema.AuthzEndpointNameLegacy].AuthnStrategies
	case schema.AuthzImplementationAuthRequest:
		defaults = schema.DefaultServerConfiguration.Endpoints.Authz[schema.AuthzEndpointNameAuthRequest].AuthnStrategies
	case schema.AuthzImplementationExtAuthz:
		defaults = schema.DefaultServerConfiguration.Endpoints.Authz[schema.AuthzEndpointNameExtAuthz].AuthnStrategies
	case schema.AuthzImplementationForwardAuth:
		defaults = schema.DefaultServerConfiguration.Endpoints.Authz[schema.AuthzEndpointNameForwardAuth].AuthnStrategies
	}

	if len(strategies) == 0 {
		copy(strategies, defaults)

		return
	}

	names := make([]string, len(strategies))

	for i, strategy := range strategies {
		if strategy.Name != "" && utils.IsStringInSlice(strategy.Name, names) {
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategyDuplicate, name, strategy.Name))
		}

		names = append(names, strategy.Name)

		switch {
		case strategy.Name == "":
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategyNoName, name, i+1))
		case !utils.IsStringInSlice(strategy.Name, validAuthzAuthnStrategies):
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategy, name, strJoinOr(validAuthzAuthnStrategies), strategy.Name))
		default:
			if utils.IsStringInSlice(strategy.Name, validAuthzAuthnHeaderStrategies) {
				if len(strategy.Schemes) == 0 {
					strategies[i].Schemes = defaults[0].Schemes
				} else {
					for _, scheme := range strategy.Schemes {
						if !utils.IsStringInSliceFold(scheme, validAuthzAuthnStrategySchemes) {
							validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzSchemes, name, i+1, strategy.Name, strJoinOr(validAuthzAuthnStrategySchemes), scheme))
						}
					}
				}
			} else if len(strategy.Schemes) != 0 {
				validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzSchemesInvalidForStrategy, name, i+1, strategy.Name))
			}
		}
	}
}
