package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// ValidateNotifier validates and update notifier configuration.
func ValidateNotifier(configuration *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if configuration.SMTP == nil && configuration.FileSystem == nil ||
		configuration.SMTP != nil && configuration.FileSystem != nil {
		validator.Push(fmt.Errorf("Notifier should be either `smtp` or `filesystem`"))

		return
	}

	if configuration.FileSystem != nil {
		if configuration.FileSystem.Filename == "" {
			validator.Push(fmt.Errorf("Filename of filesystem notifier must not be empty"))
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
		validator.Push(fmt.Errorf("Host of SMTP notifier must be provided"))
	}

	if configuration.Port == 0 {
		validator.Push(fmt.Errorf("Port of SMTP notifier must be provided"))
	}

	if configuration.Sender == "" {
		validator.Push(fmt.Errorf("Sender of SMTP notifier must be provided"))
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
