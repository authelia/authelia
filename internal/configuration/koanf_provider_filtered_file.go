package configuration

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/templates"
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

type TemplateBytesFilter struct {
	t   *template.Template
	log *logrus.Entry
}

func (f *TemplateBytesFilter) Name() (name string) {
	return filterTemplate
}

func (f *TemplateBytesFilter) Filter(in []byte) (out []byte, err error) {
	if f.t, err = f.t.Parse(string(in)); err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}

	if err = f.t.Execute(buf, nil); err != nil {
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
		NewTemplateFileFilter(),
	}
}

// NewFileFilters returns a list of BytesFilter provided they are valid.
func NewFileFilters(names []string) (filters []BytesFilter, err error) {
	filters = make([]BytesFilter, len(names))

	filterMap := map[string]int{}

	for i, name := range names {
		name = strings.ToLower(name)

		switch name {
		case filterTemplate:
			filters[i] = NewTemplateFileFilter()
		case filterExpandEnv:
			filters[i] = NewExpandEnvFileFilter()
		default:
			return nil, fmt.Errorf("invalid filter named '%s'", name)
		}

		if _, ok := filterMap[name]; ok {
			return nil, fmt.Errorf("duplicate filter named '%s'", name)
		} else {
			filterMap[name] = 1
		}
	}

	return filters, nil
}

// NewExpandEnvFileFilter returns a new BytesFilter which passes the bytes through os.Expand using special env vars.
func NewExpandEnvFileFilter() BytesFilter {
	return &ExpandEnvBytesFilter{
		log: logging.Logger().WithFields(map[string]any{filterField: filterExpandEnv}),
	}
}

// NewTemplateFileFilter returns a new BytesFilter which passes the bytes through text/template.
func NewTemplateFileFilter() BytesFilter {
	return &TemplateBytesFilter{
		log: logging.Logger().WithFields(map[string]any{filterField: filterTemplate}),
		t:   template.New("config.template").Funcs(templates.FuncMap()),
	}
}
