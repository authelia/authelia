package validator

import (
	"fmt"
	"os"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateConfiguration and adapt the configuration read from file.
func ValidateConfiguration(config *schema.Configuration, validator *schema.StructValidator) {
	var err error

	if config.CertificatesDirectory != "" {
		var info os.FileInfo

		if info, err = os.Stat(config.CertificatesDirectory); err != nil {
			validator.Push(fmt.Errorf("the location 'certificates_directory' could not be inspected: %w", err))
		} else if !info.IsDir() {
			validator.Push(fmt.Errorf("the location 'certificates_directory' refers to '%s' is not a directory", config.CertificatesDirectory))
		}
	}

	if config.JWTSecret == "" {
		validator.Push(fmt.Errorf("option 'jwt_secret' is required"))
	}

	if config.DefaultRedirectionURL != "" {
		if err = utils.IsStringAbsURL(config.DefaultRedirectionURL); err != nil {
			validator.Push(fmt.Errorf("option 'default_redirection_url' is invalid: %s", strings.ReplaceAll(err.Error(), "like 'http://' or 'https://'", "like 'ldap://' or 'ldaps://'")))
		}
	}

	ValidateTheme(config, validator)

	ValidateLog(config, validator)

	ValidateTOTP(config, validator)

	ValidateWebauthn(config, validator)

	ValidateAuthenticationBackend(&config.AuthenticationBackend, validator)

	ValidateAccessControl(config, validator)

	ValidateRules(config, validator)

	ValidateSession(&config.Session, validator)

	ValidateRegulation(config, validator)

	ValidateServer(config, validator)

	ValidateStorage(config.Storage, validator)

	ValidateNotifier(config.Notifier, validator)

	ValidateIdentityProviders(&config.IdentityProviders, validator)

	ValidateNTP(config, validator)
}
