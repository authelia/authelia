// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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

	filterMap := map[string]int{}

	for i, name := range names {
		name = strings.ToLower(name)

		switch name {
		case "template":
			filters[i] = NewTemplateFileFilter()
		case "expand-env":
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

// NewExpandEnvFileFilter is a FileFilter which passes the bytes through os.ExpandEnv.
func NewExpandEnvFileFilter() FileFilter {
	log := logging.Logger()

	return func(in []byte) (out []byte, err error) {
		out = []byte(os.ExpandEnv(string(in)))

		if log.Level >= logrus.TraceLevel {
			log.
				WithField("content", base64.RawStdEncoding.EncodeToString(out)).
				Trace("Expanded Env File Filter completed successfully")
		}

		return out, nil
	}
}

// NewTemplateFileFilter is a FileFilter which passes the bytes through text/template.
func NewTemplateFileFilter() FileFilter {
	t := template.New("config.template").Funcs(templates.FuncMap())

	log := logging.Logger()

	return func(in []byte) (out []byte, err error) {
		if t, err = t.Parse(string(in)); err != nil {
			return nil, err
		}

		buf := &bytes.Buffer{}

		if err = t.Execute(buf, nil); err != nil {
			return nil, err
		}

		out = buf.Bytes()

		if log.Level >= logrus.TraceLevel {
			log.
				WithField("content", base64.RawStdEncoding.EncodeToString(out)).
				Trace("Templated File Filter completed successfully")
		}

		return out, nil
	}
}
