package handlers

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2TokenPOST(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType: []string{oidc.GrantTypeClientCredentials},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAccessRequestIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleMissingClientAuthentication", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType: []string{oidc.GrantTypeClientCredentials},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusUnauthorized, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_client", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAccessRequestFailed, nil)
	})

	t.Run("ShouldHandleUnknownClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeClientCredentials},
			oidc.FormParameterClientID:        []string{"not-a-client"},
			testOIDCFormParameterClientSecret: []string{"not-a-secret"},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusUnauthorized, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_client", response["error"])
	})

	t.Run("ShouldGrantClientCredentialsToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCClientCredentialsClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.StorageMock.EXPECT().
			SaveOAuth2Session(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeClientCredentials},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{testOIDCScopeBearerAuthz},
			testOIDCFormParameterAudience:     []string{"https://app.example.com"},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.NotEmpty(t, response["access_token"])
		assert.Equal(t, "bearer", response["token_type"])
		assert.Equal(t, testOIDCScopeBearerAuthz, response["scope"])
		assert.NotContains(t, response, "error")
	})

	t.Run("ShouldHandleUnsupportedScopeForClientCredentials", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		client := newTestOIDCClientCredentialsClient(t)
		client.Scopes = []string{oidc.ScopeOpenID}
		client.Audience = nil

		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeClientCredentials},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{"profile"},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Contains(t, response, "error")
	})

	t.Run("ShouldHandleUnsupportedGrantTypeForClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		client := newTestOIDCClientCredentialsClient(t)
		client.Scopes = []string{oidc.ScopeOpenID}
		client.Audience = nil

		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{"abc"},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Contains(t, []any{"invalid_grant", "unauthorized_client", "invalid_client"}, response["error"])
	})
}

//nolint:gosec // This is an endpoint URL, not a credential.
const testOIDCTokenEndpoint = "https://login.example.com:8080/api/oidc/token"
