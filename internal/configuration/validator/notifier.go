package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// ValidateSession validates and update session configuration.
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
		if configuration.SMTP.Host == "" {
			validator.Push(fmt.Errorf("Host of SMTP notifier must be provided"))
		}

		if configuration.SMTP.Port == 0 {
			validator.Push(fmt.Errorf("Port of SMTP notifier must be provided"))
		}

		if configuration.SMTP.Sender == "" {
			validator.Push(fmt.Errorf("Sender of SMTP notifier must be provided"))
		}
		return
	}
}
