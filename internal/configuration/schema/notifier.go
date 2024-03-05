package schema

import (
	"crypto/tls"
	"net/mail"
	"net/url"
	"time"
)

// Notifier represents the configuration of the notifier to use when sending notifications to users.
type Notifier struct {
	DisableStartupCheck bool                `koanf:"disable_startup_check" json:"disable_startup_check" jsonschema:"default=false,title=Disable Startup Check" jsonschema_description:"Disables the notifier startup checks."`
	FileSystem          *NotifierFileSystem `koanf:"filesystem" json:"filesystem" jsonschema:"title=File System" jsonschema_description:"The File System notifier."`
	SMTP                *NotifierSMTP       `koanf:"smtp" json:"smtp" jsonschema:"title=SMTP" jsonschema_description:"The SMTP notifier."`
	TemplatePath        string              `koanf:"template_path" json:"template_path" jsonschema:"title=Template Path" jsonschema_description:"The path for notifier template overrides."`
}

// NotifierFileSystem represents the configuration of the notifier writing emails in a file.
type NotifierFileSystem struct {
	Filename string `koanf:"filename" json:"filename" jsonschema:"title=Filename" jsonschema_description:"The file path of the notifications."`
}

// NotifierSMTP represents the configuration of the SMTP server to send emails with.
type NotifierSMTP struct {
	Address             *AddressSMTP  `koanf:"address" json:"address" jsonschema:"default=smtp://localhost:25,title=Address" jsonschema_description:"The SMTP server address."`
	Timeout             time.Duration `koanf:"timeout" json:"timeout" jsonschema:"default=5 seconds,title=Timeout" jsonschema_description:"The SMTP server connection timeout."`
	Username            string        `koanf:"username" json:"username" jsonschema:"title=Username" jsonschema_description:"The username for SMTP authentication."`
	Password            string        `koanf:"password" json:"password" jsonschema:"title=Password" jsonschema_description:"The password for SMTP authentication."`
	Identifier          string        `koanf:"identifier" json:"identifier" jsonschema:"default=localhost,title=Identifier" jsonschema_description:"The identifier used during the HELO/EHLO command."`
	Sender              mail.Address  `koanf:"sender" json:"sender" jsonschema:"title=Sender" jsonschema_description:"The sender used for SMTP."`
	Subject             string        `koanf:"subject" json:"subject" jsonschema:"default=[Authelia] {title},title=Subject" jsonschema_description:"The subject format used."`
	StartupCheckAddress mail.Address  `koanf:"startup_check_address" json:"startup_check_address" jsonschema:"default=Authelia Test <test@authelia.com>,title=Startup Check Address" jsonschema_description:"The address used for the recipient in the startup check."`
	DisableRequireTLS   bool          `koanf:"disable_require_tls" json:"disable_require_tls" jsonschema:"default=false,title=Disable Require TLS" jsonschema_description:"Disables the requirement to use TLS."`
	DisableHTMLEmails   bool          `koanf:"disable_html_emails" json:"disable_html_emails" jsonschema:"default=false,title=Disable HTML Emails" jsonschema_description:"Disables the mixed content type of emails and only sends the plaintext version."`
	DisableStartTLS     bool          `koanf:"disable_starttls" json:"disable_starttls" jsonschema:"default=false,title=Disable StartTLS" jsonschema_description:"Disables the opportunistic StartTLS functionality which is useful for bad SMTP servers which advertise support for it but don't actually support it."`
	TLS                 *TLS          `koanf:"tls" json:"tls" jsonschema:"title=TLS" jsonschema_description:"The SMTP server TLS connection properties."`

	// Deprecated: use address instead.
	Host string `koanf:"host" json:"host" jsonschema:"deprecated"`

	// Deprecated: use address instead.
	Port int `koanf:"port" json:"port" jsonschema:"deprecated"`
}

// DefaultSMTPNotifierConfiguration represents default configuration parameters for the SMTP notifier.
var DefaultSMTPNotifierConfiguration = NotifierSMTP{
	Address:             &AddressSMTP{Address{true, false, -1, 25, &url.URL{Scheme: AddressSchemeSMTP, Host: "localhost:25"}}},
	Timeout:             time.Second * 5,
	Subject:             "[Authelia] {title}",
	Identifier:          "localhost",
	StartupCheckAddress: mail.Address{Name: "Authelia Test", Address: "test@authelia.com"},
	TLS: &TLS{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}
