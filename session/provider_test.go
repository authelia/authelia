package session

import (
	"testing"

	"github.com/clems4ever/authelia/authentication"

	"github.com/stretchr/testify/assert"

	"github.com/valyala/fasthttp"

	"github.com/clems4ever/authelia/configuration/schema"
)

func TestShouldInitializerSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = "example.com"
	configuration.Name = "my_session"
	configuration.Expiration = 40

	provider := NewProvider(configuration)
	session, _ := provider.GetSession(ctx)

	assert.Equal(t, NewDefaultUserSession(), session)
}

func TestShouldUpdateSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = "example.com"
	configuration.Name = "my_session"
	configuration.Expiration = 40

	provider := NewProvider(configuration)
	session, _ := provider.GetSession(ctx)

	session.Username = "john"
	session.AuthenticationLevel = authentication.TwoFactor

	_ = provider.SaveSession(ctx, session)

	session, _ = provider.GetSession(ctx)

	assert.Equal(t, UserSession{
		Username:            "john",
		AuthenticationLevel: authentication.TwoFactor,
	}, session)
}
