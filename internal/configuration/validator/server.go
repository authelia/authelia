package validator

import (
	"fmt"
	"path"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateServer checks a server configuration is correct.
func ValidateServer(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Server.Host == "" {
		config.Server.Host = schema.DefaultServerConfiguration.Host
	}

	if config.Server.Port == 0 {
		config.Server.Port = schema.DefaultServerConfiguration.Port
	}

	if config.Server.TLS.Key != "" && config.Server.TLS.Certificate == "" {
		validator.Push(fmt.Errorf(errFmtServerTLSCert))
	} else if config.Server.TLS.Key == "" && config.Server.TLS.Certificate != "" {
		validator.Push(fmt.Errorf(errFmtServerTLSKey))
	}

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
		config.Server.ReadBufferSize = schema.DefaultServerConfiguration.ReadBufferSize
	} else if config.Server.ReadBufferSize < 0 {
		validator.Push(fmt.Errorf(errFmtServerBufferSize, "read", config.Server.ReadBufferSize))
	}

	if config.Server.WriteBufferSize == 0 {
		config.Server.WriteBufferSize = schema.DefaultServerConfiguration.WriteBufferSize
	} else if config.Server.WriteBufferSize < 0 {
		validator.Push(fmt.Errorf(errFmtServerBufferSize, "write", config.Server.WriteBufferSize))
	}
}
