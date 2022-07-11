package templates

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

func templateExists(path string) (exists bool) {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	return true
}

//nolint:unparam
func loadTemplate(name, category, overridePath string) (t *template.Template, err error) {
	if overridePath != "" {
		tPath := filepath.Join(overridePath, name)

		if templateExists(tPath) {
			if t, err = template.ParseFiles(tPath); err != nil {
				return nil, fmt.Errorf("could not parse template at path '%s': %w", tPath, err)
			}

			return t, nil
		}
	}

	data, err := embedFS.ReadFile(path.Join("src", category, name))
	if err != nil {
		return nil, err
	}

	if t, err = template.New(name).Parse(string(data)); err != nil {
		panic(fmt.Errorf("failed to parse internal template: %w", err))
	}

	return t, nil
}
