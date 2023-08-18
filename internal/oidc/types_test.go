package oidc_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	fjwt "github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
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

	request := &fosite.AuthorizeRequest{
		Request: fosite.Request{
			ID:     requestID.String(),
			Form:   formValues,
			Client: &oidc.BaseClient{ID: "example"},
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

	session := oidc.NewSessionWithAuthorizeRequest(MustParseRequestURI(issuer), "primary", "john", amr, extra, authAt, consent, request)

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
	assert.Equal(t, authAt, session.Claims.AuthTime)
	assert.Equal(t, requested, session.Claims.RequestedAt)
	assert.Equal(t, issuer, session.Claims.Issuer)
	assert.Equal(t, "john", session.Claims.Extra[oidc.ClaimPreferredUsername])

	assert.Equal(t, "primary", session.Headers.Get(oidc.JWTHeaderKeyIdentifier))

	consent = &model.OAuth2ConsentSession{
		ChallengeID: uuid.New(),
		RequestedAt: requested,
	}

	session = oidc.NewSessionWithAuthorizeRequest(MustParseRequestURI(issuer), "primary", "john", nil, nil, authAt, consent, request)

	require.NotNil(t, session)
	require.NotNil(t, session.Claims)
	assert.NotNil(t, session.Claims.Extra)
	assert.Nil(t, session.Claims.AuthenticationMethodsReferences)
}

func TestPopulateClientCredentialsFlowSessionWithAccessRequest(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(ctx oidc.Context)
		ctx      oidc.Context
		request  fosite.AccessRequester
		have     *model.OpenIDSession
		expected *model.OpenIDSession
		err      string
	}{
		{
			"ShouldHandleIssuerError",
			nil,
			&TestContext{
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return nil, errors.New("an error")
				},
			},
			&fosite.AccessRequest{},
			oidc.NewSession(),
			nil,
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Failed to determine the issuer with error: an error.",
		},
		{
			"ShouldHandleClientError",
			nil,
			&TestContext{
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return &url.URL{Scheme: "https", Host: "example.com"}, nil
				},
			},
			&fosite.AccessRequest{},
			oidc.NewSession(),
			nil,
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Failed to get the client for the request.",
		},
		{
			"ShouldUpdateValues",
			func(ctx oidc.Context) {
				c := ctx.(*TestContext)

				clock := &utils.TestingClock{}

				clock.Set(time.Unix(10000000000, 0))

				c.Clock = clock
			},
			&TestContext{
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return &url.URL{Scheme: "https", Host: "example.com"}, nil
				},
			},
			&fosite.AccessRequest{
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "abc",
					},
				},
			},
			oidc.NewSession(),
			&model.OpenIDSession{
				Extra: map[string]any{},
				DefaultSession: &openid.DefaultSession{
					Headers: &fjwt.Headers{
						Extra: map[string]any{},
					},
					Claims: &fjwt.IDTokenClaims{
						Issuer:      "https://example.com",
						IssuedAt:    time.Unix(10000000000, 0),
						RequestedAt: time.Unix(10000000000, 0),
						Subject:     "abc",
						Extra:       map[string]any{},
					},
				},
				ClientID: "abc",
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(tc.ctx)
			}

			err := oidc.PopulateClientCredentialsFlowSessionWithAccessRequest(tc.ctx, tc.request, tc.have, nil)

			assert.Equal(t, "", tc.have.GetSubject())
			if len(tc.err) == 0 {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, tc.have)
			} else {
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.err)
			}
		})
	}
}

// TestContext is a minimal implementation of Context for the purpose of testing.
type TestContext struct {
	context.Context

	MockIssuerURL *url.URL
	IssuerURLFunc func() (issuerURL *url.URL, err error)
	Clock         utils.Clock
}

// IssuerURL returns the MockIssuerURL.
func (m *TestContext) IssuerURL() (issuerURL *url.URL, err error) {
	if m.IssuerURLFunc != nil {
		return m.IssuerURLFunc()
	}

	return m.MockIssuerURL, nil
}

func (m *TestContext) GetClock() utils.Clock {
	if m.Clock != nil {
		return m.Clock
	}

	return &utils.RealClock{}
}

func (m *TestContext) GetJWTWithTimeFuncOption() jwt.ParserOption {
	return jwt.WithTimeFunc(m.GetClock().Now)
}

type TestCodeStrategy struct {
	signature string
}

func (m *TestCodeStrategy) AuthorizeCodeSignature(ctx context.Context, token string) string {
	return m.signature
}

func (m *TestCodeStrategy) GenerateAuthorizeCode(ctx context.Context, requester fosite.Requester) (token string, signature string, err error) {
	return "", "", nil
}

func (m *TestCodeStrategy) ValidateAuthorizeCode(ctx context.Context, requester fosite.Requester, token string) (err error) {
	return nil
}
