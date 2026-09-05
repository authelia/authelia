package handlers

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2AuthorizationGETPushedAuthorizationRequest(t *testing.T) {
	t.Run("ShouldAuthorizePushedRequest", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)
		setupTestOIDCPARStore(t, mock)

		requestURI := mustPushTestOIDCAuthorizationRequest(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:   []string{testOIDCAuthorizationCodeID},
			oidc.FormParameterRequestURI: []string{requestURI},
		})

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "app.example.com", location.Host)
		assert.NotEmpty(t, location.Query().Get(testOIDCFormParameterCode))
		assert.Empty(t, location.Query().Get("error"))
	})

	t.Run("ShouldRedirectAnonymousUserToConsentFlowAndRetainPushedRequest", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCPARStore(t, mock)

		requestURI := mustPushTestOIDCAuthorizationRequest(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:   []string{testOIDCAuthorizationCodeID},
			oidc.FormParameterRequestURI: []string{requestURI},
		})

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusFound, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "openid_connect", location.Query().Get("flow"))
		assert.NotEmpty(t, location.Query().Get("flow_id"))

		par, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2PushedAuthorizationSession(mock.Ctx, requestURI)

		require.NoError(t, err)
		assert.False(t, par.Revoked)
	})
}

func mustPushTestOIDCAuthorizationRequest(t *testing.T, mock *mocks.MockAutheliaCtx) (requestURI string) {
	t.Helper()

	rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCPAREndpoint, url.Values{
		oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
		testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		oidc.FormParameterResponseType:    []string{oidc.ResponseTypeAuthorizationCodeFlow},
		oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
		oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		oidc.FormParameterState:           []string{"abcdefghijklmnopqrstuvwxyz"},
	})

	OAuth2PushedAuthorizationRequest(mock.Ctx, rw, r)

	require.Equal(t, http.StatusCreated, rw.Code)

	requestURI, ok := getTestOAuth2ErrorResponse(t, rw)[oidc.FormParameterRequestURI].(string)

	require.True(t, ok)
	require.NotEmpty(t, requestURI)

	return requestURI
}
