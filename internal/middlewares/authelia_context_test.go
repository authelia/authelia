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

func TestShouldExtractXRealIPAsRemoteIP(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	autheliaCtx := middlewares.AutheliaCtx{
		RequestCtx: ctx,
	}
	assert.Equal(t, "0.0.0.0", autheliaCtx.RemoteIP().String())

	ctx.Request.Header.Add("X-Forwarded-For", "10.0.0.1 , 192.168.0.1, 127.0.0.1")
	assert.Equal(t, "10.0.0.1", autheliaCtx.RemoteIP().String())

	ctx.Request.Header.Add("X-Real-Ip", "10.2.0.1")
	assert.Equal(t, "10.2.0.1", autheliaCtx.RemoteIP().String())
}
