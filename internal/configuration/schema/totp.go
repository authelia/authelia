package schema

// TOTPConfiguration represents the configuration related to TOTP options.
type TOTPConfiguration struct {
	Issuer string `mapstructure:"issuer"`
	Period int    `mapstructure:"period"`
	Skew   *int   `mapstructure:"skew"`
}

var defaultOtpSkew = 1
var DefaultTOTPConfiguration = TOTPConfiguration{
	Issuer: "Authelia",
	Period: 30,
	Skew:   &defaultOtpSkew,
}
