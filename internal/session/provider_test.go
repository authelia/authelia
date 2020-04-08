package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldInitializerSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = "example.com"
	configuration.Name = "my_session"
	configuration.Expiration = "40"

	provider := NewProvider(configuration)
	session, err := provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, NewDefaultUserSession(), session)
}

func TestShouldUpdateSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = "example.com"
	configuration.Name = "my_session"
	configuration.Expiration = "40"

	provider := NewProvider(configuration)
	session, _ := provider.GetSession(ctx)

	session.Username = "john"
	session.AuthenticationLevel = authentication.TwoFactor

	err := provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, UserSession{
		Username:            "john",
		AuthenticationLevel: authentication.TwoFactor,
	}, session)
}

func TestShouldDestroySessionAndWipeSessionData(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = "example.com"
	configuration.Name = "my_session"
	configuration.Expiration = "40"

	provider := NewProvider(configuration)
	session, err := provider.GetSession(ctx)
	require.NoError(t, err)

	session.Username = "john"
	session.AuthenticationLevel = authentication.TwoFactor

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	newUserSession, err := provider.GetSession(ctx)
	require.NoError(t, err)
	assert.Equal(t, "john", newUserSession.Username)
	assert.Equal(t, authentication.TwoFactor, newUserSession.AuthenticationLevel)

	err = provider.DestroySession(ctx)
	require.NoError(t, err)

	newUserSession, err = provider.GetSession(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", newUserSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, newUserSession.AuthenticationLevel)
}
