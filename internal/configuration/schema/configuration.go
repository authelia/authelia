package schema

import (
	"net/url"
)

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Theme                  string `koanf:"theme" yaml:"theme,omitempty" toml:"theme,omitempty" json:"theme,omitempty" jsonschema:"default=light,enum=auto,enum=light,enum=dark,enum=grey,enum=oled,title=Theme Name" jsonschema_description:"The name of the theme to apply to the web UI."`
	PortalTemplate         string `koanf:"portal_template" yaml:"portal_template,omitempty" toml:"portal_template,omitempty" json:"portal_template,omitempty" jsonschema:"title=Portal Template Name" jsonschema_description:"The name of the login portal template to load from branding assets."`
	PortalHeadline         string `koanf:"portal_headline" yaml:"portal_headline,omitempty" toml:"portal_headline,omitempty" json:"portal_headline,omitempty" jsonschema:"title=Portal Headline" jsonschema_description:"Overrides the primary headline displayed on the login portal."`
	PortalSubtitle         string `koanf:"portal_subtitle" yaml:"portal_subtitle,omitempty" toml:"portal_subtitle,omitempty" json:"portal_subtitle,omitempty" jsonschema:"title=Portal Subtitle" jsonschema_description:"Overrides the subtitle displayed beneath the login portal headline."`
	PortalTemplateSwitcher bool   `koanf:"portal_template_switcher" yaml:"portal_template_switcher,omitempty" toml:"portal_template_switcher,omitempty" json:"portal_template_switcher,omitempty" jsonschema:"default=false,title=Enable Portal Template Switcher" jsonschema_description:"If true, exposes a palette control in the login UI for switching between available templates."`
	CertificatesDirectory  string `koanf:"certificates_directory" yaml:"certificates_directory,omitempty" toml:"certificates_directory,omitempty" json:"certificates_directory,omitempty" jsonschema:"title=Certificates Directory Path" jsonschema_description:"The path to a directory which is used to determine the certificates that are trusted."`
	Default2FAMethod       string `koanf:"default_2fa_method" yaml:"default_2fa_method,omitempty" toml:"default_2fa_method,omitempty" json:"default_2fa_method,omitempty" jsonschema:"enum=totp,enum=webauthn,enum=mobile_push,title=Default 2FA method" jsonschema_description:"When a user logs in for the first time this is the 2FA method configured for them."`

	Log                   Log                   `koanf:"log" yaml:"log,omitempty" toml:"log,omitempty" json:"log,omitempty" jsonschema:"title=Log" jsonschema_description:"Logging Configuration."`
	IdentityProviders     IdentityProviders     `koanf:"identity_providers" yaml:"identity_providers,omitempty" toml:"identity_providers,omitempty" json:"identity_providers,omitempty" jsonschema:"title=Identity Providers" jsonschema_description:"Identity Providers Configuration."`
	AuthenticationBackend AuthenticationBackend `koanf:"authentication_backend" yaml:"authentication_backend,omitempty" toml:"authentication_backend,omitempty" json:"authentication_backend,omitempty" jsonschema:"title=Authentication Backend" jsonschema_description:"Authentication Backend Configuration."`
	Session               Session               `koanf:"session" yaml:"session,omitempty" toml:"session,omitempty" json:"session,omitempty" jsonschema:"title=Session" jsonschema_description:"Session Configuration."`
	TOTP                  TOTP                  `koanf:"totp" yaml:"totp,omitempty" toml:"totp,omitempty" json:"totp,omitempty" jsonschema:"title=TOTP" jsonschema_description:"Time-based One-Time Password Configuration."`
	DuoAPI                DuoAPI                `koanf:"duo_api" yaml:"duo_api,omitempty" toml:"duo_api,omitempty" json:"duo_api,omitempty" jsonschema:"title=Duo API" jsonschema_description:"Duo API Configuration."`
	AccessControl         AccessControl         `koanf:"access_control" yaml:"access_control,omitempty" toml:"access_control,omitempty" json:"access_control,omitempty" jsonschema:"title=Access Control" jsonschema_description:"Access Control Configuration."`
	NTP                   NTP                   `koanf:"ntp" yaml:"ntp,omitempty" toml:"ntp,omitempty" json:"ntp,omitempty" jsonschema:"title=NTP" jsonschema_description:"Network Time Protocol Configuration."`
	Regulation            Regulation            `koanf:"regulation" yaml:"regulation,omitempty" toml:"regulation,omitempty" json:"regulation,omitempty" jsonschema:"title=Regulation" jsonschema_description:"Regulation Configuration."`
	Storage               Storage               `koanf:"storage" yaml:"storage,omitempty" toml:"storage,omitempty" json:"storage,omitempty" jsonschema:"title=Storage" jsonschema_description:"Storage Configuration."`
	Notifier              Notifier              `koanf:"notifier" yaml:"notifier,omitempty" toml:"notifier,omitempty" json:"notifier,omitempty" jsonschema:"title=Notifier" jsonschema_description:"Notifier Configuration."`
	Server                Server                `koanf:"server" yaml:"server,omitempty" toml:"server,omitempty" json:"server,omitempty" jsonschema:"title=Server" jsonschema_description:"Server Configuration."`
	Telemetry             Telemetry             `koanf:"telemetry" yaml:"telemetry,omitempty" toml:"telemetry,omitempty" json:"telemetry,omitempty" jsonschema:"title=Telemetry" jsonschema_description:"Telemetry Configuration."`
	WebAuthn              WebAuthn              `koanf:"webauthn" yaml:"webauthn,omitempty" toml:"webauthn,omitempty" json:"webauthn,omitempty" jsonschema:"title=WebAuthn" jsonschema_description:"WebAuthn Configuration."`
	PasswordPolicy        PasswordPolicy        `koanf:"password_policy" yaml:"password_policy,omitempty" toml:"password_policy,omitempty" json:"password_policy,omitempty" jsonschema:"title=Password Policy" jsonschema_description:"Password Policy Configuration."`
	PrivacyPolicy         PrivacyPolicy         `koanf:"privacy_policy" yaml:"privacy_policy,omitempty" toml:"privacy_policy,omitempty" json:"privacy_policy,omitempty" jsonschema:"title=Privacy Policy" jsonschema_description:"Privacy Policy Configuration."`
	IdentityValidation    IdentityValidation    `koanf:"identity_validation" yaml:"identity_validation,omitempty" toml:"identity_validation,omitempty" json:"identity_validation,omitempty" jsonschema:"title=Identity Validation" jsonschema_description:"Identity Validation Configuration."`
	Definitions           Definitions           `koanf:"definitions" yaml:"definitions,omitempty" toml:"definitions,omitempty" json:"definitions,omitempty" jsonschema:"title=Definitions" jsonschema_description:"Definitions for items reused elsewhere in the configuration."`

	// Deprecated: Use the session cookies option with the same name instead.
	DefaultRedirectionURL *url.URL `koanf:"default_redirection_url" yaml:"default_redirection_url,omitempty" toml:"default_redirection_url,omitempty" json:"default_redirection_url,omitempty" jsonschema:"deprecated,format=uri,title=The default redirection URL"`
}
