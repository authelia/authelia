package session

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func mustParseURL(rawURL string) *url.URL {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return parsedURL
}

func newTestSession() (*Session, error) {
	config := schema.Session{}
	config.Cookies = []schema.SessionCookie{
		{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:       testName,
				Expiration: testExpiration,
			},
			Domain: testDomain,
		},
	}

	provider := NewProvider(config, nil)

	return provider.Get(testDomain)
}

func TestShouldInitializerSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	provider, err := newTestSession()
	assert.NoError(t, err)

	session, err := provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t, provider.NewDefaultUserSession(), session)
}

func TestShouldUpdateSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	provider, err := newTestSession()
	assert.NoError(t, err)

	session, _ := provider.GetSession(ctx)

	session.Username = testUsername
	session.AuthenticationMethodRefs.UsernameAndPassword = true
	session.AuthenticationMethodRefs.WebAuthn = true

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t, UserSession{
		CookieDomain: testDomain,
		Username:     testUsername,
		AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
			UsernameAndPassword: true,
			WebAuthn:            true,
		},
	}, session)

	assert.Equal(t, authentication.TwoFactor, session.AuthenticationLevel(false))
}

func TestShouldSetSessionAuthenticationLevels(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	timeOneFactor := time.Unix(1625048140, 0).UTC()
	timeTwoFactor := time.Unix(1625048150, 0).UTC()
	timeZeroFactor := time.Unix(0, 0).UTC()

	provider, err := newTestSession()
	assert.NoError(t, err)

	session, _ := provider.GetSession(ctx)

	session.SetOneFactorPassword(timeOneFactor, &authentication.UserDetails{Username: testUsername}, false)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	authAt, err := session.AuthenticatedTime(authorization.OneFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeOneFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.TwoFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeZeroFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.Denied)
	assert.EqualError(t, err, "invalid authorization level")
	assert.Equal(t, timeZeroFactor, authAt)

	assert.Equal(t, UserSession{
		CookieDomain:              testDomain,
		Username:                  testUsername,
		LastActivity:              timeOneFactor.Unix(),
		FirstFactorAuthnTimestamp: timeOneFactor.Unix(),
		AuthenticationMethodRefs:  authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, KnowledgeBasedAuthentication: true},
	}, session)

	assert.Equal(t, authentication.OneFactor, session.AuthenticationLevel(false))

	session.SetTwoFactorDuo(timeTwoFactor)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t, UserSession{
		CookieDomain:               testDomain,
		Username:                   testUsername,
		LastActivity:               timeTwoFactor.Unix(),
		FirstFactorAuthnTimestamp:  timeOneFactor.Unix(),
		SecondFactorAuthnTimestamp: timeTwoFactor.Unix(),
		AuthenticationMethodRefs:   authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, Duo: true, KnowledgeBasedAuthentication: true},
	}, session)

	assert.Equal(t, authentication.TwoFactor, session.AuthenticationLevel(false))

	authAt, err = session.AuthenticatedTime(authorization.OneFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeOneFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.TwoFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeTwoFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.Denied)
	assert.EqualError(t, err, "invalid authorization level")
	assert.Equal(t, timeZeroFactor, authAt)
}

func TestShouldSetSessionAuthenticationLevelsAMR(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	timeOneFactor := time.Unix(1625048140, 0).UTC()
	timeTwoFactor := time.Unix(1625048150, 0).UTC()
	timeZeroFactor := time.Unix(0, 0).UTC()

	provider, err := newTestSession()
	assert.NoError(t, err)

	session, _ := provider.GetSession(ctx)

	session.SetOneFactorPassword(timeOneFactor, &authentication.UserDetails{Username: testUsername}, false)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	authAt, err := session.AuthenticatedTime(authorization.OneFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeOneFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.TwoFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeZeroFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.Denied)
	assert.EqualError(t, err, "invalid authorization level")
	assert.Equal(t, timeZeroFactor, authAt)

	assert.Equal(t, UserSession{
		CookieDomain:              testDomain,
		Username:                  testUsername,
		LastActivity:              timeOneFactor.Unix(),
		FirstFactorAuthnTimestamp: timeOneFactor.Unix(),
		AuthenticationMethodRefs:  authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, KnowledgeBasedAuthentication: true},
	}, session)

	assert.Equal(t, authentication.OneFactor, session.AuthenticationLevel(false))

	session.SetTwoFactorWebAuthn(timeTwoFactor, true, false, false)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t, authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, WebAuthn: true, WebAuthnHardware: true, KnowledgeBasedAuthentication: true}, session.AuthenticationMethodRefs)
	assert.True(t, session.AuthenticationMethodRefs.MultiFactorAuthentication())

	authAt, err = session.AuthenticatedTime(authorization.OneFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeOneFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.TwoFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeTwoFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.Denied)
	assert.EqualError(t, err, "invalid authorization level")
	assert.Equal(t, timeZeroFactor, authAt)

	session.SetTwoFactorWebAuthn(timeTwoFactor, true, false, false)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t,
		authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, WebAuthn: true, WebAuthnHardware: true, KnowledgeBasedAuthentication: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebAuthn(timeTwoFactor, true, false, false)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t,
		authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, WebAuthn: true, WebAuthnHardware: true, KnowledgeBasedAuthentication: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebAuthn(timeTwoFactor, true, true, false)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t,
		authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, WebAuthn: true, WebAuthnUserPresence: true, WebAuthnHardware: true, KnowledgeBasedAuthentication: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebAuthn(timeTwoFactor, true, true, false)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t,
		authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, WebAuthn: true, WebAuthnUserPresence: true, WebAuthnHardware: true, KnowledgeBasedAuthentication: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebAuthn(timeTwoFactor, true, false, true)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t,
		authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, WebAuthn: true, WebAuthnUserVerified: true, WebAuthnHardware: true, KnowledgeBasedAuthentication: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebAuthn(timeTwoFactor, true, false, true)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t,
		authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, WebAuthn: true, WebAuthnUserVerified: true, WebAuthnHardware: true, KnowledgeBasedAuthentication: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorTOTP(timeTwoFactor)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t,
		authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, TOTP: true, WebAuthn: true, WebAuthnUserVerified: true, WebAuthnHardware: true, KnowledgeBasedAuthentication: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorTOTP(timeTwoFactor)

	err = provider.SaveSession(ctx, session)
	assert.NoError(t, err)

	session, err = provider.GetSession(ctx)
	assert.NoError(t, err)

	assert.Equal(t,
		authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, TOTP: true, WebAuthn: true, WebAuthnUserVerified: true, WebAuthnHardware: true, KnowledgeBasedAuthentication: true},
		session.AuthenticationMethodRefs)
}

func TestShouldDestroySessionAndWipeSessionData(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	domainSession, err := newTestSession()
	assert.NoError(t, err)

	session, err := domainSession.GetSession(ctx)
	assert.NoError(t, err)

	session.Username = testUsername
	session.AuthenticationMethodRefs.UsernameAndPassword = true
	session.AuthenticationMethodRefs.WebAuthn = true

	err = domainSession.SaveSession(ctx, session)
	assert.NoError(t, err)

	newUserSession, err := domainSession.GetSession(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testUsername, newUserSession.Username)
	assert.Equal(t, authentication.TwoFactor, newUserSession.AuthenticationLevel(false))

	err = domainSession.DestroySession(ctx)
	assert.NoError(t, err)

	newUserSession, err = domainSession.GetSession(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "", newUserSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, newUserSession.AuthenticationLevel(false))
}

func TestStartupCheckShouldSucceedForMemoryBackend(t *testing.T) {
	config := schema.Session{}
	config.Cookies = []schema.SessionCookie{
		{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:       testName,
				Expiration: testExpiration,
			},
			Domain: testDomain,
		},
	}

	provider := NewProvider(config, nil)

	assert.NoError(t, provider.StartupCheck())
}

func TestStartupCheckShouldSucceedForMemoryBackendWithMultipleDomains(t *testing.T) {
	config := schema.Session{}
	config.Cookies = []schema.SessionCookie{
		{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:       testName,
				Expiration: testExpiration,
			},
			Domain: "first.example.com",
		},
		{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:       testName,
				Expiration: testExpiration,
			},
			Domain: "second.example.com",
		},
	}

	provider := NewProvider(config, nil)

	assert.NoError(t, provider.StartupCheck())
	assert.Len(t, provider.sessions, 2)
}

func TestStartupCheckShouldShareMemoryBackendWithDuplicateCookieDomains(t *testing.T) {
	config := schema.Session{}
	config.Cookies = []schema.SessionCookie{
		{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:       testName,
				Expiration: testExpiration,
			},
			Domain:      testDomain,
			AutheliaURL: mustParseURL("https://auth.example.com"),
		},
		{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:       testName,
				Expiration: testExpiration,
			},
			Domain:      testDomain,
			AutheliaURL: mustParseURL("https://auth.example.com:8443"),
		},
	}

	provider := NewProvider(config, nil)

	assert.NoError(t, provider.StartupCheck())
	assert.Len(t, provider.sessions, 1)

	domainSession, err := provider.Get(testDomain)
	assert.NoError(t, err)
	assert.Equal(t, testDomain, domainSession.Config.Domain)
	assert.Equal(t, "auth.example.com", domainSession.Config.AutheliaURL.Host)
}

func TestStartupCheckShouldSurfaceConstructionError(t *testing.T) {
	construction := fmt.Errorf("backend connection refused")

	provider := &Provider{errStartup: construction}

	err := provider.StartupCheck()
	assert.ErrorIs(t, err, construction)
}
