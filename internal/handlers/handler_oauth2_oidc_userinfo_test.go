package handlers

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestOpenIDConnectUserinfo(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)

		OpenIDConnectUserinfo(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpUserInfoIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleMissingToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)

		OpenIDConnectUserinfo(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusUnauthorized, rw.Code)
		assert.Contains(t, rw.Header().Get(fasthttp.HeaderWWWAuthenticate), `error="request_unauthorized"`)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpUserInfoFailed, nil)
	})

	t.Run("ShouldHandleInvalidToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)

		r.Header.Set(fasthttp.HeaderAuthorization, "Bearer authelia_at_not-a-real-token")

		OpenIDConnectUserinfo(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusUnauthorized, rw.Code)
		assert.Contains(t, rw.Header().Get(fasthttp.HeaderWWWAuthenticate), `error="request_unauthorized"`)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpUserInfoFailed, nil)
	})

	t.Run("ShouldHandleTokenWithoutOpenIDScope", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCClientCredentialsClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		token := mustGetTestOIDCClientCredentialsToken(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)

		r.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+token)

		OpenIDConnectUserinfo(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusForbidden, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "insufficient_scope", response["error"])
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

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, "https://auth.example2.com/api/oidc/userinfo", nil)

		r.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+token)

		OpenIDConnectUserinfo(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpUserInfoIssuerMismatch, nil)
	})
}

const testOIDCUserinfoEndpoint = "https://login.example.com:8080/api/oidc/userinfo"
