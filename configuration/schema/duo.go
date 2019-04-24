package schema

// DuoAPIConfiguration represents the configuration related to Duo API.
type DuoAPIConfiguration struct {
	Hostname       string `yaml:"hostname"`
	IntegrationKey string `yaml:"integration_key"`
	SecretKey      string `yaml:"secret_key"`
}
