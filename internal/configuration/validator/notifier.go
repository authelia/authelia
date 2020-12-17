package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
)

// ValidateNotifier validates and update notifier configuration.
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

		log := logging.Logger() // Deprecated: final removal in 4.28.

		if configuration.SMTP.TLS == nil {
			configuration.SMTP.TLS = schema.DefaultSMTPNotifierConfiguration.TLS

			// Deprecated: final removal in 4.28.
			if configuration.SMTP.DisableVerifyCert != nil {
				log.Warnf("DEPRECATED: SMTP Notifier `disable_verify_cert` option has been replaced by `notifier.smtp.tls.skip_verify` (will be removed in 4.28.0)")

				configuration.SMTP.TLS.SkipVerify = *configuration.SMTP.DisableVerifyCert
			}
		}

		// Deprecated: final removal in 4.28.
		if configuration.SMTP.TrustedCert != "" {
			log.Warnf("DEPRECATED: SMTP Notifier `trusted_cert` option has been replaced by the global option `certificates_directory` (will be removed in 4.28.0)")
		}

		if configuration.SMTP.TLS.ServerName == "" {
			configuration.SMTP.TLS.ServerName = configuration.SMTP.Host
		}

		return
	}
}
