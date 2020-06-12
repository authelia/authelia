package middlewares_test

import (
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/mocks"
	"github.com/authelia/authelia/internal/session"
)

func TestShouldCallNextWithAutheliaCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.Configuration{}
	userProvider := mocks.NewMockUserProvider(ctrl)
	sessionProvider := session.NewProvider(configuration.Session)
	providers := middlewares.Providers{
		UserProvider:    userProvider,
		SessionProvider: sessionProvider,
	}
	nextCalled := false

	middlewares.AutheliaMiddleware(configuration, providers)(func(actx *middlewares.AutheliaCtx) {
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

	expectedURL, err := url.ParseRequestURI("https://home.example.com")
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
