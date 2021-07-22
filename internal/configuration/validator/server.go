package validator

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

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
		configuration.ReadBufferSize = schema.DefaultServerConfiguration.ReadBufferSize
	} else if configuration.ReadBufferSize < 0 {
		validator.Push(fmt.Errorf("server read buffer size must be above 0"))
	}

	if configuration.WriteBufferSize == 0 {
		configuration.WriteBufferSize = schema.DefaultServerConfiguration.WriteBufferSize
	} else if configuration.WriteBufferSize < 0 {
		validator.Push(fmt.Errorf("server write buffer size must be above 0"))
	}

	validateCORS(&configuration.CORS, validator)
}

func validateCORS(configuration *schema.CORSConfiguration, validator *schema.StructValidator) {
	if configuration.MaxAge == 0 {
		configuration.MaxAge = schema.DefaultServerConfiguration.CORS.MaxAge
	}

	if len(configuration.Origins) != 1 || configuration.Origins[0] != "*" {
		for _, origin := range configuration.Origins {
			originURL, err := url.Parse(origin)
			if err != nil {
				validator.Push(fmt.Errorf(errFmtServerCORSOriginParse, origin, err))

				continue
			}

			if originURL.Scheme != "https" {
				validator.Push(fmt.Errorf(errFmtServerCORSOriginScheme, origin))
			}

			if originURL.Path != "" || originURL.RawPath != "" {
				validator.Push(fmt.Errorf(errFmtServerCORSOriginPath, origin))
			}
		}
	}

	if len(configuration.Methods) != 1 || configuration.Methods[0] != "*" {
		for _, method := range configuration.Methods {
			if !utils.IsStringInSlice(method, validHTTPRequestMethods) {
				validator.Push(fmt.Errorf(errFmtServerCORSMethods, method, strings.Join(validHTTPRequestMethods, ", ")))
			}
		}
	}

	if len(configuration.Vary) == 0 {
		configuration.Vary = schema.DefaultServerConfiguration.CORS.Vary
	}
}
