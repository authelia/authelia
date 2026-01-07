package schema

import (
	"crypto/tls"
	"net/mail"
	"net/url"
	"time"
)

// Notifier represents the configuration of the notifier to use when sending notifications to users.
type Notifier struct {
	DisableStartupCheck bool                `koanf:"disable_startup_check" yaml:"disable_startup_check" toml:"disable_startup_check" json:"disable_startup_check" jsonschema:"default=false,title=Disable Startup Check" jsonschema_description:"Disables the notifier startup checks."`
	FileSystem          *NotifierFileSystem `koanf:"filesystem" yaml:"filesystem,omitempty" toml:"filesystem,omitempty" json:"filesystem,omitempty" jsonschema:"title=File System" jsonschema_description:"The File System notifier."`
	SMTP                *NotifierSMTP       `koanf:"smtp" yaml:"smtp,omitempty" toml:"smtp,omitempty" json:"smtp,omitempty" jsonschema:"title=SMTP" jsonschema_description:"The SMTP notifier."`
	WebhookRef          string              `koanf:"webhook_ref" yaml:"webhook_ref,omitempty" toml:"webhook_ref,omitempty" json:"webhook_ref,omitempty" jsonschema:"title=Webhook Reference" jsonschema_description:"Reference to a webhook defined in definitions.webhooks."`
	TemplatePath        string              `koanf:"template_path" yaml:"template_path,omitempty" toml:"template_path,omitempty" json:"template_path,omitempty" jsonschema:"title=Template Path" jsonschema_description:"The path for notifier template overrides."`
}

// NotifierFileSystem represents the configuration of the notifier writing emails in a file.
type NotifierFileSystem struct {
	Filename string `koanf:"filename" yaml:"filename,omitempty" toml:"filename,omitempty" json:"filename,omitempty" jsonschema:"title=Filename" jsonschema_description:"The file path of the notifications."`
}

// NotifierSMTP represents the configuration of the SMTP server to send emails with.
type NotifierSMTP struct {
	Address             *AddressSMTP  `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"default=smtp://localhost:25,title=Address" jsonschema_description:"The SMTP server address."`
	Timeout             time.Duration `koanf:"timeout" yaml:"timeout,omitempty" toml:"timeout,omitempty" json:"timeout,omitempty" jsonschema:"default=5 seconds,title=Timeout" jsonschema_description:"The SMTP server connection timeout."`
	Username            string        `koanf:"username" yaml:"username,omitempty" toml:"username,omitempty" json:"username,omitempty" jsonschema:"title=Username" jsonschema_description:"The username for SMTP authentication."`
	Password            string        `koanf:"password" yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty" jsonschema:"title=Password" jsonschema_description:"The password for SMTP authentication."`
	Identifier          string        `koanf:"identifier" yaml:"identifier,omitempty" toml:"identifier,omitempty" json:"identifier,omitempty" jsonschema:"default=localhost,title=Identifier" jsonschema_description:"The identifier used during the HELO/EHLO command."`
	Sender              mail.Address  `koanf:"sender" yaml:"sender,omitempty" toml:"sender,omitempty" json:"sender,omitempty" jsonschema:"title=Sender" jsonschema_description:"The sender used for SMTP."`
	Subject             string        `koanf:"subject" yaml:"subject,omitempty" toml:"subject,omitempty" json:"subject,omitempty" jsonschema:"default=[Authelia] {title},title=Subject" jsonschema_description:"The subject format used."`
	StartupCheckAddress mail.Address  `koanf:"startup_check_address" yaml:"startup_check_address,omitempty" toml:"startup_check_address,omitempty" json:"startup_check_address,omitempty" jsonschema:"default=Authelia Test <test@authelia.com>,title=Startup Check Address" jsonschema_description:"The address used for the recipient in the startup check."`
	DisableRequireTLS   bool          `koanf:"disable_require_tls" yaml:"disable_require_tls" toml:"disable_require_tls" json:"disable_require_tls" jsonschema:"default=false,title=Disable Require TLS" jsonschema_description:"Disables the requirement to use TLS. This means security critical information and SMTP auth credentials will be sent in the clear due to an unencrypted connection. While this option exists, we heavily discourage it and do not support it."`
	DisableHTMLEmails   bool          `koanf:"disable_html_emails" yaml:"disable_html_emails" toml:"disable_html_emails" json:"disable_html_emails" jsonschema:"default=false,title=Disable HTML Emails" jsonschema_description:"Disables the mixed content type of emails and only sends the plaintext version."`
	DisableStartTLS     bool          `koanf:"disable_starttls" yaml:"disable_starttls" toml:"disable_starttls" json:"disable_starttls" jsonschema:"default=false,title=Disable StartTLS" jsonschema_description:"Disables the opportunistic StartTLS functionality which is useful for bad SMTP servers which advertise support for it but don't actually support it."`
	TLS                 *TLS          `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty" jsonschema:"title=TLS" jsonschema_description:"The SMTP server TLS connection properties."`

	// Deprecated: use address instead.
	Host string `koanf:"host" yaml:"host,omitempty" toml:"host,omitempty" json:"host,omitempty" jsonschema:"deprecated"`

	// Deprecated: use address instead.
	Port int `koanf:"port" yaml:"port" toml:"port" json:"port" jsonschema:"deprecated"`
}

// DefaultSMTPNotifierConfiguration represents default configuration parameters for the SMTP notifier.
var DefaultSMTPNotifierConfiguration = NotifierSMTP{
	Address:             &AddressSMTP{Address{true, false, -1, 25, nil, &url.URL{Scheme: AddressSchemeSMTP, Host: "localhost:25"}}},
	Timeout:             time.Second * 5,
	Subject:             "[Authelia] {title}",
	Identifier:          "localhost",
	StartupCheckAddress: mail.Address{Name: "Authelia Test", Address: "test@authelia.com"},
	TLS: &TLS{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultWebhookConfiguration represents default configuration parameters for webhooks.
var DefaultWebhookConfiguration = Webhook{
	Method:  "POST",
	Timeout: time.Second * 5,
}
