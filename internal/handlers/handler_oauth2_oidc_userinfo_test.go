// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
		assert.Equal(t, "Bearer", rw.Header().Get(fasthttp.HeaderWWWAuthenticate))
		assert.Equal(t, "invalid_token", getTestOAuth2ErrorResponse(t, rw)["error"])

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
		assert.Contains(t, rw.Header().Get(fasthttp.HeaderWWWAuthenticate), `error="invalid_token"`)

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

		client := newTestOIDCClientCredentialsClient(t)
		client.Scopes = []string{oidc.ScopeOpenID}

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		rwt, rt := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeClientCredentials},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2TokenPOST(mock.Ctx, rwt, rt)

		require.Equal(t, http.StatusOK, rwt.Code)

		token, ok := getTestOAuth2ErrorResponse(t, rwt)["access_token"].(string)

		require.True(t, ok)

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example2.com")

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, "https://auth.example2.com/api/oidc/userinfo", nil)

		r.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+token)

		OpenIDConnectUserinfo(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpUserInfoIssuerMismatch, nil)
	})
}

func TestOpenIDConnectUserinfoAccessTokenURIQueryParameter(t *testing.T) {
	testCases := []struct {
		name     string
		enable   bool
		status   int
		expected string
	}{
		{
			"ShouldIgnoreTokenInQueryByDefault",
			false,
			http.StatusUnauthorized,
			"invalid_token",
		},
		{
			"ShouldReadTokenInQueryWhenEnabled",
			true,
			http.StatusForbidden,
			"insufficient_scope",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			config := newTestOIDCConfig(t)
			config.EnableAccessTokenURIQueryParameter = tc.enable
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCClientCredentialsClient(t)}

			setupTestOIDCProvider(t, mock, config)
			setupTestOIDCSessionStore(t, mock)

			token := mustGetTestOIDCClientCredentialsToken(t, mock)

			rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint+"?access_token="+url.QueryEscape(token), nil)

			OpenIDConnectUserinfo(mock.Ctx, rw, r)

			assert.Equal(t, tc.status, rw.Code)
			assert.Equal(t, tc.expected, getTestOAuth2ErrorResponse(t, rw)["error"])
		})
	}
}

const testOIDCUserinfoEndpoint = "https://login.example.com:8080/api/oidc/userinfo"
