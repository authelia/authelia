package validator

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateConfiguration and adapt the configuration read from file.
func ValidateConfiguration(config *schema.Configuration, validator *schema.StructValidator) {
	var err error

	if config.JWTSecret == "" {
		validator.Push(fmt.Errorf("option 'jwt_secret' is required"))
	}

	if config.DefaultRedirectionURL != "" {
		if err = utils.IsStringAbsURL(config.DefaultRedirectionURL); err != nil {
			validator.Push(fmt.Errorf("option 'default_redirection_url' is invalid: %s", strings.ReplaceAll(err.Error(), "like 'http://' or 'https://'", "like 'ldap://' or 'ldaps://'")))
		}
	}

	validateDefault2FAMethod(config, validator)

	ValidateTheme(config, validator)

	ValidateLog(config, validator)

	ValidateDuo(config, validator)

	ValidateTOTP(config, validator)

	ValidateWebauthn(config, validator)

	ValidateAuthenticationBackend(&config.AuthenticationBackend, validator)

	ValidateAccessControl(config, validator)

	ValidateRules(config, validator)

	ValidateSession(&config.Session, validator)

	ValidateRegulation(config, validator)

	ValidateServer(config, validator)

	ValidateTelemetry(config, validator)

	ValidateStorage(config.Storage, validator)

	ValidateNotifier(&config.Notifier, validator)

	ValidateIdentityProviders(&config.IdentityProviders, validator)

	ValidateNTP(config, validator)

	ValidatePasswordPolicy(&config.PasswordPolicy, validator)

	ValidatePrivacyPolicy(&config.PrivacyPolicy, validator)
}

func validateDefault2FAMethod(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Default2FAMethod == "" {
		return
	}

	if !utils.IsStringInSlice(config.Default2FAMethod, validDefault2FAMethods) {
		validator.Push(fmt.Errorf(errFmtInvalidDefault2FAMethod, config.Default2FAMethod, strings.Join(validDefault2FAMethods, "', '")))

		return
	}

	var enabledMethods []string

	if !config.TOTP.Disable {
		enabledMethods = append(enabledMethods, "totp")
	}

	if !config.Webauthn.Disable {
		enabledMethods = append(enabledMethods, "webauthn")
	}

	if !config.DuoAPI.Disable {
		enabledMethods = append(enabledMethods, "mobile_push")
	}

	if !utils.IsStringInSlice(config.Default2FAMethod, enabledMethods) {
		validator.Push(fmt.Errorf(errFmtInvalidDefault2FAMethodDisabled, config.Default2FAMethod, strings.Join(enabledMethods, "', '")))
	}
}
