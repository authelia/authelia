package utils

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestGetDirectoryLanguages(t *testing.T) {
	haveA := t.TempDir()
	haveB := t.TempDir()

	mustMkdirAll(t, filepath.Join(haveB, "en"))
	mustMkdirAll(t, filepath.Join(haveB, "fr"))
	mustMkdirAll(t, filepath.Join(haveB, "en-US"))
	mustMkdirAll(t, filepath.Join(haveB, "es"))
	mustMkdirAll(t, filepath.Join(haveB, "es-AR"))

	mustWriteFile(t, filepath.Join(haveB, "en", "portal.json"), []byte("package a"))
	mustWriteFile(t, filepath.Join(haveB, "en", "settings.json"), []byte("package a"))
	mustWriteFile(t, filepath.Join(haveB, "fr", "portal.json"), []byte("package a"))
	mustWriteFile(t, filepath.Join(haveB, "fr", "settings.json"), []byte("package a"))
	mustWriteFile(t, filepath.Join(haveB, "en-US", "portal.json"), []byte("package a"))
	mustWriteFile(t, filepath.Join(haveB, "en-US", "settings.json"), []byte("package a"))

	mustWriteFile(t, filepath.Join(haveB, "es", "portal.json"), []byte("package a"))
	mustWriteFile(t, filepath.Join(haveB, "es", "settings.json"), []byte("package a"))
	mustWriteFile(t, filepath.Join(haveB, "es-AR", "portal.json"), []byte("package a"))
	mustWriteFile(t, filepath.Join(haveB, "es-AR", "settings.json"), []byte("package a"))

	testCases := []struct {
		name     string
		have     string
		expected *Languages
		err      string
	}{
		{
			"ShouldErrorEmptyDir",
			"",
			nil,
			"stat .: os: DirFS with empty root",
		},
		{
			"ShouldNotErrorHaveA",
			haveA,
			&Languages{
				Defaults: DefaultsLanguages{
					Language: Language{
						Display: "English",
						Locale:  "en",
					},
					Namespace: "portal",
				},
			},
			"",
		},
		{
			"ShouldNotErrorHaveB",
			haveB,
			&Languages{
				Defaults: DefaultsLanguages{
					Language: Language{
						Display: "English",
						Locale:  "en",
					},
					Namespace: "portal",
				},
				Namespaces: []string{"portal", "settings"},
				Languages: []Language{
					{
						Display:    "English",
						Locale:     "en",
						Namespaces: []string{"portal", "settings"},
						Fallbacks:  []string{"en"},
						Tag:        language.MustParse("en"),
					},
					{
						Display:    "American English",
						Locale:     "en-US",
						Namespaces: []string{"portal", "settings"},
						Fallbacks:  []string{"en", "en"},
						Parent:     "en",
						Tag:        language.MustParse("en-US"),
					},
					{
						Display:    "Español",
						Locale:     "es",
						Namespaces: []string{"portal", "settings"},
						Fallbacks:  []string{"en"},
						Tag:        language.MustParse("es"),
					},
					{
						Display:    "Español",
						Locale:     "es-AR",
						Parent:     "es",
						Namespaces: []string{"portal", "settings"},
						Fallbacks:  []string{"es", "en"},
						Tag:        language.MustParse("es-AR"),
					},
					{
						Display:    "Français",
						Locale:     "fr",
						Namespaces: []string{"portal", "settings"},
						Fallbacks:  []string{"en"},
						Tag:        language.MustParse("fr"),
					},
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			langs, err := GetDirectoryLanguages(tc.have)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, langs)
			} else {
				require.NoError(t, err)
				require.NotNil(t, langs)

				assert.ElementsMatch(t, tc.expected.Namespaces, langs.Namespaces)
				assert.ElementsMatch(t, tc.expected.Languages, langs.Languages)
				assert.Equal(t, tc.expected.Defaults, langs.Defaults)
			}
		})
	}
}

func TestGetEmbeddedLanguages(t *testing.T) {
	testCases := []struct {
		name     string
		have     embed.FS
		expected *Languages
		err      string
	}{
		{},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			langs, err := GetEmbeddedLanguages(tc.have)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, langs)
			} else {
				require.NoError(t, err)
				require.NotNil(t, langs)
			}
		})
	}
}

func TestGetLocaleParentOrBaseString(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
		err      string
	}{
		{
			"ShouldHandleEnglish",
			"en",
			"en",
			"",
		},
		{
			"ShouldHandleMalformed",
			"zzzz",
			"",
			"failed to parse language 'zzzz': language: tag is not well-formed",
		},
		{
			"ShouldHandleUnknown",
			"zz",
			"",
			"failed to parse language 'zz': language: subtag \"zz\" is well-formed but unknown",
		},
		{
			"ShouldHandleSub",
			"es-AR",
			"es",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lang, err := GetLocaleParentOrBaseString(tc.have)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Equal(t, tc.expected, lang)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, lang)
			}
		})
	}
}

func mustMkdirAll(t *testing.T, p string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(p, 0o700))
}

func mustWriteFile(t *testing.T, p string, data []byte) {
	t.Helper()
	require.NoError(t, os.WriteFile(p, data, 0o600))
}
