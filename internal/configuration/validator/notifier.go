package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// ValidateNotifier validates and update notifier configuration.
func ValidateNotifier(configuration *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if configuration.SMTP == nil && configuration.FileSystem == nil {
		validator.Push(fmt.Errorf(errFmtNotifierNotConfigured))

		return
	} else if configuration.SMTP != nil && configuration.FileSystem != nil {

		if configuration.SMTP.Password != "" &&
			configuration.SMTP.Host == "" && configuration.SMTP.Port == 0 && configuration.SMTP.Sender == "" &&
			configuration.SMTP.Subject == "" && configuration.SMTP.Identifier == "" && configuration.SMTP.Username == "" {
			validator.Push(fmt.Errorf(errFmtNotifierMultipleConfiguredSecret))
		} else {
			validator.Push(fmt.Errorf(errFmtNotifierMultipleConfigured))
		}

		return
	}

	if configuration.FileSystem != nil {
		if configuration.FileSystem.Filename == "" {
			validator.Push(fmt.Errorf(errFmtNotifierFileSystemFileNameNotConfigured))
		}

		return
	}

	validateSMTPNotifier(configuration.SMTP, validator)
}

func validateSMTPNotifier(configuration *schema.SMTPNotifierConfiguration, validator *schema.StructValidator) {
	if configuration.StartupCheckAddress == "" {
		configuration.StartupCheckAddress = "test@authelia.com"
	}

	if configuration.Host == "" {
		validator.Push(fmt.Errorf(errFmtNotifierSMTPNotConfigured, "host"))
	}

	if configuration.Port == 0 {
		validator.Push(fmt.Errorf(errFmtNotifierSMTPNotConfigured, "port"))
	}

	if configuration.Sender == "" {
		validator.Push(fmt.Errorf(errFmtNotifierSMTPNotConfigured, "sender"))
	}

	if configuration.Subject == "" {
		configuration.Subject = schema.DefaultSMTPNotifierConfiguration.Subject
	}

	if configuration.Identifier == "" {
		configuration.Identifier = schema.DefaultSMTPNotifierConfiguration.Identifier
	}

	if configuration.TLS == nil {
		configuration.TLS = schema.DefaultSMTPNotifierConfiguration.TLS
	}

	if configuration.TLS.ServerName == "" {
		configuration.TLS.ServerName = configuration.Host
	}
}
