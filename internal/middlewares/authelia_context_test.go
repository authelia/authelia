package middlewares_test

import (
	"net"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

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

func TestIssuerURL(t *testing.T) {
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
				mock.Ctx.SetUserValue("base_url", tc.base)
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

func TestShouldCallNextWithAutheliaCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
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
