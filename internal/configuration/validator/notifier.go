package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/notification"
)

// ValidateNotifier validates and update notifier configuration.
func ValidateNotifier(configuration *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if configuration.SMTP == nil && configuration.FileSystem == nil {
		validator.Push(fmt.Errorf(errFmtNotifierNotConfigured))

		return
	} else if configuration.SMTP != nil && configuration.FileSystem != nil {
		validator.Push(fmt.Errorf(errFmtNotifierMultipleConfigured))

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

	if configuration.Timeout == 0 {
		configuration.Timeout = schema.DefaultSMTPNotifierConfiguration.Timeout
	}

	if configuration.Sender == "" {
		if notification.RegexpValidEmail.MatchString(configuration.Username) {
			configuration.Sender = configuration.Username
		} else {
			validator.Push(fmt.Errorf(errFmtNotifierSMTPNotConfigured, "sender"))
		}
	} else if !notification.RegexpValidEmail.MatchString(configuration.Sender) {
		validator.Push(fmt.Errorf(errFmtNotifierSMTPSenderMustBeValidEmail, configuration.Sender))
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
