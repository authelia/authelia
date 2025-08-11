package templates

import (
	"fmt"
	th "html/template"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	tt "text/template"
	"time"
)

const (
	envPrefix  = "AUTHELIA_"
	envXPrefix = "X_AUTHELIA_"
)

// IMPORTANT: This is a copy of github.com/authelia/authelia/internal/configuration's secretSuffixes except all uppercase.
// Make sure you update these at the same time.
var envSecretSuffixes = []string{
	"KEY", "SECRET", "PASSWORD", "TOKEN", "CERTIFICATE_CHAIN",
}

func isSecretEnvKey(key string) (isSecretEnvKey bool) {
	key = strings.ToUpper(key)

	if !strings.HasPrefix(key, envPrefix) && !strings.HasPrefix(key, envXPrefix) {
		return false
	}

	for _, s := range envSecretSuffixes {
		suffix := strings.ToUpper(s)

		if strings.HasSuffix(key, suffix) {
			return true
		}
	}

	return false
}

func fileExists(path string) (exists bool) {
	info, err := os.Stat(path)

	return err == nil && !info.IsDir()
}

func readTemplate(name, ext, category, overridePath string) (tPath string, embed bool, data []byte, err error) {
	if overridePath != "" {
		tPath = filepath.Join(overridePath, name+ext)

		if fileExists(tPath) {
			if data, err = os.ReadFile(tPath); err != nil {
				return tPath, false, nil, fmt.Errorf("failed to read template override at path '%s': %w", tPath, err)
			}

			return tPath, false, data, nil
		}
	}

	tPath = path.Join("embed", category, name+ext)

	if data, err = embedFS.ReadFile(tPath); err != nil {
		return tPath, true, nil, fmt.Errorf("failed to read embedded template '%s': %w", tPath, err)
	}

	return tPath, true, data, nil
}

func parseTextTemplate(name, tPath string, embed bool, data []byte) (t *tt.Template, err error) {
	if t, err = tt.New(name + extText).Funcs(FuncMap()).Parse(string(data)); err != nil {
		if embed {
			return nil, fmt.Errorf("failed to parse embedded template '%s': %w", tPath, err)
		}

		return nil, fmt.Errorf("failed to parse template override at path '%s': %w", tPath, err)
	}

	return t, nil
}

func parseHTMLTemplate(name, tPath string, embed bool, data []byte) (t *th.Template, err error) {
	if t, err = th.New(name + extHTML).Funcs(FuncMap()).Parse(string(data)); err != nil {
		if embed {
			return nil, fmt.Errorf("failed to parse embedded template '%s': %w", tPath, err)
		}

		return nil, fmt.Errorf("failed to parse template override at path '%s': %w", tPath, err)
	}

	return t, nil
}

func loadEmailTemplate(name, overridePath string) (t *EmailTemplate, err error) {
	var (
		embed bool
		tpath string
		data  []byte
	)

	t = &EmailTemplate{}

	if tpath, embed, data, err = readTemplate(name, extText, TemplateCategoryNotifications, overridePath); err != nil {
		return nil, fmt.Errorf("error occurred reading text template: %w", err)
	}

	if t.Text, err = parseTextTemplate(name, tpath, embed, data); err != nil {
		return nil, fmt.Errorf("error occurred parsing text template: %w", err)
	}

	if tpath, embed, data, err = readTemplate(name, extHTML, TemplateCategoryNotifications, overridePath); err != nil {
		return nil, fmt.Errorf("error occurred reading html template: %w", err)
	}

	if t.HTML, err = parseHTMLTemplate(name, tpath, embed, data); err != nil {
		return nil, fmt.Errorf("error occurred parsing html template: %w", err)
	}

	return t, nil
}

func strval(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func strslice(v any) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []any:
		b := make([]string, 0, len(v))

		for _, s := range v {
			if s != nil {
				b = append(b, strval(s))
			}
		}

		return b
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			l := val.Len()
			b := make([]string, 0, l)

			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				if value != nil {
					b = append(b, strval(value))
				}
			}

			return b
		default:
			if v == nil {
				return []string{}
			}

			return []string{strval(v)}
		}
	}
}

func formatHTMLTimeWithLocation(date time.Time, location *time.Location) string {
	return formatTimeWithLocation(time.DateOnly, date, location)
}

func formatTimeWithLocation(format string, date time.Time, location *time.Location) string {
	return date.In(location).Format(format)
}

func convertAnyToTime(date any) (t time.Time) {
	switch v := date.(type) {
	case time.Time:
		t = v
	case *time.Time:
		t = *v
	case int64:
		t = time.Unix(v, 0)
	case int:
		t = time.Unix(int64(v), 0)
	case int32:
		t = time.Unix(int64(v), 0)
	default:
		t = time.Now()
	}

	return t
}
