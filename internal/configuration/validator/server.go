package validator

import (
	"fmt"
	"path"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateServer checks a server configuration is correct.
func ValidateServer(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.Server.Host == "" {
		configuration.Server.Host = schema.DefaultServerConfiguration.Host
	}

	if configuration.Server.Port == 0 {
		configuration.Server.Port = schema.DefaultServerConfiguration.Port
	}

	if configuration.Server.TLS.Key != "" && configuration.Server.TLS.Certificate == "" {
		validator.Push(fmt.Errorf("server: no TLS certificate provided to accompany the TLS key, please configure the 'server.tls.certificate' option"))
	} else if configuration.Server.TLS.Key == "" && configuration.Server.TLS.Certificate != "" {
		validator.Push(fmt.Errorf("server: no TLS key provided to accompany the TLS certificate, please configure the 'server.tls.key' option"))
	}

	switch {
	case strings.Contains(configuration.Server.Path, "/"):
		validator.Push(fmt.Errorf("server path must not contain any forward slashes"))
	case !utils.IsStringAlphaNumeric(configuration.Server.Path):
		validator.Push(fmt.Errorf("server path must only be alpha numeric characters"))
	case configuration.Server.Path == "": // Don't do anything if it's blank.
	default:
		configuration.Server.Path = path.Clean("/" + configuration.Server.Path)
	}

	if configuration.Server.ReadBufferSize == 0 {
		configuration.Server.ReadBufferSize = schema.DefaultServerConfiguration.ReadBufferSize
	} else if configuration.Server.ReadBufferSize < 0 {
		validator.Push(fmt.Errorf("server read buffer size must be above 0"))
	}

	if configuration.Server.WriteBufferSize == 0 {
		configuration.Server.WriteBufferSize = schema.DefaultServerConfiguration.WriteBufferSize
	} else if configuration.Server.WriteBufferSize < 0 {
		validator.Push(fmt.Errorf("server write buffer size must be above 0"))
	}
}
