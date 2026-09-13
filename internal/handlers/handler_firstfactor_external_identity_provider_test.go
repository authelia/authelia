// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/externalidentity"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

// TestFirstFactorExternalIdentityCallbackProviderTypes drives the callback with a provider which is not an OpenID
// Connect 1.0 Provider, to assert the callback relies on nothing but the Provider interface.
func TestFirstFactorExternalIdentityCallbackProviderTypes(t *testing.T) {
	testCases := []struct {
		Name     string
		AMR      []string
		Complete func(request externalidentity.CompletionRequest) (*externalidentity.IdentityClaims, error)
		Setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Assert   func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Location string
	}{
		{
			Name: "ShouldAuthenticateLinkedIdentity",
			Complete: func(request externalidentity.CompletionRequest) (*externalidentity.IdentityClaims, error) {
				return &externalidentity.IdentityClaims{Issuer: "https://discord.com", Subject: "80351110224678912", PreferredUsername: "nelly", AuthenticationMethodsReference: []string{"pwd", "otp"}}, nil
			},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadExternalIdentityLinkBySubject(mock.Ctx, "stub", "https://discord.com", "80351110224678912").
					Return(&model.ExternalIdentityLink{ID: 7, Provider: "discord", Issuer: "https://discord.com", Subject: "80351110224678912", Username: "john"}, nil)

				mock.StorageMock.EXPECT().LoadBannedIP(mock.Ctx, model.NewIP(mock.Ctx.RemoteIP())).Return(nil, nil)
				mock.StorageMock.EXPECT().LoadBannedUser(mock.Ctx, "john").Return(nil, nil)

				mock.UserProviderMock.EXPECT().
					GetDetails("john").
					Return(&authentication.UserDetails{Username: "john", DisplayName: "John Smith", Emails: []string{"john@example.com"}}, nil)

				mock.StorageMock.EXPECT().UpdateExternalIdentityLinkSignIn(mock.Ctx, 7, gomock.Any()).Return(nil)
				mock.StorageMock.EXPECT().AppendAuthenticationLog(mock.Ctx, gomock.Any()).Return(nil)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Equal(t, "john", userSession.Username)
				assert.Equal(t, authentication.OneFactor, userSession.AuthenticationLevel(false))
				assert.Equal(t, authorization.AuthenticationMethodsReferences{FederatedIdentity: true}, userSession.AuthenticationMethodRefs,
					"a provider which does not trust the asserted amr values must not have them adopted")
			},
			Location: "https://app.example.com/",
		},
		{
			Name: "ShouldAdoptTheAuthenticationMethodsReferenceTheProviderResolves",
			AMR:  []string{"pwd", "otp"},
			Complete: func(request externalidentity.CompletionRequest) (*externalidentity.IdentityClaims, error) {
				return &externalidentity.IdentityClaims{Issuer: "https://discord.com", Subject: "80351110224678912"}, nil
			},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadExternalIdentityLinkBySubject(mock.Ctx, "stub", "https://discord.com", "80351110224678912").
					Return(&model.ExternalIdentityLink{ID: 7, Provider: "discord", Issuer: "https://discord.com", Subject: "80351110224678912", Username: "john"}, nil)

				mock.StorageMock.EXPECT().LoadBannedIP(mock.Ctx, model.NewIP(mock.Ctx.RemoteIP())).Return(nil, nil)
				mock.StorageMock.EXPECT().LoadBannedUser(mock.Ctx, "john").Return(nil, nil)

				mock.UserProviderMock.EXPECT().
					GetDetails("john").
					Return(&authentication.UserDetails{Username: "john", DisplayName: "John Smith", Emails: []string{"john@example.com"}}, nil)

				mock.StorageMock.EXPECT().UpdateExternalIdentityLinkSignIn(mock.Ctx, 7, gomock.Any()).Return(nil)
				mock.StorageMock.EXPECT().AppendAuthenticationLog(mock.Ctx, gomock.Any()).Return(nil)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Equal(t, authorization.AuthenticationMethodsReferences{FederatedIdentity: true, UsernameAndPassword: true, TOTP: true}, userSession.AuthenticationMethodRefs)
				assert.Equal(t, authentication.TwoFactor, userSession.AuthenticationLevel(false))
			},
			Location: "https://app.example.com/",
		},
		{
			Name: "ShouldHandBackUnlinkedIdentity",
			Complete: func(request externalidentity.CompletionRequest) (*externalidentity.IdentityClaims, error) {
				return &externalidentity.IdentityClaims{Issuer: "https://discord.com", Subject: "80351110224678912"}, nil
			},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadExternalIdentityLinkBySubject(mock.Ctx, "stub", "https://discord.com", "80351110224678912").
					Return(nil, storage.ErrNoExternalIdentityLink)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
			},
			Location: "https://login.example.com:8080/external-identity/link?link_provider=discord",
		},
		{
			Name: "ShouldRejectFailedCompletion",
			Complete: func(request externalidentity.CompletionRequest) (*externalidentity.IdentityClaims, error) {
				return nil, errors.New("error requesting the discord user: the discord user response is invalid")
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
			},
			Location: "https://login.example.com:8080/?external_identity_error=true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := newTestOpenIDConnectCallbackMock(t)

			defer mock.Close()

			provider := &stubExternalIdentityProvider{id: "discord", issuer: "https://discord.com", amr: tc.AMR, complete: tc.Complete}

			mock.Ctx.Providers.ExternalIdentity = externalidentity.NewProvidersWith(provider)
			mock.Ctx.SetUserValue("provider", "discord")
			mock.Ctx.Request.SetRequestURI("/api/firstfactor/external-identity/discord/callback?code=the-code&state=the-state")
			mock.Ctx.Request.SetHost("login.example.com:8080")

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.ExternalIdentity = &session.ExternalIdentityFlow{
				Provider:      "discord",
				State:         "the-state",
				CodeVerifier:  "the-verifier",
				Handle:        "the-handle",
				TargetURL:     "https://app.example.com",
				RequestMethod: fasthttp.MethodGet,
				Expires:       mock.Ctx.GetClock().Now().Add(time.Minute * 3),
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			FirstFactorExternalIdentityCallbackGET(mock.Ctx)

			assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.Location, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

			require.Len(t, provider.requests, 1)

			request := provider.requests[0]

			assert.WithinDuration(t, mock.Ctx.GetClock().Now(), request.Now, time.Second)

			request.Now = time.Time{}

			assert.Equal(t, externalidentity.CompletionRequest{
				Code:         "the-code",
				CodeVerifier: "the-verifier",
				Handle:       "the-handle",
				RedirectURI:  "https://login.example.com:8080/api/firstfactor/external-identity/discord/callback",
			}, request)

			if tc.Assert != nil {
				tc.Assert(t, mock)
			}
		})
	}
}

func TestFirstFactorExternalIdentityCallbackAuthorizationResponseIssuer(t *testing.T) {
	testCases := []struct {
		Name       string
		Required   bool
		RequireErr error
		Query      string
		Completed  bool
		Location   string
	}{
		{
			Name:     "ShouldRejectAbsentIssuerWhenRequired",
			Required: true,
			Query:    "code=the-code&state=the-state",
			Location: "https://login.example.com:8080/?external_identity_error=true",
		},
		{
			Name:     "ShouldRejectAbsentIssuerOnErrorResponseWhenRequired",
			Required: true,
			Query:    "error=access_denied&state=the-state",
			Location: "https://login.example.com:8080/?external_identity_error=true",
		},
		{
			Name:       "ShouldRejectWhenTheRequirementCannotBeDetermined",
			RequireErr: errors.New("error resolving provider 'discord': error discovering the provider: the discovery endpoint returned status code 404"),
			Query:      "code=the-code&state=the-state",
			Location:   "https://login.example.com:8080/?external_identity_error=true",
		},
		{
			Name:      "ShouldAcceptPresentIssuerWhenRequired",
			Required:  true,
			Query:     "code=the-code&state=the-state&iss=https%3A%2F%2Fdiscord.com",
			Completed: true,
			Location:  "https://login.example.com:8080/external-identity/link?link_provider=discord",
		},
		{
			Name:      "ShouldAcceptAbsentIssuerWhenNotRequired",
			Query:     "code=the-code&state=the-state",
			Completed: true,
			Location:  "https://login.example.com:8080/external-identity/link?link_provider=discord",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := newTestOpenIDConnectCallbackMock(t)

			defer mock.Close()

			provider := &stubExternalIdentityProvider{
				id: "discord", issuer: "https://discord.com",
				issRequired: tc.Required, issErr: tc.RequireErr,
				complete: func(request externalidentity.CompletionRequest) (*externalidentity.IdentityClaims, error) {
					return &externalidentity.IdentityClaims{Issuer: "https://discord.com", Subject: "80351110224678912"}, nil
				},
			}

			mock.Ctx.Providers.ExternalIdentity = externalidentity.NewProvidersWith(provider)
			mock.Ctx.SetUserValue("provider", "discord")
			mock.Ctx.Request.SetRequestURI("/api/firstfactor/external-identity/discord/callback?" + tc.Query)
			mock.Ctx.Request.SetHost("login.example.com:8080")

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.ExternalIdentity = &session.ExternalIdentityFlow{
				Provider:      "discord",
				State:         "the-state",
				CodeVerifier:  "the-verifier",
				TargetURL:     "https://app.example.com",
				RequestMethod: fasthttp.MethodGet,
				Expires:       mock.Ctx.GetClock().Now().Add(time.Minute * 3),
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			if tc.Completed {
				mock.StorageMock.EXPECT().
					LoadExternalIdentityLinkBySubject(mock.Ctx, "stub", "https://discord.com", "80351110224678912").
					Return(nil, storage.ErrNoExternalIdentityLink)
			}

			FirstFactorExternalIdentityCallbackGET(mock.Ctx)

			assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.Location, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

			if tc.Completed {
				assert.Len(t, provider.requests, 1)
			} else {
				assert.Empty(t, provider.requests, "the provider must not be contacted for a response which is rejected")
			}
		})
	}
}

type stubExternalIdentityProvider struct {
	id          string
	issuer      string
	amr         []string
	issRequired bool
	issErr      error
	complete    func(request externalidentity.CompletionRequest) (*externalidentity.IdentityClaims, error)
	requests    []externalidentity.CompletionRequest
}

func (p *stubExternalIdentityProvider) AuthorizationResponseIssuerRequired(_ context.Context) (bool, error) {
	return p.issRequired, p.issErr
}

func (p *stubExternalIdentityProvider) ID() string {
	return p.id
}

func (p *stubExternalIdentityProvider) Name() string {
	return "Stub"
}

func (p *stubExternalIdentityProvider) Type() string {
	return "stub"
}

func (p *stubExternalIdentityProvider) Issuer() string {
	return p.issuer
}

func (p *stubExternalIdentityProvider) ResponseMode() string {
	return externalidentity.ResponseModeQuery
}

func (p *stubExternalIdentityProvider) AuthenticationMethodsReference(_ []string) []string {
	return p.amr
}

func (p *stubExternalIdentityProvider) AuthorizationRequest(_ context.Context, _ random.Provider, _ externalidentity.AuthorizationRequestOptions) (*externalidentity.AuthorizationRequest, error) {
	return &externalidentity.AuthorizationRequest{URL: "https://discord.com/oauth2/authorize", State: "the-state", CodeVerifier: "the-verifier"}, nil
}

func (p *stubExternalIdentityProvider) Complete(_ context.Context, request externalidentity.CompletionRequest) (*externalidentity.IdentityClaims, error) {
	p.requests = append(p.requests, request)

	return p.complete(request)
}

func newTestExternalIdentityProvidersWithResponseMode(tokenEndpoint string, key *rsa.PrivateKey, mode string) *externalidentity.Providers {
	return externalidentity.NewProviders(&schema.AuthenticationBackendExternalIdentity{
		Providers: []schema.AuthenticationBackendExternalIdentityProvider{
			{
				ID: "example", Name: "Example", Issuer: "https://op.example.com",
				ClientID: "client", ClientSecret: "secret",
				Scopes:                   []string{"openid", "email"},
				ResponseMode:             mode,
				TokenEndpointAuthMethod:  "client_secret_basic",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
				Discovery:                schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
				Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{
					Authorization: "https://op.example.com/authorize",
					Token:         tokenEndpoint,
				},
				JSONWebKeys: []schema.JWK{
					{KeyID: "kid1", Use: "sig", Algorithm: "RS256", Key: key.Public()},
				},
			},
		},
	}, nil)
}
