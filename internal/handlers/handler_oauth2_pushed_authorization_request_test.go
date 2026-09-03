package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2PushedAuthorizationRequest(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCPAREndpoint, url.Values{
			oidc.FormParameterClientID: []string{testOIDCAuthorizationCodeID},
		})

		OAuth2PushedAuthorizationRequest(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpPARIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleUnknownClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCPAREndpoint, url.Values{
			oidc.FormParameterClientID:     []string{"not-a-client"},
			oidc.FormParameterResponseType: []string{oidc.ResponseTypeAuthorizationCodeFlow},
			oidc.FormParameterRedirectURI:  []string{testOIDCRedirectURI},
			oidc.FormParameterScope:        []string{oidc.ScopeOpenID},
		})

		OAuth2PushedAuthorizationRequest(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusCreated, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_client", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpPARFailed, nil)
	})

	t.Run("ShouldHandleInvalidRedirectURI", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCPAREndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterResponseType:    []string{oidc.ResponseTypeAuthorizationCodeFlow},
			oidc.FormParameterRedirectURI:     []string{"https://not-registered.example.com"},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2PushedAuthorizationRequest(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusCreated, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])
	})

	t.Run("ShouldHandleResponseError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2PushedAuthorizationSession(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(fmt.Errorf("disk full"))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCPAREndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterResponseType:    []string{oidc.ResponseTypeAuthorizationCodeFlow},
			oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
			oidc.FormParameterState:           []string{"abcdefghijklmnopqrstuvwxyz"},
		})

		OAuth2PushedAuthorizationRequest(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusCreated, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpPARFailed, nil)
	})

	t.Run("ShouldPushRequest", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2PushedAuthorizationSession(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

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

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Contains(t, response, oidc.FormParameterRequestURI)
		assert.NotEmpty(t, response["expires_in"])
	})
}

const testOIDCPAREndpoint = "https://login.example.com:8080/api/oidc/pushed-authorization-request"

func newTestOIDCAuthorizationCodeClient(t *testing.T) schema.IdentityProvidersOpenIDConnectClient {
	t.Helper()

	return schema.IdentityProvidersOpenIDConnectClient{
		ID:                      testOIDCAuthorizationCodeID,
		Secret:                  mustDecodeTestSecret(t, testOIDCClientSecretDigest),
		AuthorizationPolicy:     "one_factor",
		GrantTypes:              []string{oidc.GrantTypeAuthorizationCode},
		ResponseTypes:           []string{oidc.ResponseTypeAuthorizationCodeFlow},
		Scopes:                  []string{oidc.ScopeOpenID, oidc.ScopeProfile},
		RedirectURIs:            []string{testOIDCRedirectURI},
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
	}
}
