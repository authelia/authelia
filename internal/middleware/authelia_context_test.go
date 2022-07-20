package middleware_test

import (
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middleware"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestShouldCallNextWithAutheliaCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.Configuration{}
	userProvider := mocks.NewMockUserProvider(ctrl)
	sessionProvider := session.NewProvider(configuration.Session, nil)
	providers := middleware.Providers{
		UserProvider:    userProvider,
		SessionProvider: sessionProvider,
	}
	nextCalled := false

	bridge := middleware.NewBridgeBuilder(configuration, providers).Build()

	bridge(func(actx *middleware.AutheliaCtx) {
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
	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "https")
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "home.example.com")
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
	mock.Ctx.RequestCtx.Request.Header.Set("X-Forwarded-Proto", "https")
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
