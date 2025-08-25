package templates

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	provider, err := New(Config{})
	require.NoError(t, err)
	assert.NotNil(t, provider)

	dir := t.TempDir()

	provider, err = New(Config{EmailTemplatesPath: dir})
	require.NoError(t, err)
	assert.NotNil(t, provider)

	assert.Nil(t, provider.GetAssetOpenAPIIndexTemplate())
	assert.Nil(t, provider.GetAssetOpenAPISpecTemplate())
	assert.Nil(t, provider.GetAssetIndexTemplate())
	assert.NotNil(t, provider.GetEventEmailTemplate())
	assert.NotNil(t, provider.GetIdentityVerificationJWTEmailTemplate())
	assert.NotNil(t, provider.GetIdentityVerificationOTCEmailTemplate())
	assert.NotNil(t, provider.GetOpenIDConnectAuthorizeResponseFormPostTemplate())
}

func TestLoadTemplatedAssets(t *testing.T) {
	testCases := []struct {
		name        string
		badtemplate string
		missing     bool
		err         string
	}{
		{
			"ShouldPass",
			"",
			false,
			"",
		},
		{
			"ShouldFailIndex",
			"index.html",
			false,
			"error occurred loading template 'assets/public_html/index.html': template: assets/public_html/index.html:1: function \"bad\" not defined",
		},
		{
			"ShouldFailAPIIndex",
			"api/index.html",
			false,
			"error occurred loading template 'assets/public_html/api/index.html': template: assets/public_html/api/index.html:1: function \"bad\" not defined",
		},
		{
			"ShouldFailAPISpec",
			"api/openapi.yml",
			false,
			"error occurred loading template 'assets/public_html/api/openapi.yml': template: assets/public_html/api/openapi.yml:1: function \"bad\" not defined",
		},
		{
			"ShouldFailIndexMissing",
			"index.html",
			true,
			"error occurred loading template 'assets/public_html/index.html': open public_html/index.html: no such file or directory",
		},
		{
			"ShouldFailAPIIndexMissing",
			"api/index.html",
			true,
			"error occurred loading template 'assets/public_html/api/index.html': open public_html/api/index.html: no such file or directory",
		},
		{
			"ShouldFailAPISpecMissing",
			"api/openapi.yml",
			true,
			"error occurred loading template 'assets/public_html/api/openapi.yml': open public_html/api/openapi.yml: no such file or directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			assert.NoError(t, os.MkdirAll(filepath.Join(dir, "public_html"), 0700))
			assert.NoError(t, os.MkdirAll(filepath.Join(dir, "public_html", "api"), 0700))

			if tc.badtemplate == "index.html" {
				if !tc.missing {
					assert.NoError(t, os.WriteFile(filepath.Join(dir, "public_html", "index.html"), []byte("not html {{ bad template"), 0600))
				}
			} else {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, "public_html", "index.html"), []byte("not html"), 0600))
			}

			if tc.badtemplate == "api/index.html" {
				if !tc.missing {
					assert.NoError(t, os.WriteFile(filepath.Join(dir, "public_html", "api", "index.html"), []byte("not html {{ bad template"), 0600))
				}
			} else {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, "public_html", "api", "index.html"), []byte("not html"), 0600))
			}

			if tc.badtemplate == "api/openapi.yml" {
				if !tc.missing {
					assert.NoError(t, os.WriteFile(filepath.Join(dir, "public_html", "api", "openapi.yml"), []byte("not yml {{ bad template"), 0600))
				}
			} else if !tc.missing {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, "public_html", "api", "openapi.yml"), []byte("not yml"), 0600))
			}

			provider, err := New(Config{EmailTemplatesPath: dir})
			require.NoError(t, err)
			require.NotNil(t, provider)

			err = provider.LoadTemplatedAssets(&testrffs{FS: os.DirFS(dir)})

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

type testrffs struct {
	fs.FS
}

func (t *testrffs) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(t.FS, name)
}
