package configuration

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"

	"github.com/authelia/authelia/internal/configuration/validator"
)

// NewDefaultSources returns a slice of Source configured to load from specified YAML files.
func NewDefaultSources(filePaths []string, provider *Provider) (sources []Source) {
	sources = NewYAMLFileSources(filePaths)

	sources = append(sources, NewEnvironmentSource())
	sources = append(sources, NewSecretsSource(provider))

	return sources
}

// NewYAMLFileSource returns a Source configured to load from a specified YAML file. If there is an issue accessing this
// file it also returns an error.
func NewYAMLFileSource(path string) (source Source) {
	return Source{
		Name:     fmt.Sprintf("file:%s", path),
		Provider: file.Provider(path),
		Parser:   yaml.Parser(),
	}
}

// NewYAMLFileSources returns a slice of Source configured to load from specified YAML files.
func NewYAMLFileSources(paths []string) (sources []Source) {
	for _, path := range paths {
		source := NewYAMLFileSource(path)

		sources = append(sources, source)
	}

	return sources
}

// NewEnvironmentSource returns a Source configured to load from environment variables.
func NewEnvironmentSource() (source Source) {
	keyMap, ignoredKeys := getEnvConfigMap(validator.ValidKeys)

	return Source{
		Name:     "environment",
		Provider: env.ProviderWithValue(envPrefix, delimiter, koanfEnvironmentCallback(keyMap, ignoredKeys)),
		Parser:   nil,
	}
}

// NewSecretsSource returns a Source configured to load from secrets.
func NewSecretsSource(provider *Provider) (source Source) {
	keyMap := getSecretConfigMap(validator.ValidKeys)

	return Source{
		Name:     "secrets",
		Provider: env.ProviderWithValue(envPrefixAlt, delimiter, koanfEnvironmentSecretsCallback(keyMap, provider)),
		Parser:   nil,
	}
}
