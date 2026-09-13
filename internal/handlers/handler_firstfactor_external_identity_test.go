// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/identity"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestFirstFactorExternalIdentityProvidersGET(t *testing.T) {
	testCases := []struct {
		Name     string
		Setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Expected string
	}{
		{
			Name: "ShouldListConfiguredProviders",
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Providers.ExternalIdentity = newTestRelyingPartyProviders()
			},
			Expected: `{"status":"OK","data":{"providers":[{"id":"example","type":"openid_connect","name":"Example"}]}}`,
		},
		{
			Name: "ShouldListTheLogoURIWhenConfigured",
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Providers.ExternalIdentity = identity.NewProviders(&schema.AuthenticationBackendExternalIdentity{
					Providers: []schema.AuthenticationBackendExternalIdentityProvider{
						{ID: "discord", Type: "discord", Name: "Discord", ClientID: "abc", ClientSecret: "secret", LogoURI: "https://cdn.example.com/discord.png"},
					},
				}, nil)
			},
			Expected: `{"status":"OK","data":{"providers":[{"id":"discord","type":"discord","name":"Discord","logo_uri":"https://cdn.example.com/discord.png"}]}}`,
		},
		{
			Name:     "ShouldReturnEmptyListWhenNotConfigured",
			Setup:    nil,
			Expected: `{"status":"OK","data":{"providers":[]}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			FirstFactorExternalIdentityProvidersGET(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.Expected, string(mock.Ctx.Response.Body()))
		})
	}
}

func TestFirstFactorExternalIdentityPOST(t *testing.T) {
	testCases := []struct {
		Name     string
		Provider string
		Body     string
		Assert   func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			Name:     "ShouldStoreFlowStateAndReturnAuthorizationURL",
			Provider: "example",
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				body := struct {
					Data bodyPOSTExternalIdentityStartResponse `json:"data"`
				}{}

				require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &body))

				uri, err := url.Parse(body.Data.AuthorizationURL)
				require.NoError(t, err)

				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)
				require.NotNil(t, userSession.ExternalIdentity)

				assert.Equal(t, "example", userSession.ExternalIdentity.Provider)
				assert.Equal(t, userSession.ExternalIdentity.State, uri.Query().Get("state"))
				assert.Equal(t, userSession.ExternalIdentity.Nonce, uri.Query().Get("nonce"))
				assert.Equal(t, "S256", uri.Query().Get("code_challenge_method"))
				assert.NotEmpty(t, userSession.ExternalIdentity.CodeVerifier)
				assert.NotEqual(t, userSession.ExternalIdentity.CodeVerifier, uri.Query().Get("code_challenge"))
				assert.False(t, userSession.ExternalIdentity.Expires.IsZero())
				assert.Equal(t, "fr-CA", userSession.ExternalIdentity.Language)
			},
		},
		{
			Name:     "ShouldDiscardALanguageWhichIsNotALanguageTag",
			Provider: "example",
			Body:     `{"targetURL":"https://app.example.com","requestMethod":"GET","keepMeLoggedIn":false,"language":"en\r\nX-Injected: true"}`,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)
				require.NotNil(t, userSession.ExternalIdentity)

				assert.Empty(t, userSession.ExternalIdentity.Language)
			},
		},
		{
			Name:     "ShouldStoreTheFlowParameters",
			Provider: "example",
			Body:     `{"targetURL":"","requestMethod":"","keepMeLoggedIn":false,"language":"en","flow":"openid_connect","flowID":"abc","subflow":"device_authorization","userCode":"ABCD-EFGH"}`,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)
				require.NotNil(t, userSession.ExternalIdentity)

				assert.Equal(t, "openid_connect", userSession.ExternalIdentity.Flow)
				assert.Equal(t, "abc", userSession.ExternalIdentity.FlowID)
				assert.Equal(t, "device_authorization", userSession.ExternalIdentity.SubFlow)
				assert.Equal(t, "ABCD-EFGH", userSession.ExternalIdentity.UserCode)
			},
		},
		{
			Name:     "ShouldRejectAnUnknownFlow",
			Provider: "example",
			Body:     `{"keepMeLoggedIn":false,"flow":"not-a-flow","flowID":"abc"}`,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `{"status":"KO","message":"Could not start the external login."}`, string(mock.Ctx.Response.Body()))

				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)
				assert.Nil(t, userSession.ExternalIdentity)
			},
		},
		{
			Name:     "ShouldRejectAnUnknownSubflow",
			Provider: "example",
			Body:     `{"keepMeLoggedIn":false,"flow":"openid_connect","subflow":"not-a-subflow"}`,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `{"status":"KO","message":"Could not start the external login."}`, string(mock.Ctx.Response.Body()))
			},
		},
		{
			Name:     "ShouldRejectFlowParametersWithoutAFlow",
			Provider: "example",
			Body:     `{"keepMeLoggedIn":false,"flowID":"abc"}`,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `{"status":"KO","message":"Could not start the external login."}`, string(mock.Ctx.Response.Body()))
			},
		},
		{
			Name:     "ShouldRejectAnOverlongUserCode",
			Provider: "example",
			Body:     `{"keepMeLoggedIn":false,"flow":"openid_connect","subflow":"device_authorization","userCode":"ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJ"}`,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `{"status":"KO","message":"Could not start the external login."}`, string(mock.Ctx.Response.Body()))
			},
		},
		{
			Name:     "ShouldRejectUnknownProvider",
			Provider: "missing",
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `{"status":"KO","message":"Could not start the external login."}`, string(mock.Ctx.Response.Body()))

				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)
				assert.Nil(t, userSession.ExternalIdentity)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.ExternalIdentity = newTestRelyingPartyProviders()
			mock.Ctx.SetUserValue("provider", tc.Provider)
			body := tc.Body

			if body == "" {
				body = `{"targetURL":"https://app.example.com","requestMethod":"GET","keepMeLoggedIn":false,"language":"fr-CA"}`
			}

			mock.Ctx.Request.SetBodyString(body)

			FirstFactorExternalIdentityPOST(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

			if tc.Assert != nil {
				tc.Assert(t, mock)
			}
		})
	}
}

//nolint:gosec // Test Credentials.
func newTestRelyingPartyProviders() *identity.Providers {
	providers := identity.NewProviders(&schema.AuthenticationBackendExternalIdentity{
		Providers: []schema.AuthenticationBackendExternalIdentityProvider{
			{
				ID: "example", Name: "Example", Issuer: "https://op.example.com",
				ClientID: "client", ClientSecret: "secret",
				Scopes:                   []string{"openid", "email"},
				TokenEndpointAuthMethod:  "client_secret_basic",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
				Discovery:                schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
				Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{
					Authorization: "https://op.example.com/authorize",
					Token:         "https://op.example.com/token",
					JSONWebKeys:   "https://op.example.com/jwks.json",
				},
			},
		},
	}, nil)

	return providers
}
