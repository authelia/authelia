package schema

// TLSConfiguration is a representation of the TLS configuration.
type TLSConfiguration struct {
	MinimumVersion string `mapstructure:"minimum_version"`
	SkipVerify     bool   `mapstructure:"skip_verify"`
	ServerName     string `mapstructure:"server_name"`
}

// PluginConfiguration is a Authelia Plugin configuration.
type PluginConfiguration struct {
	Name string `mapstructure:"name"`
}
