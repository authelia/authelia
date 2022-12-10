package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewFileSource returns a configuration.Source configured to load from a specified YAML path. If there is an issue
// accessing this path it also returns an error.
func NewFileSource(path string) (source *FileSource) {
	return &FileSource{
		koanf: koanf.New(constDelimiter),
		path:  path,
	}
}

// NewFilteredFileSource returns a configuration.Source configured to load from a specified YAML path. If there is
// an issue accessing this path it also returns an error.
func NewFilteredFileSource(path string, filters ...FileFilter) (source *FileSource) {
	return &FileSource{
		koanf:   koanf.New(constDelimiter),
		path:    path,
		filters: filters,
	}
}

// NewDirectorySource returns a configuration.Source configured to load from a specified directory. If there is an issue
// accessing this path it also returns an error.
func NewDirectorySource(path string) (source *FileSource) {
	return &FileSource{
		koanf:     koanf.New(constDelimiter),
		directory: true,
		path:      path,
	}
}

// NewFilteredDirectorySource returns a configuration.Source configured to load from a specified directory path. If
// there is an issue accessing this path it also returns an error.
func NewFilteredDirectorySource(path string, filters ...FileFilter) (source *FileSource) {
	return &FileSource{
		koanf:     koanf.New(constDelimiter),
		path:      path,
		directory: true,
		filters:   filters,
	}
}

// NewFileSources returns a slice of configuration.Source configured to load from specified YAML files.
func NewFileSources(paths []string) (sources []*FileSource) {
	for _, path := range paths {
		source := NewFileSource(path)

		sources = append(sources, source)
	}

	return sources
}

// NewFilteredFileSources returns a slice of configuration.Source configured to load from specified YAML files.
func NewFilteredFileSources(paths []string, filters []FileFilter) (sources []*FileSource) {
	for _, path := range paths {
		source := NewFilteredFileSource(path, filters...)

		sources = append(sources, source)
	}

	return sources
}

// Name of the Source.
func (s *FileSource) Name() (name string) {
	return fmt.Sprintf("yaml file(%s)", s.path)
}

// Merge the FileSource koanf.Koanf into the provided one.
func (s *FileSource) Merge(ko *koanf.Koanf, _ *schema.StructValidator) (err error) {
	return ko.Merge(s.koanf)
}

// Load the Source into the FileSource koanf.Koanf.
func (s *FileSource) Load(val *schema.StructValidator) (err error) {
	if s.path == "" {
		return errors.New("invalid file path source configuration")
	}

	if s.directory {
		return s.loadDir(val)
	}

	return s.load(val)
}

func (s *FileSource) load(_ *schema.StructValidator) (err error) {
	return s.koanf.Load(FilteredFileProvider(s.path, s.filters...), yaml.Parser())
}

func (s *FileSource) loadDir(_ *schema.StructValidator) (err error) {
	var entries []os.DirEntry

	if entries, err = os.ReadDir(s.path); err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		switch ext := filepath.Ext(name); ext {
		case ".yml", ".yaml":
			if err = s.koanf.Load(FilteredFileProvider(filepath.Join(s.path, name), s.filters...), yaml.Parser()); err != nil {
				return err
			}
		}
	}

	return nil
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

// Load the Source into the FileSource koanf.Koanf.
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

// Load the Source into the FileSource koanf.Koanf.
func (s *MapSource) Load(_ *schema.StructValidator) (err error) {
	return s.koanf.Load(confmap.Provider(s.m, constDelimiter), nil)
}

// NewDefaultSources returns a slice of Source configured to load from specified YAML files.
func NewDefaultSources(filePaths []string, directory string, prefix, delimiter string, additional Sources) (sources []Source) {
	if len(additional.Pre) != 0 {
		copy(sources, additional.Pre)
	}

	fileSources := NewFileSources(filePaths)
	for _, source := range fileSources {
		sources = append(sources, source)
	}

	if directory != "" {
		sources = append(sources, NewDirectorySource(directory))
	}

	sources = append(sources, NewEnvironmentSource(prefix, delimiter))
	sources = append(sources, NewSecretsSource(prefix, delimiter))

	if len(additional.Post) != 0 {
		sources = append(sources, additional.Post...)
	}

	return sources
}

// NewDefaultSourcesFiltered returns a slice of Source configured to load from specified YAML files.
func NewDefaultSourcesFiltered(files []string, directory string, filters []FileFilter, prefix, delimiter string, additional Sources) (sources []Source) {
	if len(additional.Pre) != 0 {
		copy(sources, additional.Pre)
	}

	fileSources := NewFilteredFileSources(files, filters)
	for _, source := range fileSources {
		sources = append(sources, source)
	}

	if directory != "" {
		sources = append(sources, NewFilteredDirectorySource(directory, filters...))
	}

	sources = append(sources, NewEnvironmentSource(prefix, delimiter))
	sources = append(sources, NewSecretsSource(prefix, delimiter))

	if len(additional.Post) != 0 {
		sources = append(sources, additional.Post...)
	}

	return sources
}
