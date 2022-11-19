package templates

import (
	th "html/template"
	"io"
	tt "text/template"
)

// Templates is the struct which holds all the *template.Template values.
type Templates struct {
	notification NotificationTemplates
}

// NotificationTemplates are the templates for the notification system.
type NotificationTemplates struct {
	passwordReset        *EmailTemplate
	identityVerification *EmailTemplate
}

// Template covers shared implementations between the text and html template.Template.
type Template interface {
	Execute(wr io.Writer, data any) error
	ExecuteTemplate(wr io.Writer, name string, data any) error
	Name() string
	DefinedTemplates() string
}

// Config for the Provider.
type Config struct {
	EmailTemplatesPath string
}

// EmailTemplate is the template type which contains both the html and txt versions of a template.
type EmailTemplate struct {
	HTML *th.Template
	Text *tt.Template
}

// EmailPasswordResetValues are the values used for password reset templates.
type EmailPasswordResetValues struct {
	UUID        string
	Title       string
	DisplayName string
	RemoteIP    string
}

// EmailIdentityVerificationValues are the values used for the identity verification templates.
type EmailIdentityVerificationValues struct {
	UUID        string
	Title       string
	DisplayName string
	RemoteIP    string
	LinkURL     string
	LinkText    string
}
