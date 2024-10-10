package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewFileSource returns a configuration.Source configured to load from a specified path. If there is an issue
// accessing this path it also returns an error.
func NewFileSource(path string) (source *FileSource) {
	return &FileSource{
		koanf: koanf.New(constDelimiter),
		path:  path,
	}
}

// NewFilteredFileSource returns a configuration.Source configured to load from a specified path. If there is
// an issue accessing this path it also returns an error.
func NewFilteredFileSource(path string, filters ...BytesFilter) (source *FileSource) {
	return &FileSource{
		koanf:   koanf.New(constDelimiter),
		path:    path,
		filters: filters,
	}
}

// NewFileSources returns a slice of configuration.Source configured to load from specified files.
func NewFileSources(paths []string) (sources []*FileSource) {
	for _, path := range paths {
		source := NewFileSource(path)

		sources = append(sources, source)
	}

	return sources
}

// NewFilteredFileSources returns a slice of configuration.Source configured to load from specified files.
func NewFilteredFileSources(paths []string, filters []BytesFilter) (sources []*FileSource) {
	for _, path := range paths {
		source := NewFilteredFileSource(path, filters...)

		sources = append(sources, source)
	}

	return sources
}

// Name of the Source.
func (s *FileSource) Name() (name string) {
	return fmt.Sprintf("file path(%s)", s.path)
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

	var info os.FileInfo

	if info, err = os.Stat(s.path); err != nil {
		return err
	}

	if info.IsDir() {
		return s.loadDir(val)
	}

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
		case extYML, extYAML:
			if err = s.koanf.Load(FilteredFileProvider(filepath.Join(s.path, name), s.filters...), yaml.Parser()); err != nil {
				return err
			}
		}
	}

	return nil
}

// ReadFiles reads all the files associated with this FileSource.
func (s *FileSource) ReadFiles() (files []*File, err error) {
	if s.path == "" {
		return nil, errors.New("invalid file path source configuration")
	}

	var info os.FileInfo

	if info, err = os.Stat(s.path); err != nil {
		return nil, err
	}

	if info.IsDir() {
		return s.readFilesDirectory(s.path)
	}

	return s.readFilesFile(s.path)
}

func (s *FileSource) readFilesFile(path string) (files []*File, err error) {
	var file *File

	if file, err = s.readFile(path); err != nil {
		return nil, err
	}

	return []*File{file}, nil
}

func (s *FileSource) readFile(path string) (file *File, err error) {
	file = &File{
		Path: path,
	}

	if file.Data, err = FilteredFileProvider(path, s.filters...).ReadBytes(); err != nil {
		return nil, err
	}

	return file, err
}

func (s *FileSource) readFilesDirectory(path string) (files []*File, err error) {
	var entries []os.DirEntry

	if entries, err = os.ReadDir(path); err != nil {
		return nil, err
	}

	var file *File

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		switch ext := filepath.Ext(name); ext {
		case extYML, extYAML:
			if file, err = s.readFile(filepath.Join(s.path, name)); err != nil {
				return nil, err
			}

			files = append(files, file)
		}
	}

	return files, nil
}

func (s *FileSource) GetBytesFilterNames() (names []string) {
	names = make([]string, len(s.filters))

	for i, filter := range s.filters {
		names[i] = filter.Name()
	}

	return names
}

// NewBytesSource returns a configuration.Source configured to load from a specified bytes.
func NewBytesSource(data []byte) (source *BytesSource) {
	return &BytesSource{
		koanf: koanf.New(constDelimiter),
		data:  data,
	}
}

// Name of the Source.
func (b *BytesSource) Name() (name string) {
	return fmt.Sprintf("bytes data(%s)", b.data)
}

// Merge the FileSource koanf.Koanf into the provided one.
func (b *BytesSource) Merge(ko *koanf.Koanf, _ *schema.StructValidator) (err error) {
	return ko.Merge(b.koanf)
}

// Load the Source into the FileSource koanf.Koanf.
func (b *BytesSource) Load(val *schema.StructValidator) (err error) {
	if b.data == nil {
		return nil
	}

	return b.koanf.Load(b, yaml.Parser())
}

// ReadBytes reads the contents of a file on disk, passes it through any configured filters, and returns the bytes.
func (b *BytesSource) ReadBytes() (data []byte, err error) {
	if len(b.filters) == 0 {
		return b.data, nil
	}

	data = make([]byte, len(b.data))

	copy(data, b.data)

	for _, filter := range b.filters {
		if data, err = filter.Filter(data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

// Read is not supported by the filtered file koanf.Provider.
func (b *BytesSource) Read() (map[string]any, error) {
	return nil, errors.New("filtered bytes provider does not support this method")
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
	keyMap, ignoredKeys := getEnvConfigMap(schema.Keys, s.prefix, s.delimiter, deprecations, deprecationsMKM)

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
	keyMap := getSecretConfigMap(schema.Keys, s.prefix, s.delimiter, deprecations)

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
func NewDefaultSources(paths []string, prefix, delimiter string, additionalSources ...Source) (sources []Source) {
	fileSources := NewFileSources(paths)
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

// NewDefaultSourcesFiltered returns a slice of Source configured to load from specified YAML files.
func NewDefaultSourcesFiltered(paths []string, filters []BytesFilter, prefix, delimiter string, additionalSources ...Source) (sources []Source) {
	fileSources := NewFilteredFileSources(paths, filters)
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
func NewDefaultSourcesWithDefaults(paths []string, filters []BytesFilter, prefix, delimiter string, defaults []Source, additionalSources ...Source) (sources []Source) {
	sources = []Source{NewMapSource(defaultsMapSource)}

	if len(defaults) != 0 {
		sources = append(sources, defaults...)
	}

	if len(filters) == 0 {
		sources = append(sources, NewDefaultSources(paths, prefix, delimiter, additionalSources...)...)
	} else {
		sources = append(sources, NewDefaultSourcesFiltered(paths, filters, prefix, delimiter, additionalSources...)...)
	}

	return sources
}

// NewDefaultsSource is a base configuration which sets defaults, this is particularly useful at the present time for
// setting defaults that otherwise can't be set. In the future it can be used to generate documentation or be generated
// by jsonschema. It will also reduce some areas of the validation package.
func NewDefaultsSource() (source Source) {
	return NewMapSource(mapDefaults)
}
