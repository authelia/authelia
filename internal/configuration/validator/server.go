package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
)

var defaultReadBufferSize = 4096
var defaultWriteBufferSize = 4096

// ValidateServer checks a server configuration is correct.
func ValidateServer(configuration *schema.ServerConfiguration, validator *schema.StructValidator) {
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
