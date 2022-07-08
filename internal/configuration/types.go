package configuration

import (
	"github.com/knadh/koanf"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Source is an abstract representation of a configuration Source implementation.
type Source interface {
	Name() (name string)
	Merge(ko *koanf.Koanf, val *schema.StructValidator) (err error)
	Load(val *schema.StructValidator) (err error)
}

// YAMLFileSource is a configuration Source with a YAML File.
type YAMLFileSource struct {
	koanf *koanf.Koanf
	path  string
}

// EnvironmentSource is a configuration Source which loads values from the environment.
type EnvironmentSource struct {
	koanf     *koanf.Koanf
	prefix    string
	delimiter string
}

// SecretsSource loads environment variables that have a value pointing to a file.
type SecretsSource struct {
	koanf     *koanf.Koanf
	prefix    string
	delimiter string
}

// CommandLineSource loads configuration from the command line flags.
type CommandLineSource struct {
	koanf    *koanf.Koanf
	flags    *pflag.FlagSet
	callback func(flag *pflag.Flag) (string, interface{})
}

// MapSource loads configuration from the command line flags.
type MapSource struct {
	m     map[string]interface{}
	koanf *koanf.Koanf
}
