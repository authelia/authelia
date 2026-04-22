package validator

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateTheme validates and update Theme configuration.
func ValidateTheme(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Theme == "" {
		config.Theme = "light"
	}

	if !utils.IsStringInSlice(config.Theme, validThemeNames) {
		validator.Push(fmt.Errorf(errFmtThemeName, utils.StringJoinOr(validThemeNames), config.Theme))
	}

	if config.CustomCSS != "" {
		if u, err := url.Parse(config.CustomCSS); err != nil {
			validator.Push(fmt.Errorf(errFmtCustomCSSURL, config.CustomCSS, err))
		} else {
			if u.Scheme != schemeHTTPS && (u.Scheme != "" || u.Host != "" || !strings.HasPrefix(u.Path, "/")) {
				validator.Push(fmt.Errorf(errFmtCustomCSSURL, config.CustomCSS, errors.New("must be an absolute path or an https URL")))
			}
		}
	}
}
