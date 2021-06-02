package validator

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateNotifier validates and update notifier configuration.
func ValidateNotifier(configuration *schema.NotifierConfiguration, validator *schema.StructValidator) {
	providers := utils.CountNil(configuration.SMTP, configuration.FileSystem, configuration.Plugin)

	if providers == 0 {
		validator.Push(errors.New("Please configure one of the `notifier` providers (`smtp`, `filesystem`, or `plugin`)"))
		return
	}

	if providers > 1 {
		validator.Push(errors.New("Please do not configure more than one of the `notifer` providers (`smtp`, `filesystem`, or `plugin`)"))
		return
	}

	switch {
	case configuration.Plugin != nil:
		if runtime.GOOS != linux && runtime.GOOS != freebsd && runtime.GOOS != darwin {
			validator.Push(errors.New("The `notifier` plugin provider is only available on linux, freebsd, and darwin operating systems"))
			return
		}

		if configuration.Plugin.Name == "" {
			validator.Push(errors.New("The `notifier` plugin provider name must be set"))
		}
	case configuration.FileSystem != nil:
		if configuration.FileSystem.Filename == "" {
			validator.Push(fmt.Errorf("Filename of filesystem notifier must not be empty"))
		}
	case configuration.SMTP != nil:
		validateSMTPNotifier(configuration.SMTP, validator)
	}
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
