package schema

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Theme                 string `koanf:"theme"`
	CertificatesDirectory string `koanf:"certificates_directory"`
	JWTSecret             string `koanf:"jwt_secret"`
	DefaultRedirectionURL string `koanf:"default_redirection_url"`

	Log                   LogConfiguration                   `koanf:"log"`
	IdentityProviders     IdentityProvidersConfiguration     `koanf:"identity_providers"`
	AuthenticationBackend AuthenticationBackendConfiguration `koanf:"authentication_backend"`
	TOTP                  TOTPConfiguration                  `koanf:"totp"`
	Webauthn              WebauthnConfiguration              `koanf:"webauthn"`
	DuoAPI                *DuoAPIConfiguration               `koanf:"duo_api"`
	AccessControl         AccessControlConfiguration         `koanf:"access_control"`
	Regulation            RegulationConfiguration            `koanf:"regulation"`

	Server         ServerConfiguration          `koanf:"server"`
	Session        SessionConfiguration         `koanf:"session"`
	NTP            NTPConfiguration             `koanf:"ntp"`
	Storage        StorageConfiguration         `koanf:"storage"`
	Notifier       *NotifierConfiguration       `koanf:"notifier"`
	PasswordPolicy *PasswordPolicyConfiguration `koanf:"password_policy"`
}
