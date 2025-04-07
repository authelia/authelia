package schema

// DuoAPI represents the configuration related to Duo API.
type DuoAPI struct {
	Disable              bool   `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disable the Duo API integration."`
	Hostname             string `koanf:"hostname" yaml:"hostname,omitempty" toml:"hostname,omitempty" json:"hostname,omitempty" jsonschema:"format=hostname,title=Hostname" jsonschema_description:"The Hostname provided by your Duo API dashboard."`
	IntegrationKey       string `koanf:"integration_key" yaml:"integration_key,omitempty" toml:"integration_key,omitempty" json:"integration_key,omitempty" jsonschema:"title=Integration Key" jsonschema_description:"The Integration Key provided by your Duo API dashboard."`
	SecretKey            string `koanf:"secret_key" yaml:"secret_key,omitempty" toml:"secret_key,omitempty" json:"secret_key,omitempty" jsonschema:"title=Secret Key" jsonschema_description:"The Secret Key provided by your Duo API dashboard."`
	EnableSelfEnrollment bool   `koanf:"enable_self_enrollment" yaml:"enable_self_enrollment" toml:"enable_self_enrollment" json:"enable_self_enrollment" jsonschema:"default=false,title=Enable Self Enrollment" jsonschema_description:"Enable the Self Enrollment flow."`
}
