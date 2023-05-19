package validator

import (
	"fmt"
	"os"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateNotifier validates and update notifier configuration.
func ValidateNotifier(config *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if config.SMTP == nil && config.FileSystem == nil {
		validator.Push(fmt.Errorf(errFmtNotifierNotConfigured))

		return
	} else if config.SMTP != nil && config.FileSystem != nil {
		validator.Push(fmt.Errorf(errFmtNotifierMultipleConfigured))

		return
	}

	if config.FileSystem != nil {
		if config.FileSystem.Filename == "" {
			validator.Push(fmt.Errorf(errFmtNotifierFileSystemFileNameNotConfigured))
		}

		return
	}

	validateSMTPNotifier(config.SMTP, validator)

	validateNotifierTemplates(config, validator)
}

func validateNotifierTemplates(config *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if config.TemplatePath == "" {
		return
	}

	switch _, err := os.Stat(config.TemplatePath); {
	case os.IsNotExist(err):
		validator.Push(fmt.Errorf(errFmtNotifierTemplatePathNotExist, config.TemplatePath))
		return
	case err != nil:
		validator.Push(fmt.Errorf(errFmtNotifierTemplatePathUnknownError, config.TemplatePath, err))
		return
	}
}

func validateSMTPNotifier(config *schema.SMTPNotifierConfiguration, validator *schema.StructValidator) {
	validateSMTPNotifierAddress(config, validator)

	if config.StartupCheckAddress.Address == "" {
		config.StartupCheckAddress = schema.DefaultSMTPNotifierConfiguration.StartupCheckAddress
	}

	if config.Timeout == 0 {
		config.Timeout = schema.DefaultSMTPNotifierConfiguration.Timeout
	}

	if config.Sender.Address == "" {
		validator.Push(fmt.Errorf(errFmtNotifierSMTPNotConfigured, "sender"))
	}

	if config.Subject == "" {
		config.Subject = schema.DefaultSMTPNotifierConfiguration.Subject
	}

	if config.Identifier == "" {
		config.Identifier = schema.DefaultSMTPNotifierConfiguration.Identifier
	}

	if config.TLS == nil {
		config.TLS = &schema.TLSConfig{}
	}

	configDefaultTLS := &schema.TLSConfig{
		MinimumVersion: schema.DefaultSMTPNotifierConfiguration.TLS.MinimumVersion,
		MaximumVersion: schema.DefaultSMTPNotifierConfiguration.TLS.MaximumVersion,
	}

	if config.Address != nil {
		configDefaultTLS.ServerName = config.Address.Hostname()
	}

	if err := ValidateTLSConfig(config.TLS, configDefaultTLS); err != nil {
		validator.Push(fmt.Errorf(errFmtNotifierSMTPTLSConfigInvalid, err))
	}

	if config.DisableStartTLS {
		validator.PushWarning(fmt.Errorf(errFmtNotifierStartTlsDisabled))
	}
}

func validateSMTPNotifierAddress(config *schema.SMTPNotifierConfiguration, validator *schema.StructValidator) {
	if config.Address == nil {
		if config.Host == "" && config.Port == 0 { //nolint:staticcheck
			validator.Push(fmt.Errorf(errFmtNotifierSMTPNotConfigured, "address"))
		} else {
			host := config.Host //nolint:staticcheck
			port := config.Port //nolint:staticcheck

			config.Address = schema.NewSMTPAddress("", host, port)
		}
	} else {
		if config.Host != "" || config.Port != 0 { //nolint:staticcheck
			validator.Push(fmt.Errorf(errFmtNotifierSMTPAddressLegacyAndModern))
		}

		var err error

		if err = config.Address.ValidateSMTP(); err != nil {
			validator.Push(fmt.Errorf(errFmtNotifierSMTPAddress, config.Address.String(), err))
		}
	}
}
