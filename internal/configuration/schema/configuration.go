package schema

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Host                  string `koanf:"host"`
	Port                  int    `koanf:"port"`
	Theme                 string `koanf:"theme"`
	TLSCert               string `koanf:"tls_cert"`
	TLSKey                string `koanf:"tls_key"`
	CertificatesDirectory string `koanf:"certificates_directory"`
	JWTSecret             string `koanf:"jwt_secret"`
	DefaultRedirectionURL string `koanf:"default_redirection_url"`

	// TODO: DEPRECATED START. Remove in 4.33.0.
	LogLevel    string `koanf:"log_level"`
	LogFormat   string `koanf:"log_format"`
	LogFilePath string `koanf:"log_file_path"`
	// TODO: DEPRECATED END. Remove in 4.33.0.

	Log                   LogConfiguration                   `koanf:"log"`
	IdentityProviders     IdentityProvidersConfiguration     `koanf:"identity_providers"`
	AuthenticationBackend AuthenticationBackendConfiguration `koanf:"authentication_backend"`
	Session               SessionConfiguration               `koanf:"session"`
	TOTP                  *TOTPConfiguration                 `koanf:"totp"`
	DuoAPI                *DuoAPIConfiguration               `koanf:"duo_api"`
	AccessControl         AccessControlConfiguration         `koanf:"access_control"`
	Regulation            *RegulationConfiguration           `koanf:"regulation"`
	Storage               StorageConfiguration               `koanf:"storage"`
	Notifier              *NotifierConfiguration             `koanf:"notifier"`
	Server                ServerConfiguration                `koanf:"server"`
}
