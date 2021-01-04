package validator

import (
	"errors"
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// ValidateNotifier validates and update notifier configuration.
//nolint:gocyclo // TODO: Remove in 4.28. Should be able to remove this during the removal of deprecated config.
func ValidateNotifier(configuration *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if configuration.SMTP == nil && configuration.FileSystem == nil {
		validator.Push(fmt.Errorf("Notifier should be either `smtp` or `filesystem`"))
		return
	}

	if configuration.SMTP != nil && configuration.FileSystem != nil {
		validator.Push(fmt.Errorf("Notifier should be either `smtp` or `filesystem`"))
		return
	}

	if configuration.FileSystem != nil {
		if configuration.FileSystem.Filename == "" {
			validator.Push(fmt.Errorf("Filename of filesystem notifier must not be empty"))
		}

		return
	}

	if configuration.SMTP != nil {
		if configuration.SMTP.StartupCheckAddress == "" {
			configuration.SMTP.StartupCheckAddress = "test@authelia.com"
		}

		if configuration.SMTP.Host == "" {
			validator.Push(fmt.Errorf("Host of SMTP notifier must be provided"))
		}

		if configuration.SMTP.Port == 0 {
			validator.Push(fmt.Errorf("Port of SMTP notifier must be provided"))
		}

		if configuration.SMTP.Sender == "" {
			validator.Push(fmt.Errorf("Sender of SMTP notifier must be provided"))
		}

		if configuration.SMTP.Subject == "" {
			configuration.SMTP.Subject = schema.DefaultSMTPNotifierConfiguration.Subject
		}

		if configuration.SMTP.Identifier == "" {
			configuration.SMTP.Identifier = schema.DefaultSMTPNotifierConfiguration.Identifier
		}

		if configuration.SMTP.TLS == nil {
			configuration.SMTP.TLS = schema.DefaultSMTPNotifierConfiguration.TLS

			// Deprecated. Maps deprecated values to the new ones. TODO: Remove in 4.28.
			if configuration.SMTP.DisableVerifyCert != nil {
				validator.PushWarning(errors.New("DEPRECATED: SMTP Notifier `disable_verify_cert` option has been replaced by `notifier.smtp.tls.skip_verify` (will be removed in 4.28.0)"))

				configuration.SMTP.TLS.SkipVerify = *configuration.SMTP.DisableVerifyCert
			}
		}

		// Deprecated. Maps deprecated values to the new ones. TODO: Remove in 4.28.
		if configuration.SMTP.TrustedCert != "" {
			validator.PushWarning(errors.New("DEPRECATED: SMTP Notifier `trusted_cert` option has been replaced by the global option `certificates_directory` (will be removed in 4.28.0)"))
		}

		if configuration.SMTP.TLS.ServerName == "" {
			configuration.SMTP.TLS.ServerName = configuration.SMTP.Host
		}

		return
	}
}
