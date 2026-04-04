package templates

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSecretEnvKey(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		expected bool
	}{
		{"ShouldReturnFalseForKeysWithoutPrefix", []string{"A_KEY", "A_SECRET", "A_PASSWORD", "NOT_AUTHELIA_A_PASSWORD"}, false},
		{"ShouldReturnFalseForKeysWithoutSuffix", []string{"AUTHELIA_EXAMPLE", "X_AUTHELIA_EXAMPLE", "X_AUTHELIA_PASSWORD_NOT"}, false},
		{"ShouldReturnTrueForSecretKeys", []string{"AUTHELIA_JWT_SECRET", "AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET", "AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_CERTIFICATE_CHAIN", "X_AUTHELIA_JWT_SECRET", "X_AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET", "X_AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_CERTIFICATE_CHAIN"}, true},
		{"ShouldReturnTrueForSecretKeysEvenWithMixedCase", []string{"aUTHELIA_JWT_SECRET", "aUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET", "aUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_CERTIFICATE_CHAIN", "X_aUTHELIA_JWT_SECREt", "X_aUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET", "x_AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_CERTIFICATE_CHAIN"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, env := range tc.have {
				t.Run(env, func(t *testing.T) {
					assert.Equal(t, tc.expected, isSecretEnvKey(env))
				})
			}
		})
	}
}

func TestParseTemplateDirectories(t *testing.T) {
	testCases := []struct {
		name, path string
	}{
		{"Templates", "./embed"},
		{"OpenAPI", "../../api"},
		{"Generators", "../../cmd/authelia-gen/templates"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			funcMap := FuncMap()

			if tc.name == "Generators" {
				funcMap["joinX"] = FuncStringJoinX
			}

			var (
				data []byte
			)

			require.NoError(t, filepath.Walk(tc.path, func(path string, info fs.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}

				name := info.Name()

				if tc.name == "Templates" {
					name = filepath.Base(filepath.Dir(path)) + "/" + name
				}

				t.Run(name, func(t *testing.T) {
					data, err = os.ReadFile(path)

					require.NoError(t, err)

					_, err = template.New(tc.name).Funcs(funcMap).Parse(string(data))

					require.NoError(t, err)
				})

				return nil
			}))
		})
	}
}

func TestParseMiscTemplates(t *testing.T) {
	testCases := []struct {
		name, path string
	}{
		{"ReactIndex", "../../web/index.html"},
		{"ViteEnv", "../../web/.env.production"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.path)

			require.NoError(t, err)

			_, err = template.New(tc.name).Funcs(FuncMap()).Parse(string(data))

			require.NoError(t, err)
		})
	}
}

func TestLoadEmailTemplate(t *testing.T) {
	testCases := []struct {
		name  string
		tname string
		setup func(t *testing.T, tname, dir string)
		err   string
		errf  func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error)
	}{
		{
			"ShouldLoadEmailIdentityVerificationJWT",
			TemplateNameEmailIdentityVerificationJWT,
			nil,
			"",
			nil,
		},
		{
			"ShouldLoadEmailIdentityVerificationOTC",
			TemplateNameEmailIdentityVerificationOTC,
			nil,
			"",
			nil,
		},
		{
			"ShouldLoadEmailEvent",
			TemplateNameEmailEvent,
			nil,
			"",
			nil,
		},

		{
			"ShouldLoadEmailIdentityVerificationJWTWithLocal",
			TemplateNameEmailIdentityVerificationJWT,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0600))
			},
			"",
			nil,
		},
		{
			"ShouldLoadEmailIdentityVerificationOTCWithLocal",
			TemplateNameEmailIdentityVerificationOTC,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0600))
			},
			"",
			nil,
		},
		{
			"ShouldLoadEmailEventWithLocal",
			TemplateNameEmailEvent,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0600))
			},
			"",
			nil,
		},

		{
			"ShouldLoadEmailIdentityVerificationJWTWithLocalPermissionDeniedTxt",
			TemplateNameEmailIdentityVerificationJWT,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0000))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".txt")
				assert.EqualError(t, err, fmt.Sprintf("error occurred reading text template: failed to read template override at path '%s': open %s: permission denied", p, p))
			},
		},
		{
			"ShouldLoadEmailIdentityVerificationOTCWithLocalPermissionDeniedTxt",
			TemplateNameEmailIdentityVerificationOTC,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0000))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".txt")
				assert.EqualError(t, err, fmt.Sprintf("error occurred reading text template: failed to read template override at path '%s': open %s: permission denied", p, p))
			},
		},
		{
			"ShouldLoadEmailEventWithLocalPermissionDeniedTxt",
			TemplateNameEmailEvent,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0000))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".txt")
				assert.EqualError(t, err, fmt.Sprintf("error occurred reading text template: failed to read template override at path '%s': open %s: permission denied", p, p))
			},
		},
		{
			"ShouldLoadEmailIdentityVerificationJWTWithLocalPermissionDeniedHTML",
			TemplateNameEmailIdentityVerificationJWT,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0000))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0600))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".html")
				assert.EqualError(t, err, fmt.Sprintf("error occurred reading html template: failed to read template override at path '%s': open %s: permission denied", p, p))
			},
		},
		{
			"ShouldLoadEmailIdentityVerificationOTCWithLocalPermissionDeniedHTML",
			TemplateNameEmailIdentityVerificationOTC,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0000))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0600))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".html")
				assert.EqualError(t, err, fmt.Sprintf("error occurred reading html template: failed to read template override at path '%s': open %s: permission denied", p, p))
			},
		},
		{
			"ShouldLoadEmailEventWithLocalPermissionDeniedHTML",
			TemplateNameEmailEvent,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0000))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0600))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".html")
				assert.EqualError(t, err, fmt.Sprintf("error occurred reading html template: failed to read template override at path '%s': open %s: permission denied", p, p))
			},
		},

		{
			"ShouldLoadEmailIdentityVerificationJWTWithLocalBadTemplateTEXT",
			TemplateNameEmailIdentityVerificationJWT,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data {{ badtemplate"), 0600))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".txt")
				assert.EqualError(t, err, fmt.Sprintf("error occurred parsing text template: failed to parse template override at path '%s': template: IdentityVerificationJWT.txt:1: function \"badtemplate\" not defined", p))
			},
		},
		{
			"ShouldLoadEmailIdentityVerificationOTCWithLocalBadTemplateTEXT",
			TemplateNameEmailIdentityVerificationOTC,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data {{ badtemplate"), 0600))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".txt")
				assert.EqualError(t, err, fmt.Sprintf("error occurred parsing text template: failed to parse template override at path '%s': template: IdentityVerificationOTC.txt:1: function \"badtemplate\" not defined", p))
			},
		},
		{
			"ShouldLoadEmailEventWithLocalBadTemplateTEXT",
			TemplateNameEmailEvent,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data {{ badtemplate"), 0600))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".txt")
				assert.EqualError(t, err, fmt.Sprintf("error occurred parsing text template: failed to parse template override at path '%s': template: Event.txt:1: function \"badtemplate\" not defined", p))
			},
		},

		{
			"ShouldLoadEmailIdentityVerificationJWTWithLocalBadTemplateHTML",
			TemplateNameEmailIdentityVerificationJWT,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data {{ badtemplate"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0600))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".html")
				assert.EqualError(t, err, fmt.Sprintf("error occurred parsing html template: failed to parse template override at path '%s': template: IdentityVerificationJWT.html:1: function \"badtemplate\" not defined", p))
			},
		},
		{
			"ShouldLoadEmailIdentityVerificationOTCWithLocalBadTemplateHTML",
			TemplateNameEmailIdentityVerificationOTC,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data {{ badtemplate"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0600))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".html")
				assert.EqualError(t, err, fmt.Sprintf("error occurred parsing html template: failed to parse template override at path '%s': template: IdentityVerificationOTC.html:1: function \"badtemplate\" not defined", p))
			},
		},
		{
			"ShouldLoadEmailEventWithLocalBadTemplateHTML",
			TemplateNameEmailEvent,
			func(t *testing.T, tname, dir string) {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".html"), []byte("data {{ badtemplate"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, tname+".txt"), []byte("data"), 0600))
			},
			"",
			func(t *testing.T, tname, dir string, tmpl *EmailTemplate, err error) {
				p := filepath.Join(dir, tname+".html")
				assert.EqualError(t, err, fmt.Sprintf("error occurred parsing html template: failed to parse template override at path '%s': template: Event.html:1: function \"badtemplate\" not defined", p))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			if tc.setup != nil {
				tc.setup(t, tc.tname, dir)
			}

			tmpl, err := loadEmailTemplate(tc.tname, dir)

			switch {
			case tc.err != "":
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, tmpl)
			case tc.errf != nil:
				tc.errf(t, tc.tname, dir, tmpl, err)
			default:
				require.NoError(t, err)
				assert.NotNil(t, tmpl)
			}
		})
	}
}

func TestStrval(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected string
	}{
		{"ShouldReturnString", "hello", "hello"},
		{"ShouldReturnByteSlice", []byte("world"), "world"},
		{"ShouldReturnStringer", &url.URL{Scheme: "https", Host: "example.com"}, "https://example.com"},
		{"ShouldReturnFmtFallbackForInt", 42, "42"},
		{"ShouldReturnFmtFallbackForBool", true, "true"},
		{"ShouldReturnFmtFallbackForNil", nil, "<nil>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, strval(tc.input))
		})
	}
}

func TestStrslice(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected []string
	}{
		{"ShouldReturnStringSlice", []string{"a", "b"}, []string{"a", "b"}},
		{"ShouldConvertAnySlice", []any{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"ShouldSkipNilsInAnySlice", []any{"a", nil, "c"}, []string{"a", "c"}},
		{"ShouldConvertEmptyAnySlice", []any{}, []string{}},
		{"ShouldConvertIntSliceViaReflect", []int{1, 2, 3}, []string{"1", "2", "3"}},
		{"ShouldReturnEmptyForNil", nil, []string{}},
		{"ShouldWrapSingleValue", "single", []string{"single"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, strslice(tc.input))
		})
	}
}

func TestConvertAnyToTime(t *testing.T) {
	now := time.Now()
	nowPtr := &now

	testCases := []struct {
		name     string
		input    any
		expected time.Time
		approx   bool
	}{
		{"ShouldReturnTimeValue", now, now, false},
		{"ShouldReturnTimePointer", nowPtr, now, false},
		{"ShouldReturnFromInt64", int64(1000000), time.Unix(1000000, 0), false},
		{"ShouldReturnFromInt", 500, time.Unix(500, 0), false},
		{"ShouldReturnFromInt32", int32(250), time.Unix(250, 0), false},
		{"ShouldReturnNowForUnknownType", "not-a-time", time.Time{}, true},
		{"ShouldReturnNowForBool", false, time.Time{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertAnyToTime(tc.input)

			if tc.approx {
				assert.InDelta(t, time.Now().Unix(), result.Unix(), 2)
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
