package configuration

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// FilteredFile implements a koanf.Provider.
type FilteredFile struct {
	path    string
	filters []FileFilter
}

// FilteredFileProvider returns a koanf.Provider which provides filtered file output.
func FilteredFileProvider(path string, filters ...FileFilter) *FilteredFile {
	return &FilteredFile{
		path:    filepath.Clean(path),
		filters: filters,
	}
}

// ReadBytes reads the contents of a file on disk, passes it through any configured filters, and returns the bytes.
func (f *FilteredFile) ReadBytes() (data []byte, err error) {
	if data, err = os.ReadFile(f.path); err != nil {
		return nil, err
	}

	if len(data) == 0 || len(f.filters) == 0 {
		return data, nil
	}

	for _, filter := range f.filters {
		if data, err = filter(data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

// Read is not supported by the filtered file koanf.Provider.
func (f *FilteredFile) Read() (map[string]interface{}, error) {
	return nil, errors.New("filtered file provider does not support this method")
}

// FileFilter describes a func used to filter files.
type FileFilter func(in []byte) (out []byte, err error)

// NewFileFiltersDefault returns the default list of FileFilter.
func NewFileFiltersDefault() []FileFilter {
	return []FileFilter{
		NewTemplateFileFilter(),
		NewExpandEnvFileFilter(),
	}
}

// NewFileFilters returns a list of FileFilter provided they are valid.
func NewFileFilters(names []string) (filters []FileFilter, err error) {
	filters = make([]FileFilter, len(names))

	var hasTemplate, hasExpandEnv bool

	for i, name := range names {
		switch name {
		case "template":
			if hasTemplate {
				return nil, fmt.Errorf("duplicate filter named '%s'", name)
			}

			hasTemplate = true

			filters[i] = NewTemplateFileFilter()
		case "expand-env":
			if hasExpandEnv {
				return nil, fmt.Errorf("duplicate filter named '%s'", name)
			}

			hasExpandEnv = true

			filters[i] = NewExpandEnvFileFilter()
		default:
			return nil, fmt.Errorf("invalid filter named '%s'", name)
		}
	}

	return filters, nil
}

// NewExpandEnvFileFilter is a FileFilter which passes the bytes through os.ExpandEnv.
func NewExpandEnvFileFilter() FileFilter {
	return func(in []byte) (out []byte, err error) {
		return []byte(os.ExpandEnv(string(in))), nil
	}
}

// NewTemplateFileFilter is a FileFilter which passes the bytes through text/template.
func NewTemplateFileFilter() FileFilter {
	data := &templated{
		Env: map[string]string{},
	}

	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)

		if len(kv) != 2 {
			continue
		}

		data.Env[kv[0]] = kv[1]
	}

	var t *template.Template

	t = template.New("config.template").Funcs(map[string]any{
		"env": newTemplatedEnvFunc(data),
	})

	return func(in []byte) (out []byte, err error) {
		if t, err = t.Parse(string(in)); err != nil {
			return nil, err
		}

		buf := &bytes.Buffer{}

		if err = t.Execute(buf, data); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}
}

type templated struct {
	Env map[string]string
}

func newTemplatedEnvFunc(data *templated) func(key string) (value string, err error) {
	return func(key string) (value string, err error) {
		var ok bool

		if value, ok = data.Env[key]; !ok {
			return "", fmt.Errorf("environment variable %s does not exist", key)
		}

		return value, nil
	}
}
