package templates

import (
	th "html/template"
	"io"
	tt "text/template"
)

// Templates is the struct which holds all the *template.Template values.
type Templates struct {
	notification NotificationTemplates
	asset        AssetTemplates
	oidc         OpenIDConnectTemplates
}

type OpenIDConnectTemplates struct {
	formpost *th.Template
}

// AssetTemplates are templates for specific key assets.
type AssetTemplates struct {
	index *tt.Template
	api   OpenAPIAssetTemplates
}

// OpenAPIAssetTemplates are asset templates for the OpenAPI specification.
type OpenAPIAssetTemplates struct {
	index *tt.Template
	spec  *tt.Template
}

// NotificationTemplates are the templates for the notification system.
type NotificationTemplates struct {
	jwtIdentityVerification *EmailTemplate
	otcIdentityVerification *EmailTemplate
	event                   *EmailTemplate
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

// EmailEventValues are the values used for event templates.
type EmailEventValues struct {
	Title       string
	DisplayName string
	Details     map[string]any
	RemoteIP    string
}

// EmailPasswordResetValues are the values used for password reset templates.
type EmailPasswordResetValues struct {
	Title       string
	DisplayName string
	RemoteIP    string
}

// EmailIdentityVerificationJWTValues are the values used for the identity verification JWT templates.
type EmailIdentityVerificationJWTValues struct {
	Title              string
	DisplayName        string
	RemoteIP           string
	LinkURL            string
	LinkText           string
	RevocationLinkURL  string
	RevocationLinkText string
}

// EmailIdentityVerificationOTCValues are the values used for the identity verification OTP templates.
type EmailIdentityVerificationOTCValues struct {
	Title              string
	DisplayName        string
	RemoteIP           string
	OneTimeCode        string
	RevocationLinkURL  string
	RevocationLinkText string
}
