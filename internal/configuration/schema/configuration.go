package schema

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Theme                 string `koanf:"theme"`
	CertificatesDirectory string `koanf:"certificates_directory"`
	JWTSecret             string `koanf:"jwt_secret"`
	DefaultRedirectionURL string `koanf:"default_redirection_url"`
	Default2FAMethod      string `koanf:"default_2fa_method"`

	Log                   LogConfiguration               `koanf:"log"`
	IdentityProviders     IdentityProvidersConfiguration `koanf:"identity_providers"`
	AuthenticationBackend AuthenticationBackend          `koanf:"authentication_backend"`
	Session               SessionConfiguration           `koanf:"session"`
	TOTP                  TOTPConfiguration              `koanf:"totp"`
	DuoAPI                DuoAPIConfiguration            `koanf:"duo_api"`
	AccessControl         AccessControlConfiguration     `koanf:"access_control"`
	NTP                   NTPConfiguration               `koanf:"ntp"`
	Regulation            RegulationConfiguration        `koanf:"regulation"`
	Storage               StorageConfiguration           `koanf:"storage"`
	Notifier              NotifierConfiguration          `koanf:"notifier"`
	Server                ServerConfiguration            `koanf:"server"`
	Telemetry             TelemetryConfig                `koanf:"telemetry"`
	WebAuthn              WebAuthnConfiguration          `koanf:"webauthn"`
	PasswordPolicy        PasswordPolicyConfiguration    `koanf:"password_policy"`
	PrivacyPolicy         PrivacyPolicy                  `koanf:"privacy_policy"`
}
