package templates

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"text/template"

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
