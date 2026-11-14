// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

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

		store := setupTestOIDCEndSession(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, nil)

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		store.assertFlowID(t, location)

		logout := store.mustGetLast(t)

		assert.Empty(t, logout.ClientID)
		assert.Empty(t, logout.RedirectURI)
		assert.Empty(t, logout.State)
		assert.Equal(t, mock.Ctx.GetClock().Now().Add(time.Minute*5).Unix(), logout.ExpiresAt.Unix())

		assert.Equal(t, "no-store", rw.Header().Get(fasthttp.HeaderCacheControl))
		assert.Equal(t, "no-cache", rw.Header().Get(fasthttp.HeaderPragma))
	})

	t.Run("ShouldAlwaysRequireConfirmation", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		store := setupTestOIDCEndSession(t, mock)

		tokenString := mustGetTestOIDCEndSessionIDToken(t, mock, newTestOIDCEndSessionIDTokenClaims())

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint: []string{tokenString},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		store.assertFlowID(t, location)

		assert.Equal(t, testOIDCEndSessionClientID, store.mustGetLast(t).ClientID)
	})

	t.Run("ShouldErrorOnUnregisteredClientID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCEndSession(t, mock)

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

		setupTestOIDCEndSession(t, mock)

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

		store := setupTestOIDCEndSession(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID:              []string{testOIDCEndSessionClientID},
			oidc.FormParameterPostLogoutRedirectURI: []string{"https://evil.example.com/logged-out"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
		assert.Equal(t, "400", location.Query().Get("error_status_code"))
		assert.Contains(t, location.Query().Get("error_hint"), "does not match any registered post logout redirect URI")

		store.assertNone(t)
	})

	t.Run("ShouldErrorOnRegisteredPostLogoutRedirectURIWithFragment", func(t *testing.T) {
		for _, uri := range []string{testOIDCEndSessionPostLogout + "#section", testOIDCEndSessionPostLogout + "#"} {
			t.Run(uri, func(t *testing.T) {
				mock := mocks.NewMockAutheliaCtx(t)
				defer mock.Close()

				store := setupTestOIDCEndSession(t, mock)

				rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
					oidc.FormParameterClientID:              []string{testOIDCEndSessionClientID},
					oidc.FormParameterPostLogoutRedirectURI: []string{uri},
				})

				OpenIDConnectEndSession(mock.Ctx, rw, r)

				assert.Equal(t, http.StatusFound, rw.Code)

				location := mustGetTestOIDCEndSessionLocation(t, rw)

				assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
				assert.Equal(t, "invalid_request", location.Query().Get("error"))
				store.assertNone(t)
			})
		}
	})

	t.Run("ShouldStoreRegisteredPostLogoutRedirectURIAndState", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		store := setupTestOIDCEndSession(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID:              []string{testOIDCEndSessionClientID},
			oidc.FormParameterPostLogoutRedirectURI: []string{testOIDCEndSessionPostLogout},
			oidc.FormParameterState:                 []string{"state-abc123"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		store.assertFlowID(t, location)

		logout := store.mustGetLast(t)

		assert.Equal(t, testOIDCEndSessionClientID, logout.ClientID)
		assert.Equal(t, testOIDCEndSessionPostLogout, logout.RedirectURI)
		assert.Equal(t, "state-abc123", logout.State)
	})

	t.Run("ShouldStoreEachRequestSeparately", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		store := setupTestOIDCEndSession(t, mock)

		flowIDs := make([]string, 0, 2)

		for _, redirect := range []string{testOIDCEndSessionPostLogout, testOIDCEndSessionPostLogout2} {
			rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
				oidc.FormParameterClientID:              []string{testOIDCEndSessionClientID},
				oidc.FormParameterPostLogoutRedirectURI: []string{redirect},
			})

			OpenIDConnectEndSession(mock.Ctx, rw, r)

			assert.Equal(t, http.StatusFound, rw.Code)

			flowIDs = append(flowIDs, mustGetTestOIDCEndSessionLocation(t, rw).Query().Get(oidc.FormParameterFlowID))
		}

		require.Len(t, store.saved, 2)

		assert.NotEqual(t, flowIDs[0], flowIDs[1])
		assert.Equal(t, flowIDs[0], store.saved[0].FlowID.String())
		assert.Equal(t, testOIDCEndSessionPostLogout, store.saved[0].RedirectURI)
		assert.Equal(t, flowIDs[1], store.saved[1].FlowID.String())
		assert.Equal(t, testOIDCEndSessionPostLogout2, store.saved[1].RedirectURI)
	})

	t.Run("ShouldNotStoreStateWithoutPostLogoutRedirectURI", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		store := setupTestOIDCEndSession(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID: []string{testOIDCEndSessionClientID},
			oidc.FormParameterState:    []string{"state-abc123"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		store.assertFlowID(t, location)

		logout := store.mustGetLast(t)

		assert.Empty(t, logout.RedirectURI)
		assert.Empty(t, logout.State)
	})

	t.Run("ShouldErrorOnMalformedIDTokenHint", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCEndSession(t, mock)

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

		store := setupTestOIDCEndSession(t, mock)

		claims := newTestOIDCEndSessionIDTokenClaims()
		claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-time.Hour * 2))
		claims.ExpirationTime = jwt.NewNumericDate(time.Now().Add(-time.Hour))

		tokenString := mustGetTestOIDCEndSessionIDToken(t, mock, claims)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint:           []string{tokenString},
			oidc.FormParameterPostLogoutRedirectURI: []string{testOIDCEndSessionPostLogout},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		assert.Equal(t, testOIDCEndSessionPostLogout, store.mustGetLast(t).RedirectURI)
	})

	t.Run("ShouldErrorOnIDTokenHintWithMismatchedIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCEndSession(t, mock)

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

		setupTestOIDCEndSession(t, mock)

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

		setupTestOIDCEndSession(t, mock)

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

		store := setupTestOIDCEndSession(t, mock)

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
		assert.Equal(t, testOIDCEndSessionPostLogout, store.mustGetLast(t).RedirectURI)
	})

	t.Run("ShouldErrorOnIDTokenHintWithAmbiguousAudienceAndNoAuthorizedParty", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCEndSession(t, mock)

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

		setupTestOIDCEndSession(t, mock)

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

	t.Run("ShouldStorePostRequests", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		store := setupTestOIDCEndSession(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterIDTokenHint:           []string{mustGetTestOIDCEndSessionIDToken(t, mock, newTestOIDCEndSessionIDTokenClaims())},
			oidc.FormParameterPostLogoutRedirectURI: []string{testOIDCEndSessionPostLogout2},
			oidc.FormParameterState:                 []string{"state-abc123"},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusSeeOther, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, "/logout", location.Path)
		store.assertFlowID(t, location)

		logout := store.mustGetLast(t)

		assert.Equal(t, testOIDCEndSessionClientID, logout.ClientID)
		assert.Equal(t, "john", logout.Subject)
		assert.Equal(t, testOIDCEndSessionPostLogout2, logout.RedirectURI)
		assert.Equal(t, "state-abc123", logout.State)
	})

	t.Run("ShouldNotUseTheSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCEndSession(t, mock)

		for _, method := range []string{fasthttp.MethodGet, fasthttp.MethodPost} {
			rw, r := newTestOAuth2Request(t, method, testOIDCEndSessionEndpoint, url.Values{
				oidc.FormParameterClientID: []string{testOIDCEndSessionClientID},
			})

			OpenIDConnectEndSession(mock.Ctx, rw, r)

			assert.Equal(t, "/logout", mustGetTestOIDCEndSessionLocation(t, rw).Path)
			assert.Empty(t, rw.Header().Values(fasthttp.HeaderSetCookie))
			assert.Empty(t, mock.Ctx.Response.Header.PeekCookie("authelia_session"))
		}
	})

	t.Run("ShouldErrorWhenTheRequestFailsToStore", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

		mock.StorageMock.EXPECT().SaveOAuth2LogoutSession(gomock.Any(), gomock.Any()).Return(errors.New("connection refused"))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCEndSessionEndpoint, url.Values{
			oidc.FormParameterClientID: []string{testOIDCEndSessionClientID},
		})

		OpenIDConnectEndSession(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusFound, rw.Code)

		location := mustGetTestOIDCEndSessionLocation(t, rw)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "server_error", location.Query().Get("error"))
	})
}

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

type testOIDCEndSessionStore struct {
	saved []model.OAuth2LogoutSession
}

func setupTestOIDCEndSession(t *testing.T, mock *mocks.MockAutheliaCtx) (store *testOIDCEndSessionStore) {
	t.Helper()

	setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

	store = &testOIDCEndSessionStore{}

	mock.StorageMock.EXPECT().
		SaveOAuth2LogoutSession(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, logout model.OAuth2LogoutSession) error {
			store.saved = append(store.saved, logout)

			return nil
		})

	return store
}

func (s *testOIDCEndSessionStore) mustGetLast(t *testing.T) model.OAuth2LogoutSession {
	t.Helper()

	require.NotEmpty(t, s.saved, "expected the logout request to be stored")

	return s.saved[len(s.saved)-1]
}

func (s *testOIDCEndSessionStore) assertNone(t *testing.T) {
	t.Helper()

	assert.Empty(t, s.saved)
}

func (s *testOIDCEndSessionStore) assertFlowID(t *testing.T, location *url.URL) {
	t.Helper()

	query := location.Query()

	require.Len(t, query, 1, "only the flow id may be handed to the front-end")

	flowID := query.Get(oidc.FormParameterFlowID)

	_, err := uuid.Parse(flowID)
	require.NoError(t, err)

	assert.Equal(t, flowID, s.mustGetLast(t).FlowID.String())
}

func mustGetTestOIDCEndSessionLocation(t *testing.T, rw *httptest.ResponseRecorder) *url.URL {
	t.Helper()

	raw := rw.Header().Get(fasthttp.HeaderLocation)

	require.NotEmpty(t, raw)

	location, err := url.Parse(raw)
	require.NoError(t, err)

	return location
}
