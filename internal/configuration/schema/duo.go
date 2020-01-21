package schema

// DuoAPIConfiguration represents the configuration related to Duo API.
type DuoAPIConfiguration struct {
	Hostname       string `mapstructure:"hostname"`
	IntegrationKey string `mapstructure:"integration_key"`
	SecretKey      string `mapstructure:"secret_key"`
}
