package oidc_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	xjwt "authelia.com/provider/oauth2/token/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestNewSession(t *testing.T) {
	session := oidc.NewSession()

	require.NotNil(t, session)

	assert.Equal(t, "", session.ClientID)
	assert.Equal(t, "", session.Username)
	assert.Equal(t, "", session.Subject)
	require.NotNil(t, session.Claims)
	assert.NotNil(t, session.Claims.Extra)
	assert.NotNil(t, session.Extra)
	require.NotNil(t, session.Headers)
	assert.NotNil(t, session.Headers.Extra)
}

func TestNewSessionWithAuthorizeRequest(t *testing.T) {
	requestID := uuid.New()
	subject := uuid.New()

	formValues := url.Values{}

	formValues.Set(oidc.ClaimNonce, "abc123xyzauthelia")

	request := &oauthelia2.AuthorizeRequest{
		Request: oauthelia2.Request{
			ID:     requestID.String(),
			Form:   formValues,
			Client: &oidc.RegisteredClient{ID: "example"},
		},
	}

	extra := map[string]any{
		oidc.ClaimPreferredUsername: "john",
	}

	requested := time.Unix(1647332518, 0)
	authAt := time.Unix(1647332500, 0)
	issuer := examplecom
	amr := []string{oidc.AMRPasswordBasedAuthentication}

	consent := &model.OAuth2ConsentSession{
		ChallengeID: uuid.New(),
		RequestedAt: requested,
		Subject:     uuid.NullUUID{UUID: subject, Valid: true},
	}

	ctx := &TestContext{}

	ctx.Clock = clock.NewFixed(time.Unix(10000000000, 0))

	session := oidc.NewSessionWithRequester(ctx, MustParseRequestURI(issuer), "primary", "john", amr, extra, authAt, consent, request, nil)

	require.NotNil(t, session)
	require.NotNil(t, session.Extra)
	require.NotNil(t, session.Headers)
	require.NotNil(t, session.Headers.Extra)
	require.NotNil(t, session.Claims)
	require.NotNil(t, session.Claims.Extra)
	require.NotNil(t, session.Claims.AuthenticationMethodsReferences)

	assert.Equal(t, subject.String(), session.Subject)
	assert.Equal(t, "example", session.ClientID)
	assert.Greater(t, session.Claims.IssuedAt.Unix(), authAt.Unix())
	assert.Equal(t, "john", session.Username)

	assert.Equal(t, "abc123xyzauthelia", session.Claims.Nonce)
	assert.Equal(t, subject.String(), session.Claims.Subject)
	assert.Equal(t, amr, session.Claims.AuthenticationMethodsReferences)
	assert.Equal(t, xjwt.NewNumericDate(authAt.UTC()), session.Claims.AuthTime)
	assert.Equal(t, issuer, session.Claims.Issuer)
	assert.Equal(t, "john", session.Claims.Extra[oidc.ClaimPreferredUsername])

	assert.Equal(t, "primary", session.Headers.Get(oidc.JWTHeaderKeyIdentifier))

	claims := &oidc.ClaimsRequests{
		IDToken: map[string]*oidc.ClaimRequest{
			oidc.ClaimSubject:  {},
			oidc.ClaimFullName: {},
		},
		UserInfo: map[string]*oidc.ClaimRequest{
			oidc.ClaimEmail:       {},
			oidc.ClaimPhoneNumber: {},
		},
	}

	session = oidc.NewSessionWithRequester(ctx, MustParseRequestURI(issuer), "primary", "john", amr, extra, authAt, consent, request, claims)

	require.NotNil(t, session)
	require.NotNil(t, session.Extra)
	require.NotNil(t, session.Headers)
	require.NotNil(t, session.Headers.Extra)
	require.NotNil(t, session.Claims)
	require.NotNil(t, session.Claims.Extra)
	require.NotNil(t, session.Claims.AuthenticationMethodsReferences)

	assert.Equal(t, subject.String(), session.Subject)
	assert.Equal(t, "example", session.ClientID)
	assert.Greater(t, session.Claims.IssuedAt.Unix(), authAt.Unix())
	assert.Equal(t, "john", session.Username)

	assert.Equal(t, "abc123xyzauthelia", session.Claims.Nonce)
	assert.Equal(t, subject.String(), session.Claims.Subject)
	assert.Equal(t, amr, session.Claims.AuthenticationMethodsReferences)
	assert.Equal(t, authAt.UTC(), session.Claims.AuthTime.Time)
	assert.Equal(t, issuer, session.Claims.Issuer)
	assert.Equal(t, "john", session.Claims.Extra[oidc.ClaimPreferredUsername])

	assert.Equal(t, "primary", session.Headers.Get(oidc.JWTHeaderKeyIdentifier))

	consent = &model.OAuth2ConsentSession{
		ChallengeID: uuid.New(),
		RequestedAt: requested,
	}

	session = oidc.NewSessionWithRequester(ctx, MustParseRequestURI(issuer), "primary", "john", nil, nil, authAt, consent, request, nil)

	require.NotNil(t, session)
	require.NotNil(t, session.Claims)
	assert.NotNil(t, session.Claims.Extra)
	assert.Nil(t, session.Claims.AuthenticationMethodsReferences)
}

// TestContext is a minimal implementation of Context for the purpose of testing.
type TestContext struct {
	context.Context

	MockIssuerURL *url.URL
	IssuerURLFunc func() (issuerURL *url.URL, err error)
	Clock         clock.Provider
	Config        schema.Configuration
}

func (m *TestContext) Value(key any) any {
	if key == model.CtxKeyAutheliaCtx {
		return m
	}

	return m.Context.Value(key)
}

func (m *TestContext) GetRandom() (r random.Provider) {
	return random.NewMathematical()
}

func (m *TestContext) GetConfiguration() (config schema.Configuration) {
	return m.Config
}

func (m *TestContext) GetProviderUserAttributeResolver() expression.UserAttributeResolver {
	return &expression.UserAttributes{}
}

// IssuerURL returns the MockIssuerURL.
func (m *TestContext) RootURL() (issuerURL *url.URL) {
	if m.IssuerURLFunc != nil {
		if issuer, err := m.IssuerURLFunc(); err != nil {
			panic(err)
		} else {
			return issuer
		}
	}

	return m.MockIssuerURL
}

// IssuerURL returns the MockIssuerURL.
func (m *TestContext) IssuerURL() (issuerURL *url.URL, err error) {
	if m.IssuerURLFunc != nil {
		return m.IssuerURLFunc()
	}

	return m.MockIssuerURL, nil
}

func (m *TestContext) GetClock() clock.Provider {
	if m.Clock != nil {
		return m.Clock
	}

	return clock.New()
}

func (m *TestContext) GetJWTWithTimeFuncOption() jwt.ParserOption {
	return jwt.WithTimeFunc(m.GetClock().Now)
}

func (m *TestContext) GetProviderStorage() storage.Provider {
	return nil
}

func (m *TestContext) GetProviderUser() authentication.UserProvider {
	return nil
}
