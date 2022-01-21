package validator

import (
	"fmt"
	"os"
	"text/template"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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

	validateNotifierTemplates(configuration, validator)
}

func validateNotifierTemplates(configuration *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if configuration.TemplatePath != "" {
		_, err := os.Stat(configuration.TemplatePath)
		if os.IsNotExist(err) {
			validator.PushWarning(fmt.Errorf("e-mail template folder '%s' does not exists. Using default templates", configuration.TemplatePath))
			return
		}

		configuration.Template = &schema.NotifierTemplateConfiguration{}
		configuration.Template.PasswordResetHTML, err = template.ParseFiles(configuration.TemplatePath + `/PasswordResetTemplate.html`)

		if err != nil {
			validator.PushWarning(fmt.Errorf("error loading html template: %s ", err.Error()))
		}

		configuration.Template.PasswordResetText, err = template.ParseFiles(configuration.TemplatePath + `/PasswordResetTemplate.txt`)
		if err != nil {
			validator.PushWarning(fmt.Errorf("error loading text template: %s ", err.Error()))
		}

		configuration.Template.PasswordChangedHTML, err = template.ParseFiles(configuration.TemplatePath + `/PasswordChangedTemplate.html`)
		if err != nil {
			validator.PushWarning(fmt.Errorf("error loading html template: %s ", err.Error()))
		}

		configuration.Template.PasswordChangedText, err = template.ParseFiles(configuration.TemplatePath + `/PasswordChangedTemplate.txt`)
		if err != nil {
			validator.PushWarning(fmt.Errorf("error loading text template: %s ", err.Error()))
		}
	}
}

func validateSMTPNotifier(configuration *schema.SMTPNotifierConfiguration, validator *schema.StructValidator) {
	if configuration.StartupCheckAddress == "" {
		configuration.StartupCheckAddress = schema.DefaultSMTPNotifierConfiguration.StartupCheckAddress
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

	if configuration.Sender.Address == "" {
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
