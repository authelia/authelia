package validator

import (
	"fmt"
	"path"
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
		config.Server.Host = schema.DefaultServerConfig.Host
	}

	if config.Server.Port == 0 {
		config.Server.Port = schema.DefaultServerConfig.Port
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

	if config.Server.ReadBufferSize == 0 {
		config.Server.ReadBufferSize = schema.DefaultServerConfig.ReadBufferSize
	} else if config.Server.ReadBufferSize < 0 {
		validator.Push(fmt.Errorf(errFmtServerBufferSize, "read", config.Server.ReadBufferSize))
	}

	if config.Server.WriteBufferSize == 0 {
		config.Server.WriteBufferSize = schema.DefaultServerConfig.WriteBufferSize
	} else if config.Server.WriteBufferSize < 0 {
		validator.Push(fmt.Errorf(errFmtServerBufferSize, "write", config.Server.WriteBufferSize))
	}

	ValidateServerEndpoints(config, validator)
}

func ValidateServerEndpoints(config *schema.Configuration, validator *schema.StructValidator) {
	if len(config.Server.Endpoints.Authz) == 0 {
		config.Server.Endpoints.Authz = schema.DefaultServerConfig.Endpoints.Authz
	}

	for name, entrypoint := range config.Server.Endpoints.Authz {
		if !utils.IsStringInSlice(entrypoint.Implementation, validAuthzImplementations) {
			validator.Push(fmt.Errorf(errFmtServerEndpointsAuthzImplementation, name, strings.Join(validAuthzImplementations, "', '"), entrypoint.Implementation))
		}

		var strategies []string

		for _, strategy := range entrypoint.AuthnStrategies {
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
