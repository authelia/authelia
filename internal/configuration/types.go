package configuration

import (
	"github.com/knadh/koanf"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// Provider holds the koanf.Koanf instance and the schema.Configuration.
type Provider struct {
	Koanf         *koanf.Koanf
	Validator     *schema.StructValidator
	Configuration *schema.Configuration

	validation bool
}

// Source is an abstract representation of a configuration Source implementation.
type Source interface {
	Name() (name string)
	Merge(ko *koanf.Koanf) (err error)
	Load() (err error)
	Validator() (validator *schema.StructValidator)
}

// YAMLFileSource is a configuration Source with a YAML File.
type YAMLFileSource struct {
	koanf *koanf.Koanf
	path  string
}

// EnvironmentSource is a configuration Source which loads values from the environment.
type EnvironmentSource struct {
	koanf *koanf.Koanf
}

// SecretsSource loads environment variables that have a value pointing to a file.
type SecretsSource struct {
	koanf     *koanf.Koanf
	validator *schema.StructValidator
}
