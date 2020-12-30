package schema

// FileSystemNotifierConfiguration represents the configuration of the notifier writing emails in a file.
type FileSystemNotifierConfiguration struct {
	Filename string `mapstructure:"filename"`
}

// SMTPNotifierConfiguration represents the configuration of the SMTP server to send emails with.
type SMTPNotifierConfiguration struct {
	Host                string     `mapstructure:"host"`
	Port                int        `mapstructure:"port"`
	Username            string     `mapstructure:"username"`
	Password            string     `mapstructure:"password"`
	Identifier          string     `mapstructure:"identifier"`
	Sender              string     `mapstructure:"sender"`
	Subject             string     `mapstructure:"subject"`
	StartupCheckAddress string     `mapstructure:"startup_check_address"`
	DisableRequireTLS   bool       `mapstructure:"disable_require_tls"`
	DisableHTMLEmails   bool       `mapstructure:"disable_html_emails"`
	TLS                 *TLSConfig `mapstructure:"tls"`
	TrustedCert         string     `mapstructure:"trusted_cert"`        // Deprecated: Replaced with Global Option CertificatesDirectory. TODO: Remove in 4.28.
	DisableVerifyCert   *bool      `mapstructure:"disable_verify_cert"` // Deprecated: Replaced with LDAPAuthenticationBackendConfiguration.TLS.SkipVerify. TODO: Remove in 4.28.
}

// NotifierConfiguration represents the configuration of the notifier to use when sending notifications to users.
type NotifierConfiguration struct {
	DisableStartupCheck bool                             `mapstructure:"disable_startup_check"`
	FileSystem          *FileSystemNotifierConfiguration `mapstructure:"filesystem"`
	SMTP                *SMTPNotifierConfiguration       `mapstructure:"smtp"`
}

// DefaultSMTPNotifierConfiguration represents default configuration parameters for the SMTP notifier.
var DefaultSMTPNotifierConfiguration = SMTPNotifierConfiguration{
	Subject:    "[Authelia] {title}",
	Identifier: "localhost",
	TLS: &TLSConfig{
		MinimumVersion: "TLS1.2",
	},
}
