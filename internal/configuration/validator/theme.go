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
			if (u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/")) || (u.Scheme == schemeHTTPS && u.Host != "") {
				config.CustomCSSURL = u

				if config.Server.Headers.CSPTemplate != "" && u.Host != "" && !isHostAllowedInCSP(string(config.Server.Headers.CSPTemplate), u.Host) {
					validator.PushWarning(fmt.Errorf(errFmtCustomCSSCSPTemplateIncompatibility, config.CustomCSS))
				}
			} else {
				validator.Push(fmt.Errorf(errFmtCustomCSSURL, config.CustomCSS, errors.New("must be an absolute path or an https URL")))
			}
		}
	}
}

func isHostAllowedInCSP(csp, host string) bool {
	directives := strings.Split(csp, ";")
	var styleSrc, defaultSrc string

	for _, d := range directives {
		d = strings.TrimSpace(d)
		if strings.HasPrefix(d, "style-src ") {
			styleSrc = d
		} else if strings.HasPrefix(d, "default-src ") {
			defaultSrc = d
		}
	}

	if styleSrc != "" {
		return checkDirective(styleSrc, host)
	}

	if defaultSrc != "" {
		return checkDirective(defaultSrc, host)
	}

	return false
}

func checkDirective(directive, host string) bool {
	parts := strings.Fields(directive)
	if len(parts) < 2 {
		return false
	}

	for _, p := range parts[1:] {
		if p == host || p == "*" || p == "https:*" || p == "https://*" {
			return true
		}

		if strings.HasPrefix(p, "https://") {
			if strings.TrimPrefix(p, "https://") == host {
				return true
			}
		}
	}

	return false
}
