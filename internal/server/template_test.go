package server

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

const (
	assetsOpenAPIPath = "public_html/api/openapi.yml"
	localOpenAPIPath  = "../../api/openapi.yml"
)

type ReadFileOpenAPI struct{}

func (lfs *ReadFileOpenAPI) Open(name string) (fs.File, error) {
	switch name {
	case assetsOpenAPIPath:
		return os.Open(localOpenAPIPath)
	default:
		return assets.Open(name)
	}
}

func (lfs *ReadFileOpenAPI) ReadFile(name string) ([]byte, error) {
	switch name {
	case assetsOpenAPIPath:
		return os.ReadFile(localOpenAPIPath)
	default:
		return assets.ReadFile(name)
	}
}

func TestShouldTemplateOpenAPI(t *testing.T) {
	provider, err := templates.New(templates.Config{})
	require.NoError(t, err)

	fs := &ReadFileOpenAPI{}

	require.NoError(t, provider.LoadTemplatedAssets(fs))

	mock := mocks.NewMockAutheliaCtx(t)

	mock.Ctx.Configuration.Server = schema.DefaultServerConfiguration
	mock.Ctx.Configuration.Session = schema.Session{
		Cookies: []schema.SessionCookie{
			{
				Domain: "example.com",
			},
		},
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	opts := NewTemplatedFileOptions(&mock.Ctx.Configuration)

	handler := ServeTemplatedOpenAPI(provider.GetAssetOpenAPISpecTemplate(), opts)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-URI", "/api/openapi.yml")

	handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	body := string(mock.Ctx.Response.Body())

	assert.NotEqual(t, "", body)
	assert.Contains(t, body, "example: 'https://auth.example.com/?rd=https%3A%2F%2Fexample.com%2F&rm=GET'")
}

func TestServeTemplatedFile(t *testing.T) {
	tmpl, err := templates.New(templates.Config{})
	require.NoError(t, err)

	require.NoError(t, tmpl.LoadTemplatedAssets(assets))

	testCases := []struct {
		name               string
		method             string
		language           string
		cspTemplate        string
		expectedStatusCode int
		expectBody         bool
		expectCSP          bool
	}{
		{
			"ShouldServeIndexWithDefaultLanguage",
			fasthttp.MethodGet,
			"",
			"",
			fasthttp.StatusOK,
			true,
			true,
		},
		{
			"ShouldServeIndexWithCustomLanguage",
			fasthttp.MethodGet,
			"fr",
			"",
			fasthttp.StatusOK,
			true,
			true,
		},
		{
			"ShouldServeIndexWithInvalidLanguageFallback",
			fasthttp.MethodGet,
			"<script>alert(1)</script>",
			"",
			fasthttp.StatusOK,
			true,
			true,
		},
		{
			"ShouldHandleHEADRequest",
			fasthttp.MethodHead,
			"",
			"",
			fasthttp.StatusOK,
			false,
			true,
		},
		{
			"ShouldUseCustomCSPTemplate",
			fasthttp.MethodGet,
			"",
			"default-src 'self'; script-src 'nonce-${NONCE}'",
			fasthttp.StatusOK,
			true,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Configuration.Server = schema.DefaultServerConfiguration
			mock.Ctx.Configuration.Server.Headers.CSPTemplate = schema.CSPTemplate(tc.cspTemplate)
			mock.Ctx.Configuration.Session = schema.Session{
				Cookies: []schema.SessionCookie{
					{Domain: "example.com"},
				},
			}

			mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

			opts := NewTemplatedFileOptions(&mock.Ctx.Configuration)

			handler := ServeTemplatedFile(tmpl.GetAssetIndexTemplate(), opts)

			mock.Ctx.Request.Header.SetMethod(tc.method)
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com")

			if tc.language != "" {
				mock.Ctx.Request.Header.SetCookie("language", tc.language)
			}

			handler(mock.Ctx)

			assert.Equal(t, tc.expectedStatusCode, mock.Ctx.Response.StatusCode())

			if !tc.expectBody {
				assert.True(t, mock.Ctx.Response.SkipBody)
			}

			if tc.expectCSP {
				csp := string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderContentSecurityPolicy))
				assert.NotEmpty(t, csp)
			}
		})
	}
}

func TestServeTemplatedFileWithCSSOverride(t *testing.T) {
	tmpl, err := templates.New(templates.Config{})
	require.NoError(t, err)

	require.NoError(t, tmpl.LoadTemplatedAssets(assets))

	testCases := []struct {
		name     string
		withCSS  bool
		expected string
	}{
		{
			"ShouldIncludeCSSLink",
			true,
			"\"CSSOverride\":\"true\"",
		},
		{
			"ShouldNotIncludeCSSLink",
			false,
			"\"CSSOverride\":\"true\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			if tc.withCSS {
				require.NoError(t, os.WriteFile(filepath.Join(dir, fileCustomCSS), []byte("body { background: red; }"), 0644))
			}

			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Configuration.Server = schema.DefaultServerConfiguration
			mock.Ctx.Configuration.Server.AssetPath = dir
			mock.Ctx.Configuration.Session = schema.Session{
				Cookies: []schema.SessionCookie{
					{Domain: "example.com"},
				},
			}

			mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

			opts := NewTemplatedFileOptions(&mock.Ctx.Configuration)

			handler := ServeTemplatedFile(tmpl.GetAssetIndexTemplate(), opts)

			mock.Ctx.Request.Header.SetMethod(fasthttp.MethodGet)
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com")

			handler(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

			body := string(mock.Ctx.Response.Body())

			if tc.withCSS {
				assert.Contains(t, body, tc.expected)
			} else {
				assert.NotContains(t, body, tc.expected)
			}
		})
	}
}

func TestWebIndexTemplateCSSOverride(t *testing.T) {
	importTmpl := func(data string) *template.Template {
		return template.Must(template.New("test").Funcs(templates.FuncMap()).Parse(data))
	}

	data, err := os.ReadFile("../../web/index.html")
	require.NoError(t, err)

	t.Run("ShouldIncludeCSSLink", func(t *testing.T) {
		t.Run("Enabled", func(t *testing.T) {
			tt := importTmpl(string(data))

			opts := &TemplatedFileCommonData{
				CSSOverride: "true",
			}

			buf := &bytes.Buffer{}
			require.NoError(t, tt.Execute(buf, opts))

			assert.Contains(t, buf.String(), "<link rel=\"stylesheet\" href=\"/static/media/custom.css\">")
		})

		t.Run("Disabled", func(t *testing.T) {
			tt := importTmpl(string(data))

			opts := &TemplatedFileCommonData{
				CSSOverride: "false",
			}

			buf := &bytes.Buffer{}
			require.NoError(t, tt.Execute(buf, opts))

			assert.NotContains(t, buf.String(), "<link rel=\"stylesheet\" href=\"/static/media/custom.css\">")
		})
	})
}

func TestETagRootURL(t *testing.T) {
	tmpl, err := templates.New(templates.Config{})
	require.NoError(t, err)

	lfs := &ReadFileOpenAPI{}

	require.NoError(t, tmpl.LoadTemplatedAssets(lfs))

	testCases := []struct {
		name               string
		sendETag           bool
		expectedStatusCode int
	}{
		{
			"ShouldReturn200OnFirstRequest",
			false,
			fasthttp.StatusOK,
		},
		{
			"ShouldReturn304WithMatchingETag",
			true,
			fasthttp.StatusNotModified,
		},
	}

	opts := NewTemplatedFileOptions(&schema.Configuration{
		Server: schema.DefaultServerConfiguration,
	})

	innerHandler := ServeTemplatedOpenAPI(tmpl.GetAssetOpenAPISpecTemplate(), opts)
	handler := ETagRootURL(innerHandler)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Configuration.Server = schema.DefaultServerConfiguration
			mock.Ctx.Configuration.Session = schema.Session{
				Cookies: []schema.SessionCookie{
					{Domain: "example.com"},
				},
			}

			mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com")
			mock.Ctx.Request.Header.Set("X-Forwarded-URI", "/api/openapi.yml")

			if tc.sendETag {
				firstMock := mocks.NewMockAutheliaCtx(t)
				defer firstMock.Close()

				firstMock.Ctx.Configuration.Server = schema.DefaultServerConfiguration
				firstMock.Ctx.Configuration.Session = mock.Ctx.Configuration.Session
				firstMock.Ctx.Providers.SessionProvider = session.NewProvider(firstMock.Ctx.Configuration.Session, nil)
				firstMock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
				firstMock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com")
				firstMock.Ctx.Request.Header.Set("X-Forwarded-URI", "/api/openapi.yml")

				handler(firstMock.Ctx)

				etag := firstMock.Ctx.Response.Header.Peek("ETag")

				require.NotEmpty(t, etag)

				mock.Ctx.Request.Header.SetBytesKV([]byte("If-None-Match"), etag)
			}

			handler(mock.Ctx)

			assert.Equal(t, tc.expectedStatusCode, mock.Ctx.Response.StatusCode())

			etag := mock.Ctx.Response.Header.Peek("ETag")
			assert.NotEmpty(t, etag)
		})
	}
}

func TestNewTemplatedFileOptions(t *testing.T) {
	testCases := []struct {
		name                   string
		config                 *schema.Configuration
		expectedResetPassword  string
		expectedPasswordChange string
		expectedTheme          string
		expectedPasskeyLogin   string
	}{
		{
			"ShouldReturnDefaultOptions",
			&schema.Configuration{},
			"true",
			"true",
			"",
			"false",
		},
		{
			"ShouldDisableResetPassword",
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					PasswordReset: schema.AuthenticationBackendPasswordReset{
						Disable: true,
					},
				},
			},
			"false",
			"true",
			"",
			"false",
		},
		{
			"ShouldEnablePasskeyLogin",
			&schema.Configuration{
				WebAuthn: schema.WebAuthn{
					EnablePasskeyLogin: true,
				},
			},
			"true",
			"true",
			"",
			"true",
		},
		{
			"ShouldSetTheme",
			&schema.Configuration{
				Theme: "dark",
			},
			"true",
			"true",
			"dark",
			"false",
		},
		{
			"ShouldDisablePasswordChange",
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					PasswordChange: schema.AuthenticationBackendPasswordChange{
						Disable: true,
					},
				},
			},
			"true",
			"false",
			"",
			"false",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewTemplatedFileOptions(tc.config)

			assert.NotNil(t, opts)
			assert.Equal(t, tc.expectedResetPassword, opts.ResetPassword)
			assert.Equal(t, tc.expectedPasswordChange, opts.PasswordChange)
			assert.Equal(t, tc.expectedTheme, opts.Theme)
			assert.Equal(t, tc.expectedPasskeyLogin, opts.PasskeyLogin)
		})
	}
}

func TestTemplatedFileOptionsCommonData(t *testing.T) {
	testCases := []struct {
		name       string
		rememberMe string
		expectedRM string
	}{
		{
			"ShouldReturnDefaultRememberMe",
			"",
			"true",
		},
		{
			"ShouldOverrideRememberMe",
			"false",
			"false",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewTemplatedFileOptions(&schema.Configuration{})

			data := opts.CommonData("/", "/", "example.com", "nonce123", "en", "", "", tc.rememberMe)


			assert.Equal(t, "/", data.Base)
			assert.Equal(t, "example.com", data.Domain)
			assert.Equal(t, "nonce123", data.CSPNonce)
			assert.Equal(t, "en", data.Language)
			assert.Equal(t, tc.expectedRM, data.RememberMe)
		})
	}
}

func TestTemplatedFileOptionsOpenAPIData(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"ShouldReturnOpenAPIData"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewTemplatedFileOptions(&schema.Configuration{})

			data := opts.OpenAPIData("/", "/api", "example.com", "nonce123")

			assert.Equal(t, "/", data.Base)
			assert.Equal(t, "/api", data.BaseURL)
			assert.Equal(t, "example.com", data.Domain)
			assert.Equal(t, "nonce123", data.CSPNonce)
		})
	}
}
