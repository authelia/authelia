package schema

// TOTPConfiguration represents the configuration related to TOTP options.
type TOTPConfiguration struct {
	Issuer    string `koanf:"issuer"`
	Algorithm string `koanf:"algorithm"`
	Period    int    `koanf:"period"`
	Skew      *int   `koanf:"skew"`
}

var defaultOtpSkew = 1

// DefaultTOTPConfiguration represents default configuration parameters for TOTP generation.
var DefaultTOTPConfiguration = TOTPConfiguration{
	Issuer:    "Authelia",
	Algorithm: TOTPAlgorithmSHA1,
	Period:    30,
	Skew:      &defaultOtpSkew,
}
