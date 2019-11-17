package middlewares_test

import (
	"testing"

	"github.com/clems4ever/authelia/internal/session"

	"github.com/clems4ever/authelia/internal/configuration/schema"
	"github.com/clems4ever/authelia/internal/middlewares"
	"github.com/clems4ever/authelia/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
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
