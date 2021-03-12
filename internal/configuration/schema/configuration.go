package schema

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Host                  string `mapstructure:"host"`
	Port                  int    `mapstructure:"port"`
	Theme                 string `mapstructure:"theme"`
	TLSCert               string `mapstructure:"tls_cert"`
	TLSKey                string `mapstructure:"tls_key"`
	CertificatesDirectory string `mapstructure:"certificates_directory"`
	LogLevel              string `mapstructure:"log_level"`
	LogFormat             string `mapstructure:"log_format"`
	LogFilePath           string `mapstructure:"log_file_path"`
	JWTSecret             string `mapstructure:"jwt_secret"`
	DefaultRedirectionURL string `mapstructure:"default_redirection_url"`

	OAuth                 OAuthConfiguration                 `mapstructure:"oauth"`
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
