package schema

// TOTPConfiguration represents the configuration related to TOTP options.
type TOTPConfiguration struct {
	Issuer    string `mapstructure:"issuer"`
	Period    int    `mapstructure:"period"`
	Skew      *int   `mapstructure:"skew"`
	Algorithm string `mapstructure:"algorithm"`
}

var defaultOtpSkew = 1

// DefaultTOTPConfiguration represents default configuration parameters for TOTP generation.
var DefaultTOTPConfiguration = TOTPConfiguration{
	Issuer:    "Authelia",
	Period:    30,
	Skew:      &defaultOtpSkew,
	Algorithm: "sha1",
}
