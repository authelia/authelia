package schema

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Port                  int                                `yaml:"port"`
	LogsLevel             string                             `yaml:"logs_level"`
	JWTSecret             string                             `yaml:"jwt_secret"`
	DefaultRedirectionURL string                             `yaml:"default_redirection_url"`
	AuthenticationBackend AuthenticationBackendConfiguration `yaml:"authentication_backend"`
	Session               SessionConfiguration               `yaml:"session"`

	TOTP          *TOTPConfiguration          `yaml:"totp"`
	DuoAPI        *DuoAPIConfiguration        `yaml:"duo_api"`
	AccessControl *AccessControlConfiguration `yaml:"access_control"`
	Regulation    *RegulationConfiguration    `yaml:"regulation"`
	Storage       *StorageConfiguration       `yaml:"storage"`
	Notifier      *NotifierConfiguration      `yaml:"notifier"`
}
