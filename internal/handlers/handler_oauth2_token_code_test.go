// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestOAuth2TokenPOSTAuthorizationCode(t *testing.T) {
	t.Run("ShouldExchangeAuthorizationCode", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		records := setupTestOIDCAuthorizationCodeFlow(t, mock, client)

		code := mustGetTestOIDCAuthorizationCode(t, mock, newTestOIDCAuthorizationValues())

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{code},
			oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.NotEmpty(t, response["access_token"])
		assert.Equal(t, "bearer", response["token_type"])

		idToken, ok := response["id_token"].(string)

		require.True(t, ok)

		parts := strings.Split(idToken, ".")

		require.Len(t, parts, 3)

		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		require.NoError(t, err)

		claims := map[string]any{}

		require.NoError(t, json.Unmarshal(payload, &claims))
		require.Len(t, records, 1)

		for _, record := range records {
			assert.Equal(t, record.SessionID.String(), claims[oidc.ClaimSessionID])
		}
	})

	t.Run("ShouldNotExchangeAuthorizationCodeTwice", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		setupTestOIDCAuthorizationCodeFlow(t, mock, client)

		code := mustGetTestOIDCAuthorizationCode(t, mock, newTestOIDCAuthorizationValues())

		values := url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{code},
			oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
		}

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, values)

		OAuth2TokenPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		rw, r = newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, values)

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		assert.Equal(t, "invalid_grant", getTestOAuth2ErrorResponse(t, rw)["error"])
	})

	t.Run("ShouldHandleRedirectURIMismatch", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"
		client.RedirectURIs = append(client.RedirectURIs, "https://app.example.com/other")

		setupTestOIDCAuthorizationCodeFlow(t, mock, client)

		code := mustGetTestOIDCAuthorizationCode(t, mock, newTestOIDCAuthorizationValues())

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{code},
			oidc.FormParameterRedirectURI:     []string{"https://app.example.com/other"},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		assert.Equal(t, "invalid_grant", getTestOAuth2ErrorResponse(t, rw)["error"])
	})

	t.Run("ShouldHydrateJWTProfileAccessToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"
		client.AccessTokenSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
		client.AccessTokenSignedResponseKeyID = testOIDCKeyID

		setupTestOIDCAuthorizationCodeFlow(t, mock, client)

		values := newTestOIDCAuthorizationValues()
		values.Set(oidc.FormParameterScope, strings.Join([]string{oidc.ScopeOpenID, oidc.ScopeProfile}, " "))

		code := mustGetTestOIDCAuthorizationCode(t, mock, values)

		mock.SetLogLevel(logrus.DebugLevel)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{code},
			oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		assert.NotEmpty(t, getTestOAuth2ErrorResponse(t, rw)["access_token"])

		var hydrated bool

		for _, entry := range mock.Hook.AllEntries() {
			if entry.Message == "Access Request JWT Profile Claims Result" {
				hydrated = true

				assert.Contains(t, entry.Data, "extra")
			}
		}

		assert.True(t, hydrated, "the RFC9068 JWT Profile access token claims should have been hydrated")
	})
}

func TestOpenIDConnectUserinfoAuthorizationCode(t *testing.T) {
	t.Run("ShouldReturnClaims", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		setupTestOIDCAuthorizationCodeFlow(t, mock, client)

		code := mustGetTestOIDCAuthorizationCode(t, mock, newTestOIDCAuthorizationValues())

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{code},
			oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		token, ok := getTestOAuth2ErrorResponse(t, rw)["access_token"].(string)

		require.True(t, ok)

		rwu, ru := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)

		ru.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+token)

		OpenIDConnectUserinfo(mock.Ctx, rwu, ru)

		require.Equal(t, http.StatusOK, rwu.Code)

		claims := getTestOAuth2ErrorResponse(t, rwu)

		assert.NotEmpty(t, claims[oidc.ClaimSubject])
		assert.Equal(t, "no-store", rwu.Header().Get(fasthttp.HeaderCacheControl))
	})
}

func TestOAuth2TokenPOSTIssuerValidation(t *testing.T) {
	t.Run("ShouldRejectCodeRedeemedAtAnotherIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		setupTestOIDCAuthorizationCodeFlow(t, mock, client)

		code := mustGetTestOIDCAuthorizationCode(t, mock, newTestOIDCAuthorizationValues())

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example2.com")

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, "https://auth.example2.com/api/oidc/token", url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{code},
			oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		assert.Equal(t, "invalid_request", getTestOAuth2ErrorResponse(t, rw)["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAccessRequestIssuerMismatch, regexpAnyError)
	})
}

func TestOAuth2TokenPOSTErrors(t *testing.T) {
	t.Run("ShouldHandleHydrationFailure", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		client := newTestOIDCClientCredentialsClient(t)
		client.Scopes = []string{testOIDCScopeBearerAuthz, oidc.ScopeProfile}

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeClientCredentials},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{strings.Join([]string{testOIDCScopeBearerAuthz, oidc.ScopeProfile}, " ")},
			testOIDCFormParameterAudience:     []string{"https://app.example.com"},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		assert.Equal(t, "invalid_scope", getTestOAuth2ErrorResponse(t, rw)["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Access Request encountered an error while trying to populate the Client Credentials Flow requester", regexpAnyError)
	})

	t.Run("ShouldHandleAccessResponseGenerationError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCSessionIDStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		store := setupTestOIDCSessionStore(t, mock)

		code := mustGetTestOIDCAuthorizationCode(t, mock, newTestOIDCAuthorizationValues())

		store.FailSaves = true

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{code},
			oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAccessResponseCreationFailed, nil)
	})
}

func setupTestOIDCAuthorizationCodeFlow(t *testing.T, mock *mocks.MockAutheliaCtx, client schema.IdentityProvidersOpenIDConnectClient) (records map[string]model.OAuth2SessionID) {
	t.Helper()

	config := newTestOIDCConfig(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSessionStore(t, mock)
	setupTestOIDCConsentStore(t, mock)
	setupTestOIDCSubjectStore(t, mock)
	records = setupTestOIDCSessionIDStore(t, mock)
	setupTestOIDCUserDetails(t, mock)

	return records
}

func TestOAuth2TokenPOSTRefreshTokenUserDetails(t *testing.T) {
	testCases := []struct {
		name   string
		update func(details *authentication.UserDetails) error
		expect func(t *testing.T, rw *httptest.ResponseRecorder)
	}{
		{
			name: "ShouldRefreshIDTokenClaims",
			update: func(details *authentication.UserDetails) error {
				details.DisplayName = "Johnny Updated"

				return nil
			},
			expect: func(t *testing.T, rw *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rw.Code)

				assert.Equal(t, "Johnny Updated", mustGetTestOIDCIDTokenClaims(t, rw)[oidc.ClaimFullName])
			},
		},
		{
			name: "ShouldRejectRemovedUser",
			update: func(_ *authentication.UserDetails) error {
				return authentication.ErrUserNotFound
			},
			expect: func(t *testing.T, rw *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rw.Code)
				assert.Equal(t, "invalid_grant", getTestOAuth2ErrorResponse(t, rw)["error"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			client := newTestOIDCAuthorizationCodeClient(t)
			client.ConsentMode = "explicit"
			client.ClaimsPolicy = "name"
			client.GrantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken}
			client.Scopes = []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeOfflineAccess}

			config := newTestOIDCConfig(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}
			config.ClaimsPolicies["name"] = schema.IdentityProvidersOpenIDConnectClaimsPolicy{IDToken: []string{oidc.ClaimFullName}}

			setupTestOIDCProvider(t, mock, config)
			setupTestOIDCSessionStore(t, mock)
			setupTestOIDCSubjectStore(t, mock)
			setupTestOIDCSessionIDStore(t, mock)

			details := newTestOIDCUserDetails()

			var errDetails error

			mock.UserProviderMock.EXPECT().
				GetDetails(testUsername).
				AnyTimes().
				DoAndReturn(func(_ string) (*authentication.UserDetails, error) {
					return details, errDetails
				})

			mock.UserProviderMock.EXPECT().
				GetDetailsExtended(testUsername).
				AnyTimes().
				DoAndReturn(func(_ string) (*authentication.UserDetailsExtended, error) {
					if errDetails != nil {
						return nil, errDetails
					}

					return &authentication.UserDetailsExtended{UserDetails: details}, nil
				})

			scopes := []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeOfflineAccess}

			values := newTestOIDCAuthorizationValues()
			values.Set(oidc.FormParameterScope, strings.Join(scopes, " "))

			consent := newTestOIDCConsentSession(t, mock, mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID))
			consent.Form = values.Encode()
			consent.Authorized = true
			consent.RequestedScopes = scopes
			consent.GrantedScopes = scopes
			consent.RespondedAt = sql.NullTime{Time: mock.Ctx.GetClock().Now(), Valid: true}

			mock.StorageMock.EXPECT().
				LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
				AnyTimes().
				Return(consent, nil)

			mock.StorageMock.EXPECT().
				SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
				AnyTimes().
				Return(nil)

			mock.StorageMock.EXPECT().
				LoadOAuth2RefreshTokenSessionAccessSignature(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return("", nil)

			mock.StorageMock.EXPECT().
				UpdateOAuth2RefreshTokenSessionRequest(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(nil)

			mock.Ctx.Request.SetRequestURI(testOIDCAuthorizationEndpoint + "?consent_id=" + consent.ChallengeID.String())

			code := mustGetTestOIDCAuthorizationCode(t, mock, values)

			rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
				testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
				oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
				testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
				testOIDCFormParameterCode:         []string{code},
				oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
			})

			OAuth2TokenPOST(mock.Ctx, rw, r)

			require.Equal(t, http.StatusOK, rw.Code)

			assert.Equal(t, testDisplayName, mustGetTestOIDCIDTokenClaims(t, rw)[oidc.ClaimFullName])

			refresh, ok := getTestOAuth2ErrorResponse(t, rw)["refresh_token"].(string)
			require.True(t, ok)

			updated := *details
			errDetails = tc.update(&updated)
			details = &updated

			rw, r = newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
				testOIDCFormParameterGrantType:    []string{oidc.GrantTypeRefreshToken},
				oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
				testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
				oidc.GrantTypeRefreshToken:        []string{refresh},
			})

			OAuth2TokenPOST(mock.Ctx, rw, r)

			tc.expect(t, rw)
		})
	}
}

func mustGetTestOIDCIDTokenClaims(t *testing.T, rw *httptest.ResponseRecorder) (claims map[string]any) {
	t.Helper()

	idToken, ok := getTestOAuth2ErrorResponse(t, rw)["id_token"].(string)
	require.True(t, ok)

	parts := strings.Split(idToken, ".")
	require.Len(t, parts, 3)

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)

	claims = map[string]any{}

	require.NoError(t, json.Unmarshal(payload, &claims))

	return claims
}

func TestOAuth2TokenPOSTOnlineRefreshTokenOfEndedSession(t *testing.T) {
	testCases := []struct {
		name     string
		online   bool
		scopes   []string
		endCode  bool
		endGrant bool
		code     int
		refresh  bool
		mode     string
		expected int
		err      string
	}{
		{name: "ShouldNotIssueOnlineRefreshTokenWhenNotAllowed", scopes: []string{oidc.ScopeOpenID}, code: http.StatusOK},
		{name: "ShouldRefreshOnlineRefreshTokenOfLiveSession", online: true, scopes: []string{oidc.ScopeOpenID}, code: http.StatusOK, refresh: true, mode: "online", expected: http.StatusOK},
		{name: "ShouldRejectOnlineRefreshTokenOfEndedSession", online: true, scopes: []string{oidc.ScopeOpenID}, endGrant: true, code: http.StatusOK, refresh: true, mode: "online", expected: http.StatusBadRequest, err: "invalid_grant"},
		{name: "ShouldRefreshOfflineRefreshTokenOfEndedSession", online: true, scopes: []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}, endGrant: true, code: http.StatusOK, refresh: true, mode: "offline", expected: http.StatusOK},
		{name: "ShouldRejectAuthorizationCodeOfEndedSession", online: true, scopes: []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}, endCode: true, code: http.StatusBadRequest, err: "invalid_grant"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			client := newTestOIDCAuthorizationCodeClient(t)
			client.ConsentMode = "explicit"
			client.GrantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken}
			client.Scopes = []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}
			client.AllowOnlineRefreshTokens = tc.online

			config := newTestOIDCConfig(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

			setupTestOIDCProvider(t, mock, config)
			store := setupTestOIDCSessionStore(t, mock)
			setupTestOIDCSubjectStore(t, mock)
			setupTestOIDCSessionIDStore(t, mock)
			setupTestOIDCUserDetails(t, mock)

			values := newTestOIDCAuthorizationValues()
			values.Set(oidc.FormParameterScope, strings.Join(tc.scopes, " "))

			consent := newTestOIDCConsentSession(t, mock, mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID))
			consent.Form = values.Encode()
			consent.Authorized = true
			consent.RequestedScopes = tc.scopes
			consent.GrantedScopes = tc.scopes
			consent.RespondedAt = sql.NullTime{Time: mock.Ctx.GetClock().Now(), Valid: true}

			mock.StorageMock.EXPECT().
				LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
				AnyTimes().
				Return(consent, nil)

			mock.StorageMock.EXPECT().
				SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
				AnyTimes().
				Return(nil)

			mock.StorageMock.EXPECT().
				LoadOAuth2RefreshTokenSessionAccessSignature(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return("", nil)

			mock.StorageMock.EXPECT().
				UpdateOAuth2RefreshTokenSessionRequest(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(nil)

			mock.Ctx.Request.SetRequestURI(testOIDCAuthorizationEndpoint + "?consent_id=" + consent.ChallengeID.String())

			code := mustGetTestOIDCAuthorizationCode(t, mock, values)

			if tc.endCode {
				for _, session := range store.Sessions[storage.OAuth2SessionTypeAuthorizeCode] {
					require.True(t, session.SessionID.Valid)

					store.Ended[session.SessionID.String] = true
				}
			}

			rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
				testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
				oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
				testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
				testOIDCFormParameterCode:         []string{code},
				oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
			})

			OAuth2TokenPOST(mock.Ctx, rw, r)

			require.Equal(t, tc.code, rw.Code)

			response := getTestOAuth2ErrorResponse(t, rw)

			if tc.code != http.StatusOK {
				assert.Equal(t, tc.err, response["error"])

				return
			}

			refresh, ok := response["refresh_token"].(string)

			require.Equal(t, tc.refresh, ok)

			if !tc.refresh {
				return
			}

			for _, session := range store.Sessions[storage.OAuth2SessionTypeRefreshToken] {
				assert.Equal(t, tc.mode, session.RefreshTokenMode)
			}

			if tc.endGrant {
				for _, session := range store.Sessions[storage.OAuth2SessionTypeRefreshToken] {
					require.True(t, session.SessionID.Valid)

					store.Ended[session.SessionID.String] = true
				}
			}

			rw, r = newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
				testOIDCFormParameterGrantType:    []string{oidc.GrantTypeRefreshToken},
				oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
				testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
				oidc.GrantTypeRefreshToken:        []string{refresh},
			})

			OAuth2TokenPOST(mock.Ctx, rw, r)

			assert.Equal(t, tc.expected, rw.Code)

			if tc.err != "" {
				assert.Equal(t, tc.err, getTestOAuth2ErrorResponse(t, rw)["error"])

				return
			}

			require.Len(t, store.Sessions[storage.OAuth2SessionTypeRefreshToken], 2)

			for _, session := range store.Sessions[storage.OAuth2SessionTypeRefreshToken] {
				assert.Equal(t, tc.mode, session.RefreshTokenMode)
			}
		})
	}
}
