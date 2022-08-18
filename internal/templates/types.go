package templates

import (
	"text/template"
	"time"
)

// Templates is the struct which holds all the *template.Template values.
type Templates struct {
	notification NotificationTemplates
}

// NotificationTemplates are the templates for the notification system.
type NotificationTemplates struct {
	envelope             *template.Template
	passwordReset        HTMLPlainTextTemplate
	identityVerification HTMLPlainTextTemplate
}

// Format of a template.
type Format int

// Formats.
const (
	DefaultFormat Format = iota
	HTMLFormat
	PlainTextFormat
)

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
	Boundary     string
	Body         string
}
