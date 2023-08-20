package schema

import (
	"crypto/tls"
	"net/mail"
	"net/url"
	"time"
)

// FileSystemNotifierConfiguration represents the configuration of the notifier writing emails in a file.
type FileSystemNotifierConfiguration struct {
	Filename string `koanf:"filename"`
}

// SMTPNotifierConfiguration represents the configuration of the SMTP server to send emails with.
type SMTPNotifierConfiguration struct {
	Address             *AddressSMTP  `koanf:"address"`
	Timeout             time.Duration `koanf:"timeout"`
	Username            string        `koanf:"username"`
	Password            string        `koanf:"password"`
	Identifier          string        `koanf:"identifier"`
	Sender              mail.Address  `koanf:"sender"`
	Subject             string        `koanf:"subject"`
	StartupCheckAddress mail.Address  `koanf:"startup_check_address"`
	DisableRequireTLS   bool          `koanf:"disable_require_tls"`
	DisableHTMLEmails   bool          `koanf:"disable_html_emails"`
	DisableStartTLS     bool          `koanf:"disable_starttls"`
	TLS                 *TLSConfig    `koanf:"tls"`

	// Deprecated: use address instead.
	Host string `koanf:"host"`

	// Deprecated: use address instead.
	Port int `koanf:"port"`
}

// NotifierConfiguration represents the configuration of the notifier to use when sending notifications to users.
type NotifierConfiguration struct {
	DisableStartupCheck bool                             `koanf:"disable_startup_check"`
	FileSystem          *FileSystemNotifierConfiguration `koanf:"filesystem"`
	SMTP                *SMTPNotifierConfiguration       `koanf:"smtp"`
	TemplatePath        string                           `koanf:"template_path"`
}

// DefaultSMTPNotifierConfiguration represents default configuration parameters for the SMTP notifier.
var DefaultSMTPNotifierConfiguration = SMTPNotifierConfiguration{
	Address:             &AddressSMTP{Address{true, false, -1, 25, &url.URL{Scheme: AddressSchemeSMTP, Host: "localhost:25"}}},
	Timeout:             time.Second * 5,
	Subject:             "[Authelia] {title}",
	Identifier:          "localhost",
	StartupCheckAddress: mail.Address{Name: "Authelia Test", Address: "test@authelia.com"},
	TLS: &TLSConfig{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}
