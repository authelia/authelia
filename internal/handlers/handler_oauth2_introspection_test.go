package handlers

import (
	"database/sql"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2IntrospectionPOST(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCIntrospectionEndpoint, url.Values{
			testOIDCFormParameterToken: []string{"abc"},
		})

		OAuth2IntrospectionPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpIntrospectionRequestIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleMissingClientAuthentication", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCIntrospectionEndpoint, url.Values{
			testOIDCFormParameterToken: []string{"abc"},
		})

		OAuth2IntrospectionPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Contains(t, response, "error")

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpIntrospectionRequestFailed, regexpErrorInvalidClient)
	})

	t.Run("ShouldHandleInactiveToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCClientCredentialsClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.StorageMock.EXPECT().
			LoadOAuth2Session(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil, sql.ErrNoRows)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCIntrospectionEndpoint, url.Values{
			testOIDCFormParameterToken:        []string{"authelia_at_not-a-real-token"},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		})

		OAuth2IntrospectionPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, false, response["active"])

		assert.Nil(t, mock.Ctx.UserValue(middlewares.UserValueRateLimitExempt))
	})

	t.Run("ShouldHandleActiveToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCClientCredentialsClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		token := mustGetTestOIDCClientCredentialsToken(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCIntrospectionEndpoint, url.Values{
			testOIDCFormParameterToken:        []string{token},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		})

		OAuth2IntrospectionPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, true, response["active"])
		assert.Equal(t, testOIDCClientCredentialsID, response["client_id"])
		assert.Equal(t, testOIDCScopeBearerAuthz, response["scope"])
		assert.Equal(t, "https://login.example.com:8080", response["iss"])

		assert.Equal(t, true, mock.Ctx.UserValue(middlewares.UserValueRateLimitExempt))
	})

	t.Run("ShouldHandleTokenIssuedByAnotherIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCClientCredentialsClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		token := mustGetTestOIDCClientCredentialsToken(t, mock)

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example2.com")

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, "https://auth.example2.com/api/oidc/introspection", url.Values{
			testOIDCFormParameterToken:        []string{token},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		})

		OAuth2IntrospectionPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpIntrospectionRequestFailed, regexpIntrospectionIssuerMismatch)
	})
}

const testOIDCIntrospectionEndpoint = "https://login.example.com:8080/api/oidc/introspection"
