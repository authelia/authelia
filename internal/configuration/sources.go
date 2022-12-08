package configuration

import (
	"errors"
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewYAMLFileSource returns a configuration.Source configured to load from a specified YAML path. If there is an issue
// accessing this path it also returns an error.
func NewYAMLFileSource(path string) (source *YAMLFileSource) {
	return &YAMLFileSource{
		koanf: koanf.New(constDelimiter),
		path:  path,
	}
}

// NewYAMLFileTemplatedSource returns a configuration.Source configured to load from a specified YAML path. If there is
// an issue accessing this path it also returns an error.
func NewYAMLFileTemplatedSource(path string, filters ...FileFilterFunc) (source *YAMLFileSource) {
	return &YAMLFileSource{
		koanf:   koanf.New(constDelimiter),
		path:    path,
		filters: filters,
	}
}

// NewYAMLFileSources returns a slice of configuration.Source configured to load from specified YAML files.
func NewYAMLFileSources(paths []string) (sources []*YAMLFileSource) {
	for _, path := range paths {
		source := NewYAMLFileSource(path)

		sources = append(sources, source)
	}

	return sources
}

// NewYAMLFileTemplatedSources returns a slice of configuration.Source configured to load from specified YAML files.
func NewYAMLFileTemplatedSources(paths, filters []string) (sources []*YAMLFileSource) {
	filterFuncs := make([]FileFilterFunc, len(filters))

	for i, filter := range filters {
		switch filter {
		case "template":
			filterFuncs[i] = NewTemplateFileFilter()
		case "expand-env":
			filterFuncs[i] = NewExpandEnvFileFilter()
		}
	}

	for _, path := range paths {
		source := NewYAMLFileTemplatedSource(path, filterFuncs...)

		sources = append(sources, source)
	}

	return sources
}

// Name of the Source.
func (s *YAMLFileSource) Name() (name string) {
	return fmt.Sprintf("yaml file(%s)", s.path)
}

// Merge the YAMLFileSource koanf.Koanf into the provided one.
func (s *YAMLFileSource) Merge(ko *koanf.Koanf, _ *schema.StructValidator) (err error) {
	return ko.Merge(s.koanf)
}

// Load the Source into the YAMLFileSource koanf.Koanf.
func (s *YAMLFileSource) Load(_ *schema.StructValidator) (err error) {
	if s.path == "" {
		return errors.New("invalid yaml path source configuration")
	}

	return s.koanf.Load(FilteredFileProvider(s.path, s.filters...), yaml.Parser())
}

// NewEnvironmentSource returns a Source configured to load from environment variables.
func NewEnvironmentSource(prefix, delimiter string) (source *EnvironmentSource) {
	return &EnvironmentSource{
		koanf:     koanf.New(constDelimiter),
		prefix:    prefix,
		delimiter: delimiter,
	}
}

// Name of the Source.
func (s *EnvironmentSource) Name() (name string) {
	return "environment"
}

// Merge the EnvironmentSource koanf.Koanf into the provided one.
func (s *EnvironmentSource) Merge(ko *koanf.Koanf, _ *schema.StructValidator) (err error) {
	return ko.Merge(s.koanf)
}

// Load the Source into the EnvironmentSource koanf.Koanf.
func (s *EnvironmentSource) Load(_ *schema.StructValidator) (err error) {
	keyMap, ignoredKeys := getEnvConfigMap(schema.Keys, s.prefix, s.delimiter)

	return s.koanf.Load(env.ProviderWithValue(s.prefix, constDelimiter, koanfEnvironmentCallback(keyMap, ignoredKeys, s.prefix, s.delimiter)), nil)
}

// NewSecretsSource returns a Source configured to load from secrets.
func NewSecretsSource(prefix, delimiter string) (source *SecretsSource) {
	return &SecretsSource{
		koanf:     koanf.New(constDelimiter),
		prefix:    prefix,
		delimiter: delimiter,
	}
}

// Name of the Source.
func (s *SecretsSource) Name() (name string) {
	return "secrets"
}

// Merge the SecretsSource koanf.Koanf into the provided one.
func (s *SecretsSource) Merge(ko *koanf.Koanf, val *schema.StructValidator) (err error) {
	for _, key := range s.koanf.Keys() {
		value, ok := ko.Get(key).(string)

		if ok && value != "" {
			val.Push(fmt.Errorf(errFmtSecretAlreadyDefined, key))
		}
	}

	return ko.Merge(s.koanf)
}

// Load the Source into the SecretsSource koanf.Koanf.
func (s *SecretsSource) Load(val *schema.StructValidator) (err error) {
	keyMap := getSecretConfigMap(schema.Keys, s.prefix, s.delimiter)

	return s.koanf.Load(env.ProviderWithValue(s.prefix, constDelimiter, koanfEnvironmentSecretsCallback(keyMap, val)), nil)
}

// NewCommandLineSourceWithMapping creates a new command line configuration source with a map[string]string which converts
// flag names into other config key names. If includeValidKeys is true we also allow any flag with a name which matches
// the list of valid keys into the koanf.Koanf, otherwise everything not in the map is skipped. Unchanged flags are also
// skipped unless includeUnchangedKeys is set to true.
func NewCommandLineSourceWithMapping(flags *pflag.FlagSet, mapping map[string]string, includeValidKeys, includeUnchangedKeys bool) (source *CommandLineSource) {
	return &CommandLineSource{
		koanf:    koanf.New(constDelimiter),
		flags:    flags,
		callback: koanfCommandLineWithMappingCallback(mapping, includeValidKeys, includeUnchangedKeys),
	}
}

// Name of the Source.
func (s *CommandLineSource) Name() (name string) {
	return "command-line"
}

// Merge the CommandLineSource koanf.Koanf into the provided one.
func (s *CommandLineSource) Merge(ko *koanf.Koanf, val *schema.StructValidator) (err error) {
	return ko.Merge(s.koanf)
}

// Load the Source into the YAMLFileSource koanf.Koanf.
func (s *CommandLineSource) Load(_ *schema.StructValidator) (err error) {
	if s.callback != nil {
		return s.koanf.Load(posflag.ProviderWithFlag(s.flags, ".", s.koanf, s.callback), nil)
	}

	return s.koanf.Load(posflag.Provider(s.flags, ".", s.koanf), nil)
}

// NewMapSource returns a new map[string]any source.
func NewMapSource(m map[string]any) (source *MapSource) {
	return &MapSource{
		m:     m,
		koanf: koanf.New(constDelimiter),
	}
}

// Name of the Source.
func (s *MapSource) Name() (name string) {
	return "map"
}

// Merge the CommandLineSource koanf.Koanf into the provided one.
func (s *MapSource) Merge(ko *koanf.Koanf, val *schema.StructValidator) (err error) {
	return ko.Merge(s.koanf)
}

// Load the Source into the YAMLFileSource koanf.Koanf.
func (s *MapSource) Load(_ *schema.StructValidator) (err error) {
	return s.koanf.Load(confmap.Provider(s.m, constDelimiter), nil)
}

// NewDefaultSources returns a slice of Source configured to load from specified YAML files.
func NewDefaultSources(filePaths []string, prefix, delimiter string, additionalSources ...Source) (sources []Source) {
	fileSources := NewYAMLFileSources(filePaths)
	for _, source := range fileSources {
		sources = append(sources, source)
	}

	sources = append(sources, NewEnvironmentSource(prefix, delimiter))
	sources = append(sources, NewSecretsSource(prefix, delimiter))

	if len(additionalSources) != 0 {
		sources = append(sources, additionalSources...)
	}

	return sources
}

// NewDefaultSourcesExperimental returns a slice of Source configured to load from specified YAML files.
func NewDefaultSourcesExperimental(filePaths, filters []string, prefix, delimiter string, additionalSources ...Source) (sources []Source) {
	fileSources := NewYAMLFileTemplatedSources(filePaths, filters)
	for _, source := range fileSources {
		sources = append(sources, source)
	}

	sources = append(sources, NewEnvironmentSource(prefix, delimiter))
	sources = append(sources, NewSecretsSource(prefix, delimiter))

	if len(additionalSources) != 0 {
		sources = append(sources, additionalSources...)
	}

	return sources
}

// NewDefaultSourcesWithDefaults returns a slice of Source configured to load from specified YAML files with additional sources.
func NewDefaultSourcesWithDefaults(filePaths []string, prefix, delimiter string, defaults Source, additionalSources ...Source) (sources []Source) {
	sources = []Source{defaults}

	sources = append(sources, NewDefaultSources(filePaths, prefix, delimiter, additionalSources...)...)

	return sources
}
