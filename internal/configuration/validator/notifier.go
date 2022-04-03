package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/templates"
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

func validateNotifierTemplates(config *schema.NotifierConfiguration, validator *schema.StructValidator) {
	if config.TemplatePath == "" {
		return
	}

	var (
		err error
		t   *template.Template
	)

	_, err = os.Stat(config.TemplatePath)

	switch {
	case os.IsNotExist(err):
		validator.Push(fmt.Errorf(errFmtNotifierTemplatePathNotExist, config.TemplatePath))
		return
	case err != nil:
		validator.Push(fmt.Errorf(errFmtNotifierTemplatePathUnknownError, config.TemplatePath, err))
		return
	}

	if t, err = template.ParseFiles(filepath.Join(config.TemplatePath, templates.TemplateNameStep1+".html")); err == nil {
		templates.HTMLEmailTemplateStep1 = t
	} else {
		validator.PushWarning(fmt.Errorf(errFmtNotifierTemplateLoad, templates.TemplateNameStep1+".html", err))
	}

	if t, err = template.ParseFiles(filepath.Join(config.TemplatePath, templates.TemplateNameStep1+".txt")); err == nil {
		templates.PlainTextEmailTemplateStep1 = t
	} else {
		validator.PushWarning(fmt.Errorf(errFmtNotifierTemplateLoad, templates.TemplateNameStep1+".txt", err))
	}

	if t, err = template.ParseFiles(filepath.Join(config.TemplatePath, templates.TemplateNameStep2+".html")); err == nil {
		templates.HTMLEmailTemplateStep2 = t
	} else {
		validator.PushWarning(fmt.Errorf(errFmtNotifierTemplateLoad, templates.TemplateNameStep2+".html", err))
	}

	if t, err = template.ParseFiles(filepath.Join(config.TemplatePath, templates.TemplateNameStep2+".txt")); err == nil {
		templates.PlainTextEmailTemplateStep2 = t
	} else {
		validator.PushWarning(fmt.Errorf(errFmtNotifierTemplateLoad, templates.TemplateNameStep2+".txt", err))
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
