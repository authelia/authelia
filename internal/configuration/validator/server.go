package validator

import (
	"fmt"
	"path"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

var defaultReadBufferSize = 4096
var defaultWriteBufferSize = 4096

// ValidateServer checks a server configuration is correct.
func ValidateServer(configuration *schema.ServerConfiguration, validator *schema.StructValidator) {
	switch {
	case strings.Contains(configuration.Path, "/"):
		validator.Push(fmt.Errorf("server path must not contain any forward slashes"))
	case !utils.IsStringAlphaNumeric(configuration.Path):
		validator.Push(fmt.Errorf("server path must only be alpha numeric characters"))
	case configuration.Path == "": // Don't do anything if it's blank.
	default:
		configuration.Path = path.Clean("/" + configuration.Path)
	}

	if configuration.ReadBufferSize == 0 {
		configuration.ReadBufferSize = defaultReadBufferSize
	} else if configuration.ReadBufferSize < 0 {
		validator.Push(fmt.Errorf("server read buffer size must be above 0"))
	}

	if configuration.WriteBufferSize == 0 {
		configuration.WriteBufferSize = defaultWriteBufferSize
	} else if configuration.WriteBufferSize < 0 {
		validator.Push(fmt.Errorf("server write buffer size must be above 0"))
	}
}
