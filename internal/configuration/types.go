package configuration

import (
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Source is an abstract representation of a configuration configuration.Source implementation.
type Source interface {
	Name() (name string)
	Merge(ko *koanf.Koanf, val *schema.StructValidator) (err error)
	Load(val *schema.StructValidator) (err error)
}

// FileSource is a file configuration.Source.
type FileSource struct {
	koanf     *koanf.Koanf
	provider  *FilteredFile
	providers map[string]*FilteredFile
	path      string
	filters   []BytesFilter
}

// BytesSource is a raw bytes configuration.Source.
type BytesSource struct {
	koanf   *koanf.Koanf
	data    []byte
	filters []BytesFilter
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

// MapSource is a configuration.Source which loads configuration from the command line flags.
type MapSource struct {
	m     map[string]any
	koanf *koanf.Koanf
}

// File represents a file path and data content as bytes.
type File struct {
	Path string
	Data []byte
}
