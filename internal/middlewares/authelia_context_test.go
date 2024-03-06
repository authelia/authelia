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
				mock.Ctx.RequestCtx.Request.Header.SetBytesV(fasthttp.HeaderXForwardedFor, tc.have)
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

func TestAutheliaCtx_IssuerURL(t *testing.T) {
	testCases := []struct {
		name              string
		proto, host, base string
		expectedProto     string
		expected          string
		err               string
	}{
		{
			name:  "Standard",
			proto: "https", host: "auth.example.com", base: "",
			expected: "https://auth.example.com",
		},
		{
			name:  "StandardHTTP",
			proto: "http", host: "auth.example.com", base: "",
			expected: "http://auth.example.com",
		},
		{
			name:  "NoProto",
			proto: "", host: "auth.example.com", base: "",
			expected:      "https://auth.example.com",
			expectedProto: "https",
		},
		{
			name:  "Base",
			proto: "https", host: "example.com", base: "auth",
			expected: "https://example.com/auth",
		},
		{
			name:  "NoHost",
			proto: "https", host: "", base: "",
			err: "missing required X-Forwarded-Host header",
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

			actual, theError := mock.Ctx.IssuerURL()

			if len(tc.err) == 0 {
				assert.NoError(t, theError)
				require.NotNil(t, actual)

				assert.Equal(t, tc.expected, actual.String())

				if len(tc.expectedProto) == 0 {
					assert.Equal(t, tc.proto, actual.Scheme)
				} else {
					assert.Equal(t, tc.expectedProto, actual.Scheme)
				}

				assert.Equal(t, tc.host, actual.Host)
				assert.Equal(t, tc.base, actual.Path)
			} else {
				assert.Nil(t, actual)
				assert.EqualError(t, theError, tc.err)
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

// Test getOriginalURL.
func TestShouldGetOriginalURLFromOriginalURLHeader(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://home.example.com")
	originalURL, err := mock.Ctx.GetXOriginalURLOrXForwardedURL()
	assert.NoError(t, err)

	expectedURL, err := url.ParseRequestURI("https://home.example.com")
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)
}

func TestShouldGetOriginalURLFromForwardedHeadersWithoutURI(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "home.example.com")
	originalURL, err := mock.Ctx.GetXOriginalURLOrXForwardedURL()
	assert.NoError(t, err)

	expectedURL, err := url.ParseRequestURI("https://home.example.com/")
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)
}

func TestShouldGetOriginalURLFromForwardedHeadersWithURI(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	mock.Ctx.Request.Header.Set("X-Original-URL", "htt-ps//home?-.example.com")
	_, err := mock.Ctx.GetXOriginalURLOrXForwardedURL()
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to parse X-Original-URL header: parse \"htt-ps//home?-.example.com\": invalid URI for request")
}

func TestShouldFallbackToNonXForwardedHeaders(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)

	defer mock.Close()

	mock.Ctx.RequestCtx.Request.SetRequestURI("/2fa/one-time-password")
	mock.Ctx.RequestCtx.Request.SetHost("auth.example.com:1234")

	assert.Equal(t, []byte("http"), mock.Ctx.XForwardedProto())
	assert.Equal(t, []byte("auth.example.com:1234"), mock.Ctx.GetXForwardedHost())
	assert.Equal(t, []byte("/2fa/one-time-password"), mock.Ctx.GetXForwardedURI())
}

func TestShouldOnlyFallbackToNonXForwardedHeadersWhenNil(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)

	mock.Ctx.RequestCtx.Request.SetRequestURI("/2fa/one-time-password")
	mock.Ctx.RequestCtx.Request.SetHost("localhost")
	mock.Ctx.RequestCtx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com:1234")
	mock.Ctx.RequestCtx.Request.Header.Set("X-Forwarded-URI", "/base/2fa/one-time-password")
	mock.Ctx.RequestCtx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	mock.Ctx.RequestCtx.Request.Header.Set("X-Forwarded-Method", fasthttp.MethodGet)

	assert.Equal(t, []byte("https"), mock.Ctx.XForwardedProto())
	assert.Equal(t, []byte("auth.example.com:1234"), mock.Ctx.GetXForwardedHost())
	assert.Equal(t, []byte("/base/2fa/one-time-password"), mock.Ctx.GetXForwardedURI())
	assert.Equal(t, []byte(fasthttp.MethodGet), mock.Ctx.XForwardedMethod())
}

func TestShouldDetectXHR(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.RequestCtx.Request.Header.Set(fasthttp.HeaderXRequestedWith, "XMLHttpRequest")

	assert.True(t, mock.Ctx.IsXHR())
}

func TestShouldDetectNonXHR(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	assert.False(t, mock.Ctx.IsXHR())
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
