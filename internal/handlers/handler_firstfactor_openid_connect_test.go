package handlers

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidcrp"
)

func TestFirstFactorOpenIDConnectProvidersGET(t *testing.T) {
	testCases := []struct {
		Name     string
		Setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Expected string
	}{
		{
			Name: "ShouldListConfiguredProviders",
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Providers.OpenIDConnectRelyingParty = newTestRelyingPartyProviders()
			},
			Expected: `{"status":"OK","data":{"providers":[{"id":"example","name":"Example"}]}}`,
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

			FirstFactorOpenIDConnectProvidersGET(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.Expected, string(mock.Ctx.Response.Body()))
		})
	}
}

func TestFirstFactorOpenIDConnectPOST(t *testing.T) {
	testCases := []struct {
		Name     string
		Provider string
		Assert   func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			Name:     "ShouldStoreFlowStateAndReturnAuthorizationURL",
			Provider: "example",
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				body := struct {
					Data bodyPOSTOpenIDConnectStartResponse `json:"data"`
				}{}

				require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &body))

				uri, err := url.Parse(body.Data.AuthorizationURL)
				require.NoError(t, err)

				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)
				require.NotNil(t, userSession.OpenIDConnect)

				assert.Equal(t, "example", userSession.OpenIDConnect.Provider)
				assert.Equal(t, userSession.OpenIDConnect.State, uri.Query().Get("state"))
				assert.Equal(t, userSession.OpenIDConnect.Nonce, uri.Query().Get("nonce"))
				assert.Equal(t, "S256", uri.Query().Get("code_challenge_method"))
				assert.NotEmpty(t, userSession.OpenIDConnect.CodeVerifier)
				assert.NotEqual(t, userSession.OpenIDConnect.CodeVerifier, uri.Query().Get("code_challenge"))
				assert.False(t, userSession.OpenIDConnect.Expires.IsZero())
			},
		},
		{
			Name:     "ShouldRejectUnknownProvider",
			Provider: "missing",
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `{"status":"KO","message":"Could not start the external login."}`, string(mock.Ctx.Response.Body()))

				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)
				assert.Nil(t, userSession.OpenIDConnect)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.OpenIDConnectRelyingParty = newTestRelyingPartyProviders()
			mock.Ctx.SetUserValue("provider", tc.Provider)
			mock.Ctx.Request.SetBodyString(`{"targetURL":"https://app.example.com","requestMethod":"GET","keepMeLoggedIn":false}`)

			FirstFactorOpenIDConnectPOST(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

			if tc.Assert != nil {
				tc.Assert(t, mock)
			}
		})
	}
}

//nolint:gosec // Test Credentials.
func newTestRelyingPartyProviders() *oidcrp.Providers {
	providers := oidcrp.NewProviders(&schema.AuthenticationBackendOpenIDConnect{
		Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
			{
				ID: "example", Name: "Example", Issuer: "https://op.example.com",
				ClientID: "client", ClientSecret: "secret",
				Scopes:                   []string{"openid", "email"},
				TokenEndpointAuthMethod:  "client_secret_basic",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
				Discovery:                schema.AuthenticationBackendOpenIDConnectProviderDiscovery{Disable: true},
				Endpoints: schema.AuthenticationBackendOpenIDConnectProviderEndpoints{
					Authorization: "https://op.example.com/authorize",
					Token:         "https://op.example.com/token",
					JSONWebKeys:   "https://op.example.com/jwks.json",
				},
			},
		},
	}, nil)

	return providers
}
