package middlewares_test

import (
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
	originalURL, err := mock.Ctx.GetOriginalURL()
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
	originalURL, err := mock.Ctx.GetOriginalURL()
	assert.NoError(t, err)

	expectedURL, err := url.ParseRequestURI("https://home.example.com/")
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)
}

func TestShouldGetOriginalURLFromForwardedHeadersWithURI(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	mock.Ctx.Request.Header.Set("X-Original-URL", "htt-ps//home?-.example.com")
	_, err := mock.Ctx.GetOriginalURL()
	assert.Error(t, err)
	assert.Equal(t, "Unable to parse URL extracted from X-Original-URL header: parse \"htt-ps//home?-.example.com\": invalid URI for request", err.Error())
}

func TestShouldFallbackToNonXForwardedHeaders(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.RequestCtx.Request.SetRequestURI("/2fa/one-time-password")
	mock.Ctx.RequestCtx.Request.SetHost("auth.example.com:1234")

	assert.Equal(t, []byte("http"), mock.Ctx.XForwardedProto())
	assert.Equal(t, []byte("auth.example.com:1234"), mock.Ctx.XForwardedHost())
	assert.Equal(t, []byte("/2fa/one-time-password"), mock.Ctx.XForwardedURI())
}

func TestShouldOnlyFallbackToNonXForwardedHeadersWhenNil(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.RequestCtx.Request.SetRequestURI("/2fa/one-time-password")
	mock.Ctx.RequestCtx.Request.SetHost("localhost")
	mock.Ctx.RequestCtx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com:1234")
	mock.Ctx.RequestCtx.Request.Header.Set("X-Forwarded-URI", "/base/2fa/one-time-password")
	mock.Ctx.RequestCtx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	mock.Ctx.RequestCtx.Request.Header.Set("X-Forwarded-Method", "GET")

	assert.Equal(t, []byte("https"), mock.Ctx.XForwardedProto())
	assert.Equal(t, []byte("auth.example.com:1234"), mock.Ctx.XForwardedHost())
	assert.Equal(t, []byte("/base/2fa/one-time-password"), mock.Ctx.XForwardedURI())
	assert.Equal(t, []byte("GET"), mock.Ctx.XForwardedMethod())
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

	assert.Equal(t, []string{model.SecondFactorMethodTOTP, model.SecondFactorMethodWebauthn}, mock.Ctx.AvailableSecondFactorMethods())

	mock.Ctx.Configuration.DuoAPI.Disable = false

	assert.Equal(t, []string{model.SecondFactorMethodTOTP, model.SecondFactorMethodWebauthn, model.SecondFactorMethodDuo}, mock.Ctx.AvailableSecondFactorMethods())

	mock.Ctx.Configuration.TOTP.Disable = true

	assert.Equal(t, []string{model.SecondFactorMethodWebauthn, model.SecondFactorMethodDuo}, mock.Ctx.AvailableSecondFactorMethods())

	mock.Ctx.Configuration.Webauthn.Disable = true

	assert.Equal(t, []string{model.SecondFactorMethodDuo}, mock.Ctx.AvailableSecondFactorMethods())

	mock.Ctx.Configuration.DuoAPI.Disable = true

	assert.Equal(t, []string{}, mock.Ctx.AvailableSecondFactorMethods())
}
