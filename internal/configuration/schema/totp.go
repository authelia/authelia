package schema

// TOTPConfiguration represents the configuration related to TOTP options.
type TOTPConfiguration struct {
	Issuer string `mapstructure:"issuer"`
	Period uint   `mapstructure:"period"`
	Skew   uint   `mapstrucutre:"skew"`
}
