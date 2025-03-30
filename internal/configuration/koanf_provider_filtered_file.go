package configuration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pelletier/go-toml/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// FilteredFile implements a koanf.Provider.
type FilteredFile struct {
	data    []byte
	path    string
	filters []BytesFilter
}

// FilteredFileProvider returns a koanf.Provider which provides filtered file output.
func FilteredFileProvider(path string, filters ...BytesFilter) *FilteredFile {
	return &FilteredFile{
		path:    filepath.Clean(path),
		filters: filters,
	}
}

// ReadBytes reads the contents of a file on disk, passes it through any configured filters, and returns the bytes.
func (f *FilteredFile) ReadBytes() (data []byte, err error) {
	if f.data != nil {
		return f.data, nil
	}

	if f.data, err = os.ReadFile(f.path); err != nil {
		return nil, err
	}

	if len(f.data) == 0 || len(f.filters) == 0 {
		return f.data, nil
	}

	for _, filter := range f.filters {
		if f.data, err = filter.Filter(f.data); err != nil {
			return nil, err
		}
	}

	return f.data, nil
}

// Read is not supported by the filtered file koanf.Provider.
func (f *FilteredFile) Read() (map[string]any, error) {
	return nil, errors.New("filtered file provider does not support this method")
}

// BytesFilter is an interface describing utility structures that filter bytes into desired bytes.
type BytesFilter interface {
	Name() (name string)
	Filter(in []byte) (out []byte, err error)
}

type ExpandEnvBytesFilter struct {
	log *logrus.Entry
}

func (f *ExpandEnvBytesFilter) Name() (name string) {
	return filterExpandEnv
}

func (f *ExpandEnvBytesFilter) Filter(in []byte) (out []byte, err error) {
	out = []byte(os.Expand(string(in), templates.FuncGetEnv))

	if f.log.Level >= logrus.TraceLevel {
		f.log.
			WithField("content", base64.RawStdEncoding.EncodeToString(out)).
			Trace("Expanded Env File Filter completed successfully")
	}

	return out, nil
}

// TemplateBytesFilterValues holds the values provided to the templates filter.
type TemplateBytesFilterValues struct {
	Values   map[string]any
	Authelia map[string]any
}

// TemplateBytesFilter is a templates bytes filter, it allows filtering bytes through the go template engine.
type TemplateBytesFilter struct {
	t    *template.Template
	log  *logrus.Entry
	data TemplateBytesFilterValues
}

func (f *TemplateBytesFilter) Name() (name string) {
	return filterTemplate
}

func (f *TemplateBytesFilter) Filter(in []byte) (out []byte, err error) {
	var t *template.Template

	if t, err = f.t.Parse(string(in)); err != nil {
		return nil, err
	}

	f.t = t

	buf := &bytes.Buffer{}

	if err = f.t.Execute(buf, f.data); err != nil {
		return nil, err
	}

	out = buf.Bytes()

	if f.log.Level >= logrus.TraceLevel {
		f.log.
			WithField("content", base64.RawStdEncoding.EncodeToString(out)).
			Trace("Templated File Filter completed successfully")
	}

	return out, nil
}

// NewFileFiltersDefault returns the default list of BytesFilter.
func NewFileFiltersDefault() []BytesFilter {
	return []BytesFilter{
		NewExpandEnvFileFilter(),
		NewTemplateFileFilter(nil),
	}
}

// NewFileFilters returns a list of BytesFilter provided they are valid. Each path in valuesFiles is loaded in order and
// deep-merged over the previously-loaded values so that later files override earlier ones.
func NewFileFilters(valuesFiles []string, names ...string) (filters []BytesFilter, err error) {
	filters = make([]BytesFilter, len(names))

	var values map[string]any

	if values, err = loadValuesFiles(valuesFiles); err != nil {
		return nil, err
	}

	filterMap := map[string]int{}

	for i, name := range names {
		name = strings.ToLower(name)

		switch name {
		case filterTemplate:
			filters[i] = NewTemplateFileFilter(values)
		case filterExpandEnv:
			filters[i] = NewExpandEnvFileFilter()
		default:
			return nil, fmt.Errorf("invalid filter named '%s'", name)
		}

		if _, ok := filterMap[name]; ok {
			return nil, fmt.Errorf("duplicate filter named '%s'", name)
		}

		filterMap[name] = 1
	}

	return filters, nil
}

// loadValuesFiles loads each file in order and deep-merges them onto a single values map. Later files override earlier
// ones at every level. An empty or nil slice returns nil values.
func loadValuesFiles(paths []string) (values map[string]any, err error) {
	if len(paths) == 0 {
		return nil, nil
	}

	values = map[string]any{}

	for _, path := range paths {
		var loaded map[string]any

		if loaded, err = loadValuesFile(path); err != nil {
			return nil, err
		}

		mergeValues(values, loaded)
	}

	return values, nil
}

// loadValuesFile reads and parses a values file. The format is selected from the file extension: .yml/.yaml for YAML,
// .json for JSON, .toml for TOML. An empty path returns nil values; an unsupported extension returns an error.
func loadValuesFile(path string) (values map[string]any, err error) {
	if path == "" {
		return nil, nil
	}

	var data []byte

	if data, err = os.ReadFile(path); err != nil {
		return nil, fmt.Errorf("error reading values file: %w", err)
	}

	values = map[string]any{}

	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case extYML, extYAML:
		err = yaml.Unmarshal(data, &values)
	case extJSON:
		err = json.Unmarshal(data, &values)
	case extTOML:
		err = toml.Unmarshal(data, &values)
	default:
		return nil, fmt.Errorf("error parsing values file: unsupported extension '%s': must be one of '.yml', '.yaml', '.json', or '.toml'", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing values file: %w", err)
	}

	return values, nil
}

// mergeValues deep-merges src into dst. When both values for a key are themselves maps, they are recursively merged;
// otherwise the src value replaces the dst value. dst is mutated in place.
func mergeValues(dst, src map[string]any) {
	for k, sv := range src {
		dv, ok := dst[k]
		if !ok {
			dst[k] = sv

			continue
		}

		dmap, dok := dv.(map[string]any)
		smap, sok := sv.(map[string]any)

		if dok && sok {
			mergeValues(dmap, smap)

			continue
		}

		dst[k] = sv
	}
}

// NewExpandEnvFileFilter returns a new BytesFilter which passes the bytes through os.Expand using special env vars.
func NewExpandEnvFileFilter() BytesFilter {
	return &ExpandEnvBytesFilter{
		log: logging.Logger().WithFields(map[string]any{filterField: filterExpandEnv}),
	}
}

// NewTemplateFileFilter returns a new BytesFilter which passes the bytes through text/template.
func NewTemplateFileFilter(values map[string]any) BytesFilter {
	data := TemplateBytesFilterValues{
		Values:   values,
		Authelia: map[string]any{},
	}

	if data.Values == nil {
		data.Values = map[string]any{}
	}

	data.Authelia["Version"] = utils.Version()
	data.Authelia["Build"] = map[string]any{
		"Tag":    utils.BuildTag,
		"State":  utils.BuildState,
		"Extra":  utils.BuildExtra,
		"Date":   utils.BuildDate,
		"Commit": utils.BuildCommit,
		"Branch": utils.BuildBranch,
		"Number": utils.BuildNumber,
	}

	return &TemplateBytesFilter{
		log:  logging.Logger().WithFields(map[string]any{filterField: filterTemplate}),
		t:    template.New("config.template").Funcs(templates.FuncMap()),
		data: data,
	}
}
