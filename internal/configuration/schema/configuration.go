package schema

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	TLSCert string `mapstructure:"tls_cert"`
	TLSKey  string `mapstructure:"tls_key"`

	LogLevel    string `mapstructure:"log_level"`
	LogFilePath string `mapstructure:"log_file_path"`

	// This secret is used by the identity validation process to forge JWT tokens
	// representing the permission to proceed with the operation.
	JWTSecret                 string `mapstructure:"jwt_secret"`
	DefaultRedirectionURL     string `mapstructure:"default_redirection_url"`
	GoogleAnalyticsTrackingID string `mapstructure:"google_analytics"`

	AuthenticationBackend AuthenticationBackendConfiguration `mapstructure:"authentication_backend"`
	Session               SessionConfiguration               `mapstructure:"session"`

	TOTP          *TOTPConfiguration         `mapstructure:"totp"`
	DuoAPI        *DuoAPIConfiguration       `mapstructure:"duo_api"`
	AccessControl AccessControlConfiguration `mapstructure:"access_control"`
	Regulation    *RegulationConfiguration   `mapstructure:"regulation"`
	Storage       StorageConfiguration       `mapstructure:"storage"`
	Notifier      *NotifierConfiguration     `mapstructure:"notifier"`
}
