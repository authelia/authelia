package schema

// TOTP represents the configuration related to TOTP options.
type TOTP struct {
	Disable    bool   `koanf:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the TOTP 2FA functionality"`
	Issuer     string `koanf:"issuer" json:"issuer" jsonschema:"default=Authelia,title=Issuer" jsonschema_description:"The issuer value for generated TOTP keys"`
	Algorithm  string `koanf:"algorithm" json:"algorithm" jsonschema:"default=SHA1,enum=SHA1,enum=SHA256,enum=SHA512,title=Algorithm" jsonschema_description:"The algorithm value for generated TOTP keys"`
	Digits     uint   `koanf:"digits" json:"digits" jsonschema:"default=6,enum=6,enum=8,title=Digits" jsonschema_description:"The digits value for generated TOTP keys"`
	Period     uint   `koanf:"period" json:"period" jsonschema:"default=30,title=Period" jsonschema_description:"The period value for generated TOTP keys"`
	Skew       *uint  `koanf:"skew" json:"skew" jsonschema:"default=1,title=Skew" jsonschema_description:"The permitted skew for generated TOTP keys"`
	SecretSize uint   `koanf:"secret_size" json:"secret_size" jsonschema:"default=32,minimum=20,title=Secret Size" jsonschema_description:"The secret size for generated TOTP keys"`
}

var defaultOtpSkew = uint(1)

// DefaultTOTPConfiguration represents default configuration parameters for TOTP generation.
var DefaultTOTPConfiguration = TOTP{
	Issuer:     "Authelia",
	Algorithm:  TOTPAlgorithmSHA1,
	Digits:     6,
	Period:     30,
	Skew:       &defaultOtpSkew,
	SecretSize: TOTPSecretSizeDefault,
}
