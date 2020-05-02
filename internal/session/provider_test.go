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
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration)
	session, err := provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, NewDefaultUserSession(), session)
}

func TestShouldUpdateSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration)
	session, _ := provider.GetSession(ctx)

	session.Username = testUsername
	session.AuthenticationLevel = authentication.TwoFactor

	err := provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, UserSession{
		Username:            testUsername,
		AuthenticationLevel: authentication.TwoFactor,
	}, session)
}

func TestShouldDestroySessionAndWipeSessionData(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration)
	session, err := provider.GetSession(ctx)
	require.NoError(t, err)

	session.Username = testUsername
	session.AuthenticationLevel = authentication.TwoFactor

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	newUserSession, err := provider.GetSession(ctx)
	require.NoError(t, err)
	assert.Equal(t, testUsername, newUserSession.Username)
	assert.Equal(t, authentication.TwoFactor, newUserSession.AuthenticationLevel)

	err = provider.DestroySession(ctx)
	require.NoError(t, err)

	newUserSession, err = provider.GetSession(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", newUserSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, newUserSession.AuthenticationLevel)
}
