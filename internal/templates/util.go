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

func readTemplate(name, category, overridePath string) (tPath string, embed bool, data []byte, err error) {
	if overridePath != "" {
		tPath = filepath.Join(overridePath, name)

		if templateExists(tPath) {
			if data, err = os.ReadFile(tPath); err != nil {
				return tPath, false, nil, fmt.Errorf("failed to read template override at path '%s': %w", tPath, err)
			}

			return tPath, false, data, nil
		}
	}

	tPath = path.Join("src", category, name)

	if data, err = embedFS.ReadFile(tPath); err != nil {
		return tPath, true, nil, fmt.Errorf("failed to read embedded template '%s': %w", tPath, err)
	}

	return tPath, true, data, nil
}

func parseTemplate(name, tPath string, embed bool, data []byte) (t *template.Template, err error) {
	if t, err = template.New(name).Parse(string(data)); err != nil {
		if embed {
			return nil, fmt.Errorf("failed to parse embedded template '%s': %w", tPath, err)
		}

		return nil, fmt.Errorf("failed to parse template override at path '%s': %w", tPath, err)
	}

	return t, nil
}

//nolint:unparam
func loadTemplate(name, category, overridePath string) (t *template.Template, err error) {
	var (
		embed bool
		tPath string
		data  []byte
	)

	if tPath, embed, data, err = readTemplate(name, category, overridePath); err != nil {
		return nil, err
	}

	return parseTemplate(name, tPath, embed, data)
}
