package schema

// TOTPConfiguration represents the configuration related to TOTP options.
type TOTPConfiguration struct {
	Disable          bool   `koanf:"disable"`
	Issuer           string `koanf:"issuer"`
	DefaultAlgorithm string `koanf:"algorithm"`
	DefaultDigits    int    `koanf:"digits"`
	DefaultPeriod    int    `koanf:"period"`
	Skew             *int   `koanf:"skew"`
	SecretSize       int    `koanf:"secret_size"`

	AllowedAlgorithms []string `koanf:"allowed_algorithms"`
	AllowedDigits     []int    `koanf:"allowed_digits"`
	AllowedPeriods    []int    `koanf:"allowed_periods"`
}

var defaultTOTPSkew = 1

// DefaultTOTPConfiguration represents default configuration parameters for TOTP generation.
var DefaultTOTPConfiguration = TOTPConfiguration{
	Issuer:           "Authelia",
	DefaultAlgorithm: TOTPAlgorithmSHA1,
	DefaultDigits:    6,
	DefaultPeriod:    30,
	Skew:             &defaultTOTPSkew,
	SecretSize:       TOTPSecretSizeDefault,
}
