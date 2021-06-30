package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldInitializerSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration, nil)
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

	provider := NewProvider(configuration, nil)
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

func TestShouldSetSessionAuthenticationLevels(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}

	timeOneFactor := time.Unix(1625048140, 0)
	timeTwoFactor := time.Unix(1625048150, 0)
	timeFalseFactor := time.Unix(1625048160, 0)

	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration, nil)
	session, _ := provider.GetSession(ctx)

	session.SetOneFactor(timeOneFactor, &authentication.UserDetails{Username: testUsername}, false)

	err := provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, timeOneFactor, session.AuthenticatedAt(authorization.OneFactor))
	assert.Equal(t, time.Unix(0, 0), session.AuthenticatedAt(authorization.TwoFactor))
	assert.Equal(t, timeOneFactor, session.AuthenticatedAt(authorization.Denied))

	assert.Equal(t, UserSession{
		Username:            testUsername,
		AuthenticationLevel: authentication.OneFactor,
		LastActivity:        timeOneFactor.Unix(),
		FirstFactorAuthn:    timeOneFactor.Unix(),
	}, session)

	session.SetOneFactor(timeTwoFactor, &authentication.UserDetails{Username: testUsername}, false)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, UserSession{
		Username:            testUsername,
		AuthenticationLevel: authentication.OneFactor,
		LastActivity:        timeTwoFactor.Unix(),
		FirstFactorAuthn:    timeOneFactor.Unix(),
	}, session)

	assert.Equal(t, timeOneFactor, session.AuthenticatedAt(authorization.OneFactor))
	assert.Equal(t, time.Unix(0, 0), session.AuthenticatedAt(authorization.TwoFactor))

	session.SetTwoFactor(timeTwoFactor)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, UserSession{
		Username:            testUsername,
		AuthenticationLevel: authentication.TwoFactor,
		LastActivity:        timeTwoFactor.Unix(),
		FirstFactorAuthn:    timeOneFactor.Unix(),
		SecondFactorAuthn:   timeTwoFactor.Unix(),
	}, session)

	assert.Equal(t, timeOneFactor, session.AuthenticatedAt(authorization.OneFactor))
	assert.Equal(t, timeTwoFactor, session.AuthenticatedAt(authorization.TwoFactor))
	assert.Equal(t, timeTwoFactor, session.AuthenticatedAt(authorization.Denied))

	session.SetTwoFactor(timeFalseFactor)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, UserSession{
		Username:            testUsername,
		AuthenticationLevel: authentication.TwoFactor,
		LastActivity:        timeFalseFactor.Unix(),
		FirstFactorAuthn:    timeOneFactor.Unix(),
		SecondFactorAuthn:   timeTwoFactor.Unix(),
	}, session)

	assert.Equal(t, timeOneFactor, session.AuthenticatedAt(authorization.OneFactor))
	assert.Equal(t, timeTwoFactor, session.AuthenticatedAt(authorization.TwoFactor))
}

func TestShouldDestroySessionAndWipeSessionData(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration, nil)
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
