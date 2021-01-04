package schema

// TLSConfig is a representation of the TLS configuration.
type TLSConfig struct {
	MinimumVersion string `mapstructure:"minimum_version"`
	SkipVerify     bool   `mapstructure:"skip_verify"`
	ServerName     string `mapstructure:"server_name"`
}
