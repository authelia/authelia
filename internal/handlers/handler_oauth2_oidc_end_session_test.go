// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

const (
	testOIDCEndSessionEndpoint = "https://login.example.com:8080/api/oidc/end-session"

	testOIDCEndSessionClientID    = "end-session-client"
	testOIDCEndSessionPostLogout  = "https://app.example.com/logged-out"
	testOIDCEndSessionPostLogout2 = "https://app.example.com/goodbye"
)

func newTestOIDCEndSessionConfig(t *testing.T) *schema.IdentityProvidersOpenIDConnect {
	t.Helper()

	config := newTestOIDCConfig(t)

	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{
		{
			ID:                      testOIDCEndSessionClientID,
			Public:                  true,
			AuthorizationPolicy:     "one_factor",
			RedirectURIs:            []string{"https://app.example.com/callback"},
			PostLogoutRedirectURIs:  []string{testOIDCEndSessionPostLogout, testOIDCEndSessionPostLogout2},
			Scopes:                  []string{oidc.ScopeOpenID, oidc.ScopeProfile},
			ResponseTypes:           []string{oidc.ResponseTypeAuthorizationCodeFlow},
			GrantTypes:              []string{oidc.GrantTypeAuthorizationCode},
			TokenEndpointAuthMethod: oidc.ClientAuthMethodNone,
		},
	}

	return config
}

// mustGetTestOIDCEndSessionIDToken mints an ID Token suitable for use as an id_token_hint.
func mustGetTestOIDCEndSessionIDToken(t *testing.T, mock *mocks.MockAutheliaCtx, claims *jwt.IDTokenClaims) string {
	t.Helper()

	client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, testOIDCEndSessionClientID)
	require.NoError(t, err)

	tokenString, _, err := mock.Ctx.Providers.OpenIDConnect.Strategy.JWT.Encode(mock.Ctx, claims.ToMapClaims(), jwt.WithIDTokenClient(client))
	require.NoError(t, err)

	return tokenString
}

func newTestOIDCEndSessionIDTokenClaims() *jwt.IDTokenClaims {
	now := time.Now()

	return &jwt.IDTokenClaims{
		JTI:             "abc123",
		Issuer:          "https://login.example.com:8080",
		Subject:         "john",
		Audience:        []string{testOIDCEndSessionClientID},
		AuthorizedParty: testOIDCEndSessionClientID,
		IssuedAt:        jwt.NewNumericDate(now),
		ExpirationTime:  jwt.NewNumericDate(now.Add(time.Hour)),
		AccessTokenHash: "at-hash-abc123",
	}
}

func TestOpenIDConnectEndSession(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, nil)

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})

	t.Run("ShouldRedirectToLogoutWithoutParameters", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, nil)

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		assert.Equal(t, "true", location.Query().Get("confirm"))
		assert.Empty(t, location.Query().Get("rd"))

		assert.Equal(t, "no-store", rw.Header().Get(fasthttp.HeaderCacheControl))
		assert.Equal(t, "no-cache", rw.Header().Get(fasthttp.HeaderPragma))
	})

	t.Run("ShouldAlwaysRequireConfirmation", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		tokenString := mustGetTestOIDCEndSessionIDToken(t, mock, newTestOIDCEndSessionIDTokenClaims())

		// A valid id_token_hint permits skipping confirmation per the specification, but Authelia is deliberately
		// stricter and always requires the resource owner to confirm the logout.
		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint: []string{tokenString},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		assert.Equal(t, "true", location.Query().Get("confirm"))
	})

	t.Run("ShouldErrorOnUnregisteredClientID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID: []string{"not-a-client"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_client", location.Query().Get("error"))
	})

	t.Run("ShouldErrorOnPostLogoutRedirectURIWithoutClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterPostLogoutRedirectURI: []string{testOIDCEndSessionPostLogout},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
		assert.Contains(t, location.Query().Get("error_hint"), "requires either the 'client_id' or 'id_token_hint' parameter")
	})

	t.Run("ShouldErrorOnUnregisteredPostLogoutRedirectURI", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID:              []string{testOIDCEndSessionClientID},
			oidc.FormParameterPostLogoutRedirectURI: []string{"https://evil.example.com/logged-out"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
		assert.Contains(t, location.Query().Get("error_hint"), "does not match any registered post logout redirect URI")
	})

	t.Run("ShouldRedirectWithRegisteredPostLogoutRedirectURIAndState", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID:              []string{testOIDCEndSessionClientID},
			oidc.FormParameterPostLogoutRedirectURI: []string{testOIDCEndSessionPostLogout},
			oidc.FormParameterState:                 []string{"state-abc123"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		assert.Equal(t, "true", location.Query().Get("confirm"))
		assert.Equal(t, testOIDCEndSessionPostLogout, location.Query().Get("rd"))
		assert.Equal(t, "state-abc123", location.Query().Get("state"))
	})

	t.Run("ShouldNotIncludeStateWithoutPostLogoutRedirectURI", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID: []string{testOIDCEndSessionClientID},
			oidc.FormParameterState:    []string{"state-abc123"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		assert.Empty(t, location.Query().Get("state"))
	})

	t.Run("ShouldErrorOnMalformedIDTokenHint", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint: []string{"not-a-jwt"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
	})

	t.Run("ShouldAcceptExpiredIDTokenHint", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		claims := newTestOIDCEndSessionIDTokenClaims()
		claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-time.Hour * 2))
		claims.ExpirationTime = jwt.NewNumericDate(time.Now().Add(-time.Hour))

		tokenString := mustGetTestOIDCEndSessionIDToken(t, mock, claims)

		// The specification explicitly requires that an expired id_token_hint is still accepted.
		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint:           []string{tokenString},
			oidc.FormParameterPostLogoutRedirectURI: []string{testOIDCEndSessionPostLogout},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		assert.Equal(t, "true", location.Query().Get("confirm"))
		assert.Equal(t, testOIDCEndSessionPostLogout, location.Query().Get("rd"))
	})

	t.Run("ShouldErrorOnIDTokenHintWithMismatchedIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		claims := newTestOIDCEndSessionIDTokenClaims()
		claims.Issuer = "https://not-authelia.example.com"

		tokenString := mustGetTestOIDCEndSessionIDToken(t, mock, claims)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint: []string{tokenString},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
	})

	t.Run("ShouldErrorOnUnsupportedMethod", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodDelete, testOIDCEndSessionEndpoint, nil)

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
	})

	t.Run("ShouldErrorOnIDTokenHintWithoutSubject", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		claims := newTestOIDCEndSessionIDTokenClaims()
		claims.Subject = ""

		tokenString := mustGetTestOIDCEndSessionIDToken(t, mock, claims)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint: []string{tokenString},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
	})

	t.Run("ShouldResolveClientFromAuthorizedPartyWhenAudienceIsAmbiguous", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		// Multiple audiences are ambiguous, so the authorized party identifies the client.
		claims := newTestOIDCEndSessionIDTokenClaims()
		claims.Audience = []string{testOIDCEndSessionClientID, "another-client"}

		tokenString := mustGetTestOIDCEndSessionIDToken(t, mock, claims)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint:           []string{tokenString},
			oidc.FormParameterPostLogoutRedirectURI: []string{testOIDCEndSessionPostLogout},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		assert.Equal(t, testOIDCEndSessionPostLogout, location.Query().Get("rd"))
	})

	t.Run("ShouldErrorOnIDTokenHintWithAmbiguousAudienceAndNoAuthorizedParty", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		claims := newTestOIDCEndSessionIDTokenClaims()
		claims.Audience = []string{testOIDCEndSessionClientID, "another-client"}
		claims.AuthorizedParty = ""

		tokenString := mustGetTestOIDCEndSessionIDToken(t, mock, claims)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint: []string{tokenString},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
	})

	t.Run("ShouldNotRedirectErrorToUnvalidatedPostLogoutRedirectURI", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID:              []string{"not-a-client"},
			oidc.FormParameterPostLogoutRedirectURI: []string{"https://evil.example.com/logged-out"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.NotContains(t, location.String(), "evil.example.com")
	})

	t.Run("ShouldAcceptPostRequests", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID:              []string{testOIDCEndSessionClientID},
			oidc.FormParameterPostLogoutRedirectURI: []string{testOIDCEndSessionPostLogout2},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		assert.Equal(t, testOIDCEndSessionPostLogout2, location.Query().Get("rd"))
	})
}

func mustGetTestOIDCEndSessionLocation(t *testing.T, rw *httptest.ResponseRecorder) *url.URL {
	t.Helper()

	raw := rw.Header().Get(fasthttp.HeaderLocation)

	require.NotEmpty(t, raw)

	location, err := url.Parse(raw)
	require.NoError(t, err)

	return location
}
