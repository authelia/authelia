package schema

// FileSystemNotifierConfiguration represents the configuration of the notifier writing emails in a file.
type FileSystemNotifierConfiguration struct {
	Filename string `mapstructure:"filename"`
}

// SMTPNotifierConfiguration represents the configuration of the SMTP server to send emails with.
type SMTPNotifierConfiguration struct {
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	Sender            string `mapstructure:"sender"`
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	PortTLS           bool   `mapstructure:"port_tls"`
	TrustedCert       string `mapstructure:"trusted_cert"`
	DisableVerifyCert bool   `mapstructure:"disable_verify_cert"`
	DisableRequireTLS bool   `mapstructure:"disable_require_tls"`
}

// NotifierConfiguration represents the configuration of the notifier to use when sending notifications to users.
type NotifierConfiguration struct {
	FileSystem *FileSystemNotifierConfiguration `mapstructure:"filesystem"`
	SMTP       *SMTPNotifierConfiguration       `mapstructure:"smtp"`
}
