package validator

import (
	"fmt"
	"path"
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
		validateFileExists(config.Server.TLS.Key, validator, "key")
	}

	if config.Server.TLS.Certificate != "" {
		validateFileExists(config.Server.TLS.Certificate, validator, "certificate")
	}

	if config.Server.TLS.Key == "" && config.Server.TLS.Certificate == "" &&
		len(config.Server.TLS.ClientCertificates) > 0 {
		validator.Push(fmt.Errorf(errFmtServerTLSClientAuthNoAuth))
	}

	for _, clientCertPath := range config.Server.TLS.ClientCertificates {
		validateFileExists(clientCertPath, validator, "client_certificates")
	}
}

// ValidateServer checks the server configuration is correct.
func ValidateServer(config *schema.Configuration, validator *schema.StructValidator) {
	ValidateServerAddress(config, validator)
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

// ValidateServerAddress checks the configured server address is correct.
func ValidateServerAddress(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.Address == nil {
		if config.Server.Host == "" && config.Server.Port == 0 { //nolint:staticcheck
			config.Server.Address = schema.DefaultServerConfiguration.Address
		} else {
			host := config.Server.Host //nolint:staticcheck
			port := config.Server.Port //nolint:staticcheck

			if host == "" {
				host = schema.DefaultServerConfiguration.Address.Hostname()
			}

			if port == 0 {
				port = schema.DefaultServerConfiguration.Address.Port()
			}

			config.Server.Address = &schema.AddressTCP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeTCP, host, port)}
		}
	} else {
		if config.Server.Host != "" || config.Server.Port != 0 { //nolint:staticcheck
			validator.Push(fmt.Errorf(errFmtServerAddressLegacyAndModern))
		}

		var err error

		if err = config.Server.Address.ValidateHTTP(); err != nil {
			validator.Push(fmt.Errorf(errFmtServerAddress, config.Server.Address.String(), err))
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
			case authzImplementationLegacy, authzImplementationExtAuthz:
				if strings.HasPrefix(name, oName+"/") {
					validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzPrefixDuplicate, name, oName, oEndpoint.Implementation))
				}
			default:
				continue
			}
		}

		validateServerEndpointsAuthzStrategies(name, endpoint.AuthnStrategies, validator)
	}
}

func validateServerEndpointsAuthzEndpoint(config *schema.Configuration, name string, endpoint schema.ServerAuthzEndpoint, validator *schema.StructValidator) {
	if name == legacy {
		switch endpoint.Implementation {
		case authzImplementationLegacy:
			break
		case "":
			endpoint.Implementation = authzImplementationLegacy

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

func validateServerEndpointsAuthzStrategies(name string, strategies []schema.ServerAuthzEndpointAuthnStrategy, validator *schema.StructValidator) {
	names := make([]string, len(strategies))

	for _, strategy := range strategies {
		if utils.IsStringInSlice(strategy.Name, names) {
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategyDuplicate, name, strategy.Name))
		}

		names = append(names, strategy.Name)

		if !utils.IsStringInSlice(strategy.Name, validAuthzAuthnStrategies) {
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzStrategy, name, strJoinOr(validAuthzAuthnStrategies), strategy.Name))
		}
	}
}

// validateFileExists checks whether a file exist.
func validateFileExists(path string, validator *schema.StructValidator, opt string) {
	exist, err := utils.FileExists(path)
	if err != nil {
		validator.Push(fmt.Errorf(errFmtServerTLSFileNotExistErr, opt, path, err))
	} else if !exist {
		validator.Push(fmt.Errorf(errFmtServerTLSFileNotExist, opt, path))
	}
}
