package schema

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Theme                 string `mapstructure:"theme"`
	CertificatesDirectory string `mapstructure:"certificates_directory"`
	JWTSecret             string `mapstructure:"jwt_secret"`
	DefaultRedirectionURL string `mapstructure:"default_redirection_url"`

	Host        string `koanf:"host"`          // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	Port        int    `koanf:"port"`          // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	TLSCert     string `koanf:"tls_cert"`      // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	TLSKey      string `koanf:"tls_key"`       // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	LogLevel    string `koanf:"log_level"`     // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	LogFormat   string `koanf:"log_format"`    // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	LogFilePath string `koanf:"log_file_path"` // Deprecated: To be Removed. TODO: Remove in 4.33.0.

	Logging               LogConfiguration                   `mapstructure:"log"`
	IdentityProviders     IdentityProvidersConfiguration     `mapstructure:"identity_providers"`
	AuthenticationBackend AuthenticationBackendConfiguration `mapstructure:"authentication_backend"`
	Session               SessionConfiguration               `mapstructure:"session"`
	TOTP                  *TOTPConfiguration                 `mapstructure:"totp"`
	DuoAPI                *DuoAPIConfiguration               `mapstructure:"duo_api"`
	AccessControl         AccessControlConfiguration         `mapstructure:"access_control"`
	Regulation            *RegulationConfiguration           `mapstructure:"regulation"`
	Storage               StorageConfiguration               `mapstructure:"storage"`
	Notifier              *NotifierConfiguration             `mapstructure:"notifier"`
	Server                ServerConfiguration                `mapstructure:"server"`
}
