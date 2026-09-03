package server

import (
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
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

const tmplTestPortalIndex = `<!doctype html>
<html lang="{{ .Language }}">
    <head>
        <meta property="csp-nonce" content="{{ .CSPNonce }}" />
    </head>
    <body data-theme="{{ .Theme }}" data-rememberme="{{ .RememberMe }}">
        <div id="root"></div>
    </body>
</html>`

var reTestCSPNonce = regexp.MustCompile(`'nonce-([a-zA-Z0-9]{32})'`)

// ReadFilePortalIndex substitutes the portal index template so the per request values can be asserted without having
// built the frontend.
type ReadFilePortalIndex struct{}

func (lfs *ReadFilePortalIndex) Open(name string) (fs.File, error) {
	return assets.Open(name)
}

func (lfs *ReadFilePortalIndex) ReadFile(name string) ([]byte, error) {
	switch name {
	case "public_html/index.html":
		return []byte(tmplTestPortalIndex), nil
	default:
		return assets.ReadFile(name)
	}
}

func TestServeTemplatedFileShouldServeIdentityWithPerRequestNonce(t *testing.T) {
	tmpl, err := templates.New(templates.Config{})
	require.NoError(t, err)

	require.NoError(t, tmpl.LoadTemplatedAssets(&ReadFilePortalIndex{}))

	config := &schema.Configuration{Server: schema.DefaultServerConfiguration, Theme: "grey"}

	handler := ServeTemplatedFile(tmpl.GetAssetIndexTemplate(), NewTemplatedFileOptions(config))

	nonces := make([]string, 0, 2)

	for range 2 {
		mock := mocks.NewMockAutheliaCtx(t)

		mock.Ctx.Configuration.Server = schema.DefaultServerConfiguration

		mock.Ctx.Request.Header.Set(fasthttp.HeaderAcceptEncoding, "br, gzip, deflate")
		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com")

		handler(mock.Ctx)

		require.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		// The index is templated per request so it can't be pre-compressed, and it's small enough that compressing it
		// on the fly isn't worthwhile.
		assert.Empty(t, mock.Ctx.Response.Header.Peek(fasthttp.HeaderContentEncoding))
		assert.Contains(t, string(mock.Ctx.Response.Header.ContentType()), "text/html")

		body := string(mock.Ctx.Response.Body())

		assert.Contains(t, body, `data-theme="grey"`)

		matches := reTestCSPNonce.FindStringSubmatch(string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderContentSecurityPolicy)))
		require.Len(t, matches, 2)

		assert.Contains(t, body, `<meta property="csp-nonce" content="`+matches[1]+`" />`)

		nonces = append(nonces, matches[1])

		mock.Close()
	}

	assert.NotEqual(t, nonces[0], nonces[1])
}

// The direct handler test above can't show that the route which actually serves the portal is the templated one, so
// this drives the registered "/" route to prove the index reaching a client is neither compressed nor shared between
// requests.
func TestHandlerMainShouldServeTemplatedIndexUncompressed(t *testing.T) {
	provider, err := templates.New(templates.Config{})
	require.NoError(t, err)

	require.NoError(t, provider.LoadTemplatedAssets(&ReadFilePortalIndex{}))

	providers := middlewares.NewProvidersBasic()
	providers.Templates = provider

	config := &schema.Configuration{
		Server: schema.Server{
			Address:   schema.DefaultServerConfiguration.Address,
			Endpoints: schema.DefaultServerConfiguration.Endpoints,
		},
		Theme: "grey",
	}

	handler, err := handlerMain(config, providers)

	require.NoError(t, err)

	testCases := []struct {
		name           string
		method         string
		acceptEncoding string
	}{
		{"ShouldServeIdentityToBrotliGET", fasthttp.MethodGet, "br"},
		{"ShouldServeIdentityToGzipGET", fasthttp.MethodGet, "gzip"},
		{"ShouldServeIdentityToBrotliHEAD", fasthttp.MethodHead, "br"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nonces := make([]string, 0, 2)

			for range 2 {
				ctx := newAssetRequestCtx(tc.method, "/", tc.acceptEncoding)

				ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
				ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com")

				handler(ctx)

				require.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())

				assert.Empty(t, ctx.Response.Header.Peek(fasthttp.HeaderContentEncoding))
				assert.Contains(t, string(ctx.Response.Header.ContentType()), "text/html")

				matches := reTestCSPNonce.FindStringSubmatch(string(ctx.Response.Header.Peek(fasthttp.HeaderContentSecurityPolicy)))
				require.Len(t, matches, 2)

				if tc.method == fasthttp.MethodGet {
					body := string(ctx.Response.Body())

					assert.Contains(t, body, `data-theme="grey"`)
					assert.Contains(t, body, `<meta property="csp-nonce" content="`+matches[1]+`" />`)
				} else {
					assert.True(t, ctx.Response.SkipBody)
					assert.Empty(t, ctx.Response.Body())

					contentLength, err := strconv.Atoi(string(ctx.Response.Header.Peek(fasthttp.HeaderContentLength)))

					require.NoError(t, err)
					assert.Positive(t, contentLength)
				}

				nonces = append(nonces, matches[1])
			}

			assert.NotEqual(t, nonces[0], nonces[1])
		})
	}
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
			assert.Regexp(t, `^"[0-9a-f]{40}"$`, string(etag), "etag should be a quoted opaque string")
		})
	}
}

func TestNewTemplatedFileOptions(t *testing.T) {
	testCases := []struct {
		name                       string
		config                     *schema.Configuration
		expectedResetPassword      string
		expectedPasswordChange     string
		expectedTheme              string
		expectedPasskeyLogin       string
		expectedOpenIDConnectLogin string
	}{
		{
			"ShouldReturnDefaultOptions",
			&schema.Configuration{},
			"true",
			"true",
			"",
			"false",
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
			"false",
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
			"false",
		},
		{
			"ShouldEnableOpenIDConnectLogin",
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					OpenIDConnect: &schema.AuthenticationBackendOpenIDConnect{
						Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
							{ID: "example", Name: "Example", Issuer: "https://example.com"},
						},
					},
				},
			},
			"true",
			"true",
			"",
			"false",
			"true",
		},
		{
			"ShouldNotEnableOpenIDConnectLoginWithoutProviders",
			&schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					OpenIDConnect: &schema.AuthenticationBackendOpenIDConnect{},
				},
			},
			"true",
			"true",
			"",
			"false",
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
			assert.Equal(t, tc.expectedOpenIDConnectLogin, opts.OpenIDConnectLogin)
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

			data := opts.CommonData("/", "/", "example.com", "nonce123", "en", "", tc.rememberMe)

			assert.Equal(t, "/", data.Base)
			assert.Equal(t, "example.com", data.Domain)
			assert.Equal(t, "nonce123", data.CSPNonce)
			assert.Equal(t, "en", data.Language)
			assert.Equal(t, tc.expectedRM, data.RememberMe)
			assert.Equal(t, opts.OpenIDConnectLogin, data.OpenIDConnectLogin)
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
