package templates

import (
	"io"
	"time"
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

// Format of a template.
type Format int

// Formats.
const (
	DefaultFormat Format = iota
	HTMLFormat
	PlainTextFormat
)

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

// EmailEnvelopeValues are  the values used for the email envelopes.
type EmailEnvelopeValues struct {
	ProcessID    int
	UUID         string
	Host         string
	ServerName   string
	SenderDomain string
	Identifier   string
	From         string
	To           string
	Subject      string
	Date         time.Time
}
