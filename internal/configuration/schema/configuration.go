package schema

import (
	"net/url"
)

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Theme                 string `koanf:"theme" json:"theme" jsonschema:"default=light,enum=auto,enum=light,enum=dark,enum=grey,title=Theme Name" jsonschema_description:"The name of the theme to apply to the web UI."`
	CertificatesDirectory string `koanf:"certificates_directory" json:"certificates_directory" jsonschema:"title=Certificates Directory Path" jsonschema_description:"The path to a directory which is used to determine the certificates that are trusted."`
	Default2FAMethod      string `koanf:"default_2fa_method" json:"default_2fa_method" jsonschema:"enum=totp,enum=webauthn,enum=mobile_push,title=Default 2FA method" jsonschema_description:"When a user logs in for the first time this is the 2FA method configured for them."`

	Log                   Log                   `koanf:"log" json:"log" jsonschema:"title=Log" jsonschema_description:"Logging Configuration."`
	IdentityProviders     IdentityProviders     `koanf:"identity_providers" json:"identity_providers" jsonschema:"title=Identity Providers" jsonschema_description:"Identity Providers Configuration."`
	AuthenticationBackend AuthenticationBackend `koanf:"authentication_backend" json:"authentication_backend" jsonschema:"title=Authentication Backend" jsonschema_description:"Authentication Backend Configuration."`
	Session               Session               `koanf:"session" json:"session" jsonschema:"title=Session" jsonschema_description:"Session Configuration."`
	TOTP                  TOTP                  `koanf:"totp" json:"totp" jsonschema:"title=TOTP" jsonschema_description:"Time-based One-Time Password Configuration."`
	DuoAPI                DuoAPI                `koanf:"duo_api" json:"duo_api" jsonschema:"title=Duo API" jsonschema_description:"Duo API Configuration."`
	AccessControl         AccessControl         `koanf:"access_control" json:"access_control" jsonschema:"title=Access Control" jsonschema_description:"Access Control Configuration."`
	NTP                   NTP                   `koanf:"ntp" json:"ntp" jsonschema:"title=NTP" jsonschema_description:"Network Time Protocol Configuration."`
	Regulation            Regulation            `koanf:"regulation" json:"regulation" jsonschema:"title=Regulation" jsonschema_description:"Regulation Configuration."`
	Storage               Storage               `koanf:"storage" json:"storage" jsonschema:"title=Storage" jsonschema_description:"Storage Configuration."`
	Notifier              Notifier              `koanf:"notifier" json:"notifier" jsonschema:"title=Notifier" jsonschema_description:"Notifier Configuration."`
	Server                Server                `koanf:"server" json:"server" jsonschema:"title=Server" jsonschema_description:"Server Configuration."`
	Telemetry             Telemetry             `koanf:"telemetry" json:"telemetry" jsonschema:"title=Telemetry" jsonschema_description:"Telemetry Configuration."`
	WebAuthn              WebAuthn              `koanf:"webauthn" json:"webauthn" jsonschema:"title=WebAuthn" jsonschema_description:"WebAuthn Configuration."`
	PasswordPolicy        PasswordPolicy        `koanf:"password_policy" json:"password_policy" jsonschema:"title=Password Policy" jsonschema_description:"Password Policy Configuration."`
	PrivacyPolicy         PrivacyPolicy         `koanf:"privacy_policy" json:"privacy_policy" jsonschema:"title=Privacy Policy" jsonschema_description:"Privacy Policy Configuration."`
	IdentityValidation    IdentityValidation    `koanf:"identity_validation" json:"identity_validation" jsonschema:"title=Identity Validation" jsonschema_description:"Identity Validation Configuration."`
	Definitions           Definitions           `koanf:"definitions" json:"definitions" jsonschema:"title=Definitions" jsonschema_description:"Definitions for items reused elsewhere in the configuration."`

	// Deprecated: Use the session cookies option with the same name instead.
	DefaultRedirectionURL *url.URL `koanf:"default_redirection_url" json:"default_redirection_url" jsonschema:"deprecated,format=uri,title=The default redirection URL"`
}
