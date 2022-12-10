package configuration

import (
	"github.com/knadh/koanf"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Source is an abstract representation of a configuration.Source implementation.
type Source interface {
	Name() (name string)
	Merge(ko *koanf.Koanf, val *schema.StructValidator) (err error)
	Load(val *schema.StructValidator) (err error)
}

// FileSource is a YAML file configuration.Source.
type FileSource struct {
	koanf     *koanf.Koanf
	path      string
	directory bool
	filters   []FileFilter
}

// EnvironmentSource is a configuration.Source which loads values from the environment.
type EnvironmentSource struct {
	koanf     *koanf.Koanf
	prefix    string
	delimiter string
}

// SecretsSource is a configuration.Source which loads environment variables that have a value pointing to a file.
type SecretsSource struct {
	koanf     *koanf.Koanf
	prefix    string
	delimiter string
}

// CommandLineSource is a configuration.Source which loads configuration from the command line flags.
type CommandLineSource struct {
	koanf    *koanf.Koanf
	flags    *pflag.FlagSet
	callback func(flag *pflag.Flag) (string, any)
}

// MapSource is a configuration.Source which loads configuration from a map[string]any.
type MapSource struct {
	m     map[string]any
	koanf *koanf.Koanf
}
