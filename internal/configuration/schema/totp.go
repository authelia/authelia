package schema

// TOTPConfiguration represents the configuration related to TOTP options.
type TOTPConfiguration struct {
	Issuer string `mapstructure:"issuer"`
	Period int    `mapstructure:"period"`
	Skew   *int   `mapstructure:"skew"`
}
