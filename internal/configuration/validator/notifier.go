package validator

import (
	"fmt"
	"os"
	"text/template"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	autheliaTemplates "github.com/authelia/authelia/v4/internal/templates"
)

// ValidateNotifier validates and update notifier configuration.
func ValidateNotifier(config *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if config == nil || (config.SMTP == nil && config.FileSystem == nil) {
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

func validateNotifierTemplates(configuration *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if configuration.TemplatePath != "" {
		_, err := os.Stat(configuration.TemplatePath)
		if os.IsNotExist(err) {
			validator.PushWarning(fmt.Errorf("e-mail template folder '%s' does not exists. Using default templates", configuration.TemplatePath))
			return
		}

		if t, err := template.ParseFiles(configuration.TemplatePath + `/PasswordResetStep1.html`); err == nil {
			autheliaTemplates.HTMLEmailTemplateStep1 = t
		} else {
			validator.PushWarning(fmt.Errorf("error loading html template: %s ", err.Error()))
		}

		if t, err := template.ParseFiles(configuration.TemplatePath + `/PasswordResetStep1.txt`); err == nil {
			autheliaTemplates.PlainTextEmailTemplateStep1 = t
		} else {
			validator.PushWarning(fmt.Errorf("error loading text template: %s ", err.Error()))
		}

		if t, err := template.ParseFiles(configuration.TemplatePath + `/PasswordResetStep2.html`); err == nil {
			autheliaTemplates.HTMLEmailTemplateStep2 = t
		} else {
			validator.PushWarning(fmt.Errorf("error loading html template: %s ", err.Error()))
		}

		if t, err := template.ParseFiles(configuration.TemplatePath + `/PasswordResetStep2.txt`); err == nil {
			autheliaTemplates.PlainTextEmailTemplateStep2 = t
		} else {
			validator.PushWarning(fmt.Errorf("error loading text template: %s ", err.Error()))
		}
	}
}

func validateSMTPNotifier(config *schema.SMTPNotifierConfiguration, validator *schema.StructValidator) {
	if config.StartupCheckAddress == "" {
		config.StartupCheckAddress = schema.DefaultSMTPNotifierConfiguration.StartupCheckAddress
	}

	if config.Host == "" {
		validator.Push(fmt.Errorf(errFmtNotifierSMTPNotConfigured, "host"))
	}

	if config.Port == 0 {
		validator.Push(fmt.Errorf(errFmtNotifierSMTPNotConfigured, "port"))
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
		config.TLS = schema.DefaultSMTPNotifierConfiguration.TLS
	}

	if config.TLS.ServerName == "" {
		config.TLS.ServerName = config.Host
	}
}
