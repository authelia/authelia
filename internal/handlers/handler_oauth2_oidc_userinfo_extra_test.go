package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOpenIDConnectUserinfoExtra(t *testing.T) {
	t.Run("ShouldRejectRefreshToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "explicit"
		client.GrantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken}
		client.Scopes = []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeOfflineAccess}

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		subject := mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

		values := newTestOIDCAuthorizationValues()
		values.Set(oidc.FormParameterScope, strings.Join([]string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}, " "))

		consent := newTestOIDCConsentSession(t, mock, subject)
		consent.Form = values.Encode()
		consent.Authorized = true
		consent.RequestedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}
		consent.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}
		consent.RespondedAt = sql.NullTime{Time: mock.Ctx.GetClock().Now(), Valid: true}

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			AnyTimes().
			Return(nil)

		mock.Ctx.Request.SetRequestURI(testOIDCAuthorizationEndpoint + "?consent_id=" + consent.ChallengeID.String())

		rwa, ra := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationGET(mock.Ctx, rwa, ra)

		require.Equal(t, http.StatusSeeOther, rwa.Code)

		location, err := url.Parse(rwa.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)
		require.Empty(t, location.Query().Get("error"), location.Query().Get("error_description"))

		code := location.Query().Get(testOIDCFormParameterCode)

		require.NotEmpty(t, code)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{code},
			oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		refreshToken, ok := getTestOAuth2ErrorResponse(t, rw)["refresh_token"].(string)

		require.True(t, ok, "the offline_access scope should yield a refresh token")

		rwu, ru := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)

		ru.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+refreshToken)

		OpenIDConnectUserinfo(mock.Ctx, rwu, ru)

		assert.NotEqual(t, http.StatusOK, rwu.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpUserInfoNotAccessToken, nil)
	})

	t.Run("ShouldHandleClientNoLongerRegistered", func(t *testing.T) {
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

		configWithoutClient := newTestOIDCConfig(t)
		configWithoutClient.Clients = nil

		setupTestOIDCProvider(t, mock, configWithoutClient)

		rwu, ru := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)

		ru.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+token)

		OpenIDConnectUserinfo(mock.Ctx, rwu, ru)

		assert.NotEqual(t, http.StatusOK, rwu.Code)
	})

	t.Run("ShouldReturnSignedResponse", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"
		client.UserinfoSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
		client.UserinfoSignedResponseKeyID = testOIDCKeyID

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

		assert.Equal(t, middlewares.ContentTypeApplicationJWT, rwu.Header().Get(fasthttp.HeaderContentType))
		assert.Len(t, strings.Split(rwu.Body.String(), "."), 3)
	})

	t.Run("ShouldReturnClientCredentialsClaims", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		client := newTestOIDCClientCredentialsClient(t)
		client.Scopes = []string{oidc.ScopeOpenID}

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeClientCredentials},
			oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
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

		assert.Equal(t, testOIDCClientCredentialsID, claims[oidc.ClaimSubject])
	})
}

func TestOpenIDConnectUserinfoClaimsErrors(t *testing.T) {
	mustGetStandardFlowToken := func(t *testing.T, mock *mocks.MockAutheliaCtx) (token string) {
		t.Helper()

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

		return token
	}

	t.Run("ShouldFallBackToClientCredentialsClaimsWhenUserCannotBeLoaded", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		identifier := model.UserOpaqueIdentifier{Service: "openid", Username: testUsername, Identifier: uuid.Must(uuid.NewRandom())}

		mock.StorageMock.EXPECT().
			LoadUserOpaqueIdentifierBySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&identifier, nil)

		var broken bool

		mock.StorageMock.EXPECT().
			LoadUserOpaqueIdentifier(gomock.Any(), gomock.Any()).
			AnyTimes().
			DoAndReturn(func(_ any, _ uuid.UUID) (*model.UserOpaqueIdentifier, error) {
				if broken {
					return nil, fmt.Errorf("error in db")
				}

				value := identifier

				return &value, nil
			})

		token := mustGetStandardFlowToken(t, mock)

		broken = true

		rwu, ru := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)

		ru.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+token)

		OpenIDConnectUserinfo(mock.Ctx, rwu, ru)

		require.Equal(t, http.StatusOK, rwu.Code)

		claims := getTestOAuth2ErrorResponse(t, rwu)

		assert.NotEmpty(t, claims[oidc.ClaimSubject])

		var logged bool

		for _, entry := range mock.Hook.AllEntries() {
			if strings.Contains(entry.Message, "error occurred loading user information") {
				logged = true
			}
		}

		assert.True(t, logged, "the failure to load the user should be logged for a standard flow token")
	})
}

func TestOpenIDConnectUserinfoSigningError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	client := newTestOIDCClientCredentialsClient(t)
	client.Scopes = []string{oidc.ScopeOpenID}
	client.UserinfoSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256

	config := newTestOIDCConfigNoRSAKey(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSessionStore(t, mock)

	rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
		testOIDCFormParameterGrantType:    []string{oidc.GrantTypeClientCredentials},
		oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
		testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
	})

	OAuth2TokenPOST(mock.Ctx, rw, r)

	require.Equal(t, http.StatusOK, rw.Code)

	token, ok := getTestOAuth2ErrorResponse(t, rw)["access_token"].(string)

	require.True(t, ok)

	rwu, ru := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)

	ru.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+token)

	OpenIDConnectUserinfo(mock.Ctx, rwu, ru)

	assert.Equal(t, http.StatusInternalServerError, rwu.Code)

	assert.NotContains(t, rwu.Header().Get(fasthttp.HeaderContentType), "jwt")
	assert.Equal(t, "{}", strings.TrimSpace(rwu.Body.String()))
}
