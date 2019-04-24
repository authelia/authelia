package schema

// FileSystemNotifierConfiguration represents the configuration of the notifier writing emails in a file.
type FileSystemNotifierConfiguration struct {
	Filename string `yaml:"filename"`
}

// EmailNotifierConfiguration represents the configuration of the email service notifier (like GMAIL API).
type EmailNotifierConfiguration struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Sender   string `yaml:"sender"`
	Service  string `yaml:"service"`
}

// SMTPNotifierConfiguration represents the configuration of the SMTP server to send emails with.
type SMTPNotifierConfiguration struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Secure   string `yaml:"secure"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Sender   string `yaml:"sender"`
}

// NotifierConfiguration representes the configuration of the notifier to use when sending notifications to users.
type NotifierConfiguration struct {
	FileSystem *FileSystemNotifierConfiguration `yaml:"filesystem"`
	Email      *EmailNotifierConfiguration      `yaml:"email"`
	SMTP       *SMTPNotifierConfiguration       `yaml:"smtp"`
}
