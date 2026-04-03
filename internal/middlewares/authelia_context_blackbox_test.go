package middlewares_test

import (
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestNewRequestLogger(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	ctx.SetUserValue(middlewares.UserValueKeyRawURI, "abc")

	_ = middlewares.NewRequestLogger(ctx)
}

func TestAutheliaCtx_RemoteIP(t *testing.T) {
	testCases := []struct {
		name     string
		have     []byte
		expected net.IP
	}{
		{"ShouldDefaultToRemoteAddr", nil, net.ParseIP("127.0.0.127")},
		{"ShouldParseProperlyFormattedXFFWithIPv4", []byte("192.168.1.1, 127.0.0.1"), net.ParseIP("192.168.1.1")},
		{"ShouldParseProperlyFormattedXFFWithIPv6", []byte("2001:db8:85a3:8d3:1319:8a2e:370:7348, 127.0.0.1"), net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348")},
		{"ShouldFallbackToRemoteAddrOnImproperlyFormattedXFFWithIPv6", []byte("[2001:db8:85a3:8d3:1319:8a2e:370:7348], 127.0.0.1"), net.ParseIP("127.0.0.127")},
		{"ShouldFallbackToRemoteAddrOnBlankXFFHeader", []byte(""), net.ParseIP("127.0.0.127")},
		{"ShouldFallbackToRemoteAddrOnBlankXFFEntry", []byte(", 127.0.0.1"), net.ParseIP("127.0.0.127")},
		{"ShouldFallbackToRemoteAddrOnBadXFFEntry", []byte("abc, 127.0.0.1"), net.ParseIP("127.0.0.127")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.SetRemoteAddr(&net.TCPAddr{Port: 80, IP: net.ParseIP("127.0.0.127")})

			if tc.have != nil {
				mock.Ctx.Request.Header.SetBytesV(fasthttp.HeaderXForwardedFor, tc.have)
			}

			assert.Equal(t, tc.expected, mock.Ctx.RemoteIP())
		})
	}
}

func TestContentTypes(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(ctx *middlewares.AutheliaCtx) (err error)
		expected string
	}{
		{
			name: "ApplicationJSON",
			setup: func(ctx *middlewares.AutheliaCtx) (err error) {
				ctx.SetContentTypeApplicationJSON()

				return nil
			},
			expected: "application/json; charset=utf-8",
		},
		{
			name: "ApplicationYAML",
			setup: func(ctx *middlewares.AutheliaCtx) (err error) {
				ctx.SetContentTypeApplicationYAML()

				return nil
			},
			expected: "application/yaml; charset=utf-8",
		},
		{
			name: "TextPlain",
			setup: func(ctx *middlewares.AutheliaCtx) (err error) {
				ctx.SetContentTypeTextPlain()

				return nil
			},
			expected: "text/plain; charset=utf-8",
		},
		{
			name: "TextHTML",
			setup: func(ctx *middlewares.AutheliaCtx) (err error) {
				ctx.SetContentTypeTextHTML()

				return nil
			},
			expected: "text/html; charset=utf-8",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			assert.NoError(t, tc.setup(mock.Ctx))

			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Header.ContentType()))
		})
	}
}

func TestAutheliaCtx_RootURL(t *testing.T) {
	testCases := []struct {
		name              string
		proto, host, base string
		expected          string
	}{
		{
			name:  "Standard",
			proto: "https", host: "auth.example.com", base: "",
			expected: "https://auth.example.com",
		},
		{
			name:  "Base",
			proto: "https", host: "example.com", base: "auth",
			expected: "https://example.com/auth",
		},
		{
			name:  "NoHost",
			proto: "https", host: "", base: "",
			expected: "https:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, tc.proto)
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tc.host)

			if tc.base != "" {
				mock.Ctx.SetUserValue(middlewares.UserValueKeyBaseURL, tc.base)
			}

			actual := mock.Ctx.RootURL()

			require.NotNil(t, actual)

			assert.Equal(t, tc.expected, actual.String())
			assert.Equal(t, tc.proto, actual.Scheme)
			assert.Equal(t, tc.host, actual.Host)
			assert.Equal(t, tc.base, actual.Path)
		})
	}
}

func TestAutheliaCtx_AuthzPath(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		expected []byte
	}{
		{
			"ShouldReturnValue",
			"exy",
			[]byte("exy"),
		},
		{
			"ShouldReturnValue",
			nil,
			[]byte(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.SetUserValue(middlewares.UserValueRouterKeyExtAuthzPath, tc.have)

			assert.Equal(t, tc.expected, mock.Ctx.AuthzPath())
		})
	}
}

func TestAutheliaCtx_RootURLSlash(t *testing.T) {
	testCases := []struct {
		name              string
		proto, host, base string
		expected          string
		expectedPath      string
	}{
		{
			name:  "Standard",
			proto: "https", host: "auth.example.com", base: "",
			expected:     "https://auth.example.com/",
			expectedPath: "/",
		},
		{
			name:  "StandardWithSlash",
			proto: "https", host: "auth.example.com", base: "/",
			expected:     "https://auth.example.com/",
			expectedPath: "/",
		},
		{
			name:  "Base",
			proto: "https", host: "example.com", base: "auth",
			expected:     "https://example.com/auth/",
			expectedPath: "auth/",
		},
		{
			name:  "NoHost",
			proto: "https", host: "", base: "",
			expected:     "https:///",
			expectedPath: "/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, tc.proto)
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, tc.host)

			if tc.base != "" {
				mock.Ctx.SetUserValue(middlewares.UserValueKeyBaseURL, tc.base)
			}

			actual := mock.Ctx.RootURLSlash()

			require.NotNil(t, actual)

			assert.Equal(t, tc.expected, actual.String())
			assert.Equal(t, tc.proto, actual.Scheme)
			assert.Equal(t, tc.host, actual.Host)
			assert.Equal(t, tc.expectedPath, actual.Path)
		})
	}
}

func TestAutheliaCtx_GetCookieDomain(t *testing.T) {
	testCases := []struct {
		name     string
		headers  map[string]string
		config   schema.Session
		expected string
		err      string
	}{
		{
			"ShouldHandleXForwarded",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "auth.example.com",
			},
			schema.Session{
				Cookies: []schema.SessionCookie{
					{
						Domain: "example.com",
					},
				},
			},
			"example.com",
			"",
		},
		{
			"ShouldHandleXForwardedError",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "a!@#*(&!@#(*uth.example.com",
			},
			schema.Session{
				Cookies: []schema.SessionCookie{
					{
						Domain: "example.com",
					},
				},
			},
			"example.com",
			`unable to retrieve cookie domain: failed to parse X-Forwarded Headers: parse "https://a!@#*(&!@#(*uth.example.com/": invalid character "#" in host name`,
		},
		{
			"ShouldHandleXOriginal",
			map[string]string{
				"X-Original-URL": "https://auth.example.com",
			},
			schema.Session{
				Cookies: []schema.SessionCookie{
					{
						Domain: "example.com",
					},
				},
			},
			"example.com",
			"",
		},
		{
			"ShouldHandleXOriginalErr",
			map[string]string{
				"X-Original-URL": "https://aut@#$*(&@#*($&@h.example.com",
			},
			schema.Session{
				Cookies: []schema.SessionCookie{
					{
						Domain: "example.com",
					},
				},
			},
			"",
			`unable to retrieve cookie domain: failed to parse X-Original-URL header: parse "https://aut@#$*(&@#*($&@h.example.com": net/url: invalid userinfo`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{
				Session: tc.config,
			}

			for k, v := range tc.headers {
				ctx.Request.Header.Set(k, v)
			}

			middleware := middlewares.NewAutheliaCtx(ctx, config, middlewares.Providers{})

			actual, err := middleware.GetCookieDomain()

			if len(tc.err) == 0 {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAutheliaCtx_GetXOriginalURLOrXForwardedURL(t *testing.T) {
	testCases := []struct {
		name     string
		headers  map[string]string
		expected string
		err      string
	}{
		{
			"ShouldHandleXForwarded",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "auth.example.com",
			},
			"https://auth.example.com/",
			"",
		},
		{
			"ShouldHandleXForwardedWithPath",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "auth.example.com",
				"X-Forwarded-URI":              "/abc",
			},
			"https://auth.example.com/abc",
			"",
		},
		{
			"ShouldHandleXForwardedError",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "a!@#*(&!@#(*uth.example.com",
			},
			"",
			`failed to parse X-Forwarded Headers: parse "https://a!@#*(&!@#(*uth.example.com/": invalid character "#" in host name`,
		},
		{
			"ShouldHandleXOriginal",
			map[string]string{
				"X-Original-URL": "https://auth.example.com/",
			},
			"https://auth.example.com/",
			"",
		},
		{
			"ShouldHandleXOriginalErr",
			map[string]string{
				"X-Original-URL": "https://aut@#$*(&@#*($&@h.example.com",
			},
			"",
			`failed to parse X-Original-URL header: parse "https://aut@#$*(&@#*($&@h.example.com": net/url: invalid userinfo`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{}
			providers := middlewares.Providers{}

			for k, v := range tc.headers {
				ctx.Request.Header.Set(k, v)
			}

			middleware := middlewares.NewAutheliaCtx(ctx, config, providers)

			actual, err := middleware.GetXOriginalURLOrXForwardedURL()

			if len(tc.err) == 0 {
				assert.NoError(t, err)

				expected, err := url.Parse(tc.expected)
				require.NoError(t, err)

				assert.Equal(t, expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAutheliaCtx_GetOrigin(t *testing.T) {
	testCases := []struct {
		name     string
		headers  map[string]string
		expected string
		err      string
	}{
		{
			"ShouldHandleXForwarded",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "auth.example.com",
			},
			"https://auth.example.com",
			"",
		},
		{
			"ShouldHandleXForwardedWithPath",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "auth.example.com",
				"X-Forwarded-URI":              "/abc",
			},
			"https://auth.example.com",
			"",
		},
		{
			"ShouldHandleXForwardedError",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "a!@#*(&!@#(*uth.example.com",
			},
			"",
			`failed to parse X-Forwarded Headers: parse "https://a!@#*(&!@#(*uth.example.com/": invalid character "#" in host name`,
		},
		{
			"ShouldHandleXOriginal",
			map[string]string{
				"X-Original-URL": "https://auth.example.com/",
			},
			"https://auth.example.com",
			"",
		},
		{
			"ShouldHandleXOriginalErr",
			map[string]string{
				"X-Original-URL": "https://aut@#$*(&@#*($&@h.example.com",
			},
			"",
			`failed to parse X-Original-URL header: parse "https://aut@#$*(&@#*($&@h.example.com": net/url: invalid userinfo`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{}
			providers := middlewares.Providers{}

			for k, v := range tc.headers {
				ctx.Request.Header.Set(k, v)
			}

			middleware := middlewares.NewAutheliaCtx(ctx, config, providers)

			actual, err := middleware.GetOrigin()

			if len(tc.err) == 0 {
				assert.NoError(t, err)

				expected, err := url.Parse(tc.expected)
				require.NoError(t, err)

				assert.Equal(t, expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAutheliaCtx_IssuerURL(t *testing.T) {
	testCases := []struct {
		name     string
		headers  map[string]string
		expected string
		err      string
	}{
		{
			"ShouldHandleXForwarded",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "auth.example.com",
			},
			"https://auth.example.com",
			"",
		},
		{
			"ShouldHandleXForwardedWithPath",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "http",
				fasthttp.HeaderXForwardedHost:  "auth.example.com",
				"X-Forwarded-URI":              "/abc",
			},
			"http://auth.example.com",
			"",
		},
		{
			"ShouldHandleXForwardedWithNoScheme",
			map[string]string{
				fasthttp.HeaderXForwardedHost: "auth.example.com",
				"X-Forwarded-URI":             "/abc",
			},
			"http://auth.example.com",
			"",
		},
		{
			"ShouldHandleXForwardedWithNoHost",
			map[string]string{
				"X-Forwarded-URI": "/abc",
			},
			"",
			"missing required X-Forwarded-Host header",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{}
			providers := middlewares.Providers{}

			for k, v := range tc.headers {
				ctx.Request.Header.Set(k, v)
			}

			middleware := middlewares.NewAutheliaCtx(ctx, config, providers)

			actual, err := middleware.IssuerURL()

			if len(tc.err) == 0 {
				assert.NoError(t, err)

				expected, err := url.Parse(tc.expected)
				require.NoError(t, err)

				assert.Equal(t, expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAutheliaCtx_AcceptsMIME(t *testing.T) {
	testCases := []struct {
		name     string
		headers  map[string]string
		mime     string
		expected bool
	}{
		{
			"ShouldHandleEmpty",
			map[string]string{},
			"text/plain",
			false,
		},
		{
			"ShouldHandleAccept",
			map[string]string{
				"Accept": "text/plain",
			},
			"text/plain",
			true,
		},
		{
			"ShouldHandleAcceptMany",
			map[string]string{
				"Accept": "application/xml;q=0.9, text/plain",
			},
			"text/plain",
			true,
		},
		{
			"ShouldHandleAcceptManyWeighted",
			map[string]string{
				"Accept": "application/xml;q=0.9, text/plain",
			},
			"application/xml",
			true,
		},
		{
			"ShouldHandleAcceptAny",
			map[string]string{
				"Accept": "*/*",
			},
			"123/456",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{}
			providers := middlewares.Providers{}

			for k, v := range tc.headers {
				ctx.Request.Header.Set(k, v)
			}

			middleware := middlewares.NewAutheliaCtx(ctx, config, providers)

			assert.Equal(t, tc.expected, middleware.AcceptsMIME(tc.mime))
		})
	}
}

func TestAutheliaCtx_GetXForwardedURL(t *testing.T) {
	testCases := []struct {
		name     string
		headers  map[string]string
		expected *url.URL
		err      string
	}{
		{
			"ShouldHandleNoHeaders",
			map[string]string{},
			nil,
			"missing required X-Forwarded-Host header",
		},
		{
			"ShouldHandlePresent",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "auth.example.com",
			},
			&url.URL{Scheme: "https", Host: "auth.example.com", Path: "/"},
			"",
		},
		{
			"ShouldHandlePresentWithPath",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "auth.example.com",
				"X-Forwarded-URI":              "/ac",
			},
			&url.URL{Scheme: "https", Host: "auth.example.com", Path: "/ac"},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{}
			providers := middlewares.Providers{}

			for k, v := range tc.headers {
				ctx.Request.Header.Set(k, v)
			}

			middleware := middlewares.NewAutheliaCtx(ctx, config, providers)

			actual, err := middleware.GetXForwardedURL()
			assert.Equal(t, tc.expected, actual)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAutheliaCtx_SetSpecialRedirect(t *testing.T) {
	testCases := []struct {
		name           string
		have           string
		status         int
		expected       string
		body           string
		expectedStatus int
	}{
		{
			"ShouldHandleTooLow",
			"https://example.com",
			123,
			"https://example.com/",
			`<a href="https://example.com/">302 Found</a>`,
			302,
		},
		{
			"ShouldHandleNormal",
			"https://example.com",
			301,
			"https://example.com/",
			`<a href="https://example.com/">301 Moved Permanently</a>`,
			301,
		},
		{
			"ShouldHandleUnauthorized",
			"https://example.com",
			401,
			"https://example.com/",
			`<a href="https://example.com/">401 Unauthorized</a>`,
			401,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{}
			providers := middlewares.Providers{}

			middleware := middlewares.NewAutheliaCtx(ctx, config, providers)

			middleware.SpecialRedirectNoBody(tc.have, tc.status)

			assert.Equal(t, tc.expected, string(ctx.Response.Header.Peek("Location")))
			assert.Equal(t, "", string(ctx.Response.Body()))
			assert.Equal(t, tc.expectedStatus, ctx.Response.StatusCode())

			middleware.SpecialRedirect(tc.have, tc.status)

			assert.Equal(t, tc.expected, string(ctx.Response.Header.Peek("Location")))
			assert.Equal(t, tc.body, string(ctx.Response.Body()))
			assert.Equal(t, tc.expectedStatus, ctx.Response.StatusCode())
		})
	}
}

func TestAutheliaCtx_GetRandom(t *testing.T) {
	testCases := []struct {
		name string
		have random.Provider
	}{
		{
			"ShouldHandleNil",
			nil,
		},
		{
			"ShouldHandleMathematical",
			random.NewMathematical(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{}
			providers := middlewares.Providers{
				Random: tc.have,
			}

			middleware := middlewares.NewAutheliaCtx(ctx, config, providers)

			assert.Equal(t, tc.have, middleware.GetRandom())
		})
	}
}

func TestAutheliaCtx_ParseBody(t *testing.T) {
	type Example struct {
		A int `json:"123" valid:"required"`
	}

	type ExampleValidated struct {
		A int    `json:"123" valid:"required"`
		B string `json:"abc" valid:"required"`
	}

	testCases := []struct {
		name        string
		have        []byte
		expected    any
		destination any
		err         string
	}{
		{
			"ShouldHandleNil",
			nil,
			nil,
			nil,
			"unable to parse body: unexpected end of JSON input",
		},
		{
			"ShouldHandleBadBody",
			[]byte("{123}"),
			nil,
			nil,
			"unable to parse body: invalid character '1' looking for beginning of object key string",
		},
		{
			"ShouldHandleBadObject",
			[]byte(`{"123": 123}`),
			nil,
			nil,
			"unable to validate body: function only accepts structs; got map",
		},
		{
			"ShouldHandleBadObject",
			[]byte(`{"123": 123}`),
			nil,
			nil,
			"unable to validate body: function only accepts structs; got map",
		},
		{
			"ShouldHandleSuccess",
			[]byte(`{"123": 123}`),
			&Example{A: 123},
			&Example{},
			"",
		},
		{
			"ShouldHandleFailValidator",
			[]byte(`{}`),
			&ExampleValidated{},
			&ExampleValidated{},
			"unable to validate body: 123: non zero value required;abc: non zero value required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{}
			providers := middlewares.Providers{}

			middleware := middlewares.NewAutheliaCtx(ctx, config, providers)

			middleware.Request.SetBody(tc.have)

			err := middleware.ParseBody(tc.destination)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, tc.destination)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAutheliaCtx_GetSessionProvider_Errors(t *testing.T) {
	testCases := []struct {
		name     string
		headers  map[string]string
		config   schema.Session
		expected string
		err      string
	}{
		{
			"ShouldHandleXForwardedError",
			map[string]string{
				fasthttp.HeaderXForwardedProto: "https",
				fasthttp.HeaderXForwardedHost:  "a!@#*(&!@#(*uth.example.com",
			},
			schema.Session{
				Cookies: []schema.SessionCookie{
					{
						Domain: "example.com",
					},
				},
			},
			"example.com",
			`unable to retrieve session cookie domain: failed to parse X-Forwarded Headers: parse "https://a!@#*(&!@#(*uth.example.com/": invalid character "#" in host name`,
		},
		{
			"ShouldHandleXOriginalErr",
			map[string]string{
				"X-Original-URL": "https://aut@#$*(&@#*($&@h.example.com",
			},
			schema.Session{
				Cookies: []schema.SessionCookie{
					{
						Domain: "example.com",
					},
				},
			},
			"",
			`unable to retrieve session cookie domain: failed to parse X-Original-URL header: parse "https://aut@#$*(&@#*($&@h.example.com": net/url: invalid userinfo`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{
				Session: tc.config,
			}

			for k, v := range tc.headers {
				ctx.Request.Header.Set(k, v)
			}

			middleware := middlewares.NewAutheliaCtx(ctx, config, middlewares.Providers{})

			actual, err := middleware.GetSessionProvider()

			if len(tc.err) == 0 {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestAutheliaCtx_GetCookieDomainSessionProvider(t *testing.T) {
	testCases := []struct {
		name   string
		domain string
		config *schema.Session
		err    string
	}{
		{
			"ShouldHandleEmptyString",
			"",
			nil,
			"unable to retrieve session cookie domain provider: no configured session cookie domain matches the domain ''",
		},
		{
			"ShouldHandleUnconfiguredProvider",
			"example.com",
			nil,
			"unable to retrieve session cookie domain provider: no session provider is configured",
		},
		{
			"ShouldHandleConfiguredProvider",
			"example.com",
			&schema.Session{
				Cookies: []schema.SessionCookie{
					{
						Domain: "example.com",
					},
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			config := schema.Configuration{}
			providers := middlewares.Providers{}

			if tc.config != nil {
				providers.SessionProvider = session.NewProvider(*tc.config, nil)
			}

			middleware := middlewares.NewAutheliaCtx(ctx, config, providers)

			actual, err := middleware.GetCookieDomainSessionProvider(tc.domain)

			if len(tc.err) == 0 {
				assert.NoError(t, err)
				assert.NotNil(t, actual)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestShouldCallNextWithAutheliaCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := &fasthttp.RequestCtx{}
	configuration := schema.Configuration{}
	userProvider := mocks.NewMockUserProvider(ctrl)
	sessionProvider := session.NewProvider(configuration.Session, nil)
	providers := middlewares.Providers{
		UserProvider:    userProvider,
		SessionProvider: sessionProvider,
		Random:          random.NewMathematical(),
	}
	nextCalled := false

	middleware := middlewares.NewBridgeBuilder(configuration, providers).Build()

	middleware(func(actx *middlewares.AutheliaCtx) {
		// Authelia context wraps the request.
		assert.Equal(t, ctx, actx.RequestCtx)

		nextCalled = true
	})(ctx)

	assert.True(t, nextCalled)
}

func TestShouldFallbackToNonXForwardedHeaders(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)

	defer mock.Close()

	mock.Ctx.Request.SetRequestURI("/2fa/one-time-password")
	mock.Ctx.Request.SetHost("auth.example.com:1234")

	assert.Equal(t, []byte("http"), mock.Ctx.XForwardedProto())
	assert.Equal(t, []byte("auth.example.com:1234"), mock.Ctx.GetXForwardedHost())
	assert.Equal(t, []byte("/2fa/one-time-password"), mock.Ctx.GetXForwardedURI())
}

func TestShouldOnlyFallbackToNonXForwardedHeadersWhenNil(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)

	mock.Ctx.Request.SetRequestURI("/2fa/one-time-password")
	mock.Ctx.Request.SetHost("localhost")
	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com:1234")
	mock.Ctx.Request.Header.Set("X-Forwarded-URI", "/base/2fa/one-time-password")
	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	mock.Ctx.Request.Header.Set("X-Forwarded-Method", fasthttp.MethodGet)

	assert.Equal(t, []byte("https"), mock.Ctx.XForwardedProto())
	assert.Equal(t, []byte("auth.example.com:1234"), mock.Ctx.GetXForwardedHost())
	assert.Equal(t, []byte("/base/2fa/one-time-password"), mock.Ctx.GetXForwardedURI())
	assert.Equal(t, []byte(fasthttp.MethodGet), mock.Ctx.XForwardedMethod())
}

func TestShouldDetectXHR(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set(fasthttp.HeaderXRequestedWith, "XMLHttpRequest")

	assert.True(t, mock.Ctx.IsXHR())
}

func TestShouldDetectNonXHR(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	assert.False(t, mock.Ctx.IsXHR())
}

func TestAutheliaCtxMisc(t *testing.T) {
	ctx := middlewares.NewAutheliaCtx(&fasthttp.RequestCtx{}, schema.Configuration{}, middlewares.Providers{})

	assert.NotNil(t, ctx.GetConfiguration())
	assert.NotNil(t, ctx.GetProviders())
	assert.NotNil(t, ctx.GetLogger())
}

func TestShouldReturnCorrectSecondFactorMethods(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Configuration.DuoAPI.Disable = true

	assert.Equal(t, []string{model.SecondFactorMethodTOTP, model.SecondFactorMethodWebAuthn}, mock.Ctx.AvailableSecondFactorMethods())

	mock.Ctx.Configuration.DuoAPI.Disable = false

	assert.Equal(t, []string{model.SecondFactorMethodTOTP, model.SecondFactorMethodWebAuthn, model.SecondFactorMethodDuo}, mock.Ctx.AvailableSecondFactorMethods())

	mock.Ctx.Configuration.TOTP.Disable = true

	assert.Equal(t, []string{model.SecondFactorMethodWebAuthn, model.SecondFactorMethodDuo}, mock.Ctx.AvailableSecondFactorMethods())

	mock.Ctx.Configuration.WebAuthn.Disable = true

	assert.Equal(t, []string{model.SecondFactorMethodDuo}, mock.Ctx.AvailableSecondFactorMethods())

	mock.Ctx.Configuration.DuoAPI.Disable = true

	assert.Equal(t, []string{}, mock.Ctx.AvailableSecondFactorMethods())
}

func TestAutheliaCtx_QueryFuncs(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.SetHost("example.com")
	mock.Ctx.Request.SetRequestURI("/?rd=example&authelia_url=example2")

	assert.Equal(t, []byte("example"), mock.Ctx.QueryArgRedirect())
	assert.Equal(t, []byte("example2"), mock.Ctx.QueryArgAutheliaURL())
}

func TestAutheliaCtx_HeaderFuncs(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Authelia-URL", "abc")
	mock.Ctx.Request.Header.Set("X-Original-Method", "cheese")

	assert.Equal(t, []byte("abc"), mock.Ctx.XAutheliaURL())
	assert.Equal(t, []byte("cheese"), mock.Ctx.XOriginalMethod())
}

func TestAutheliaCtx_SetStatusCodes(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	mock.Ctx.ReplyBadRequest()

	assert.Equal(t, 400, mock.Ctx.Response.StatusCode())

	mock.Ctx.ReplyUnauthorized()

	assert.Equal(t, 401, mock.Ctx.Response.StatusCode())

	mock.Ctx.ReplyForbidden()

	assert.Equal(t, 403, mock.Ctx.Response.StatusCode())
}

func TestAutheliaCtx_GetTargetURICookieDomain(t *testing.T) {
	testCases := []struct {
		name     string
		have     *url.URL
		config   []schema.SessionCookie
		expected string
		secure   bool
	}{
		{
			"ShouldReturnEmptyNil",
			nil,
			[]schema.SessionCookie{},
			"",
			false,
		},
		{
			"ShouldReturnEmptyNoMatch",
			&url.URL{Scheme: "https", Host: "example.com"},
			[]schema.SessionCookie{},
			"",
			false,
		},
		{
			"ShouldReturnDomain",
			&url.URL{Scheme: "https", Host: "example.com"},
			[]schema.SessionCookie{
				{
					Domain:              "example.com",
					SessionCookieCommon: schema.SessionCookieCommon{},
				},
			},
			"example.com",
			true,
		},
		{
			"ShouldReturnDomain",
			&url.URL{Scheme: "http", Host: "example.com"},
			[]schema.SessionCookie{
				{
					Domain:              "example.com",
					SessionCookieCommon: schema.SessionCookieCommon{},
				},
			},
			"example.com",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Configuration.Session.Cookies = tc.config

			assert.Equal(t, tc.expected, mock.Ctx.GetCookieDomainFromTargetURI(tc.have))
			assert.Equal(t, tc.secure, mock.Ctx.IsSafeRedirectionTargetURI(tc.have))
		})
	}
}

func TestAutheliaCtx_GetDefaultRedirectionURL(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.example4.com/consent")

	assert.Nil(t, mock.Ctx.GetDefaultRedirectionURL())

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.example.com/consent")

	assert.Equal(t, &url.URL{Scheme: "https", Host: "www.example.com"}, mock.Ctx.GetDefaultRedirectionURL())

	mock2 := mocks.NewMockAutheliaCtx(t)
	defer mock2.Close()

	mock2.Ctx.Request.Header.Set("X-Original-URL", "https://auth.example2.com/consent")

	assert.Equal(t, &url.URL{Scheme: "https", Host: "www.example2.com"}, mock2.Ctx.GetDefaultRedirectionURL())
}
