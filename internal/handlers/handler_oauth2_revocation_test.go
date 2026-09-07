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

func TestOAuth2RevocationPOST(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCRevocationEndpoint, url.Values{
			testOIDCFormParameterToken: []string{"abc"},
		})

		OAuth2RevocationPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpRevocationRequestIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleMissingClientAuthentication", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCRevocationEndpoint, url.Values{
			testOIDCFormParameterToken: []string{"abc"},
		})

		OAuth2RevocationPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Contains(t, response, "error")

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpRevocationRequestFailed, nil)
	})

	t.Run("ShouldRevokeToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCClientCredentialsClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		token := mustGetTestOIDCClientCredentialsToken(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCRevocationEndpoint, url.Values{
			testOIDCFormParameterToken:        []string{token},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		})

		OAuth2RevocationPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		rwi, ri := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCIntrospectionEndpoint, url.Values{
			testOIDCFormParameterToken:        []string{token},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		})

		OAuth2IntrospectionPOST(mock.Ctx, rwi, ri)

		require.Equal(t, http.StatusOK, rwi.Code)

		assert.Equal(t, false, getTestOAuth2ErrorResponse(t, rwi)["active"])
	})

	t.Run("ShouldHandleUnknownToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCClientCredentialsClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCRevocationEndpoint, url.Values{
			testOIDCFormParameterToken:        []string{"authelia_at_not-a-real-token"},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		})

		OAuth2RevocationPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusOK, rw.Code)
	})
}

const testOIDCRevocationEndpoint = "https://login.example.com:8080/api/oidc/revocation"
