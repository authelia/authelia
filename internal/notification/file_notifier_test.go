package notification

import (
	"bufio"
	"context"
	"net/mail"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestFileNotifier_StartupCheck(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func(base string) string
		expectErr  bool
		verifyFile func(t *testing.T, path string)
	}{
		{
			name: "ShouldReturnErrorWhenParentIsFile",
			setup: func(base string) string {
				parent := filepath.Join(base, "notadir")
				require.NoError(t, os.WriteFile(parent, []byte("x"), 0o600))
				return filepath.Join(parent, "notify.log")
			},
			expectErr:  true,
			verifyFile: func(t *testing.T, path string) {},
		},
		{
			name: "ShouldSucceedWhenParentIsDir",
			setup: func(base string) string {
				parent := filepath.Join(base, "adir")
				require.NoError(t, os.MkdirAll(parent, 0o755))
				return filepath.Join(parent, "notify.log")
			},
			expectErr: false,
			verifyFile: func(t *testing.T, path string) {
				info, err := os.Stat(filepath.Dir(path))
				require.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "ShouldCreateParentDirectoryWhenMissing",
			setup: func(base string) string {
				return filepath.Join(base, "nested", "dir", "notify.log")
			},
			expectErr: false,
			verifyFile: func(t *testing.T, path string) {
				parent := filepath.Dir(path)
				info, err := os.Stat(parent)
				require.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			path := tc.setup(base)
			n := NewFileNotifier(schema.NotifierFileSystem{Filename: path})

			err := n.StartupCheck()

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				tc.verifyFile(t, path)
			}
		})
	}
}

func TestFileNotifier_Send(t *testing.T) {
	testCases := []struct {
		name            string
		appendMode      bool
		setpathbase     bool
		preContent      string
		subject         string
		data            any
		expectContains  []string
		expectNotHas    []string
		expectErrSubstr string
	}{
		{
			name:           "ShouldTruncateExistingFile",
			appendMode:     false,
			preContent:     "OLD",
			subject:        "SubjectOne",
			data:           map[string]string{"User": "World"},
			expectContains: []string{"SubjectOne", "World"},
			expectNotHas:   []string{"OLD"},
		},
		{
			name:           "ShouldAppendToExistingFile",
			appendMode:     true,
			preContent:     "BASE\n",
			subject:        "SubjectTwo",
			data:           map[string]string{"User": "Alice"},
			expectContains: []string{"BASE\n", "SubjectTwo", "Alice"},
		},
		{
			name:            "ShouldReturnErrorWhenOpenFileFails",
			appendMode:      false,
			setpathbase:     true,
			preContent:      "",
			subject:         "X",
			data:            map[string]string{"User": "Y"},
			expectErrSubstr: "failed to open file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			filePath := filepath.Join(base, "notify.log")

			if tc.setpathbase {
				filePath = base
			}

			if tc.preContent != "" && tc.name != "ShouldReturnErrorWhenOpenFileFails" {
				require.NoError(t, os.WriteFile(filePath, []byte(tc.preContent), 0o600))
			}

			n := NewFileNotifier(schema.NotifierFileSystem{Filename: filePath})
			n.append = tc.appendMode

			tmpl := template.Must(template.New("text").Parse("Hello {{ .User }}"))
			et := &templates.EmailTemplate{Text: tmpl}

			rcpt := mail.Address{Name: "John Doe", Address: "john@example.com"}
			err := n.Send(context.Background(), rcpt, tc.subject, et, tc.data)

			if tc.expectErrSubstr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrSubstr)

				return
			}

			require.NoError(t, err)

			f, err := os.Open(filePath)
			require.NoError(t, err)

			defer f.Close()

			buf := bufio.NewScanner(f)

			var content string
			for buf.Scan() {
				content += buf.Text() + "\n"
			}

			require.NoError(t, buf.Err())

			for _, s := range tc.expectContains {
				assert.Contains(t, content, s)
			}

			for _, s := range tc.expectNotHas {
				assert.NotContains(t, content, s)
			}
		})
	}
}
