// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"authelia.com/provider/jose"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestOAuth2AuthorizationGET(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		require.NotPanics(t, func() {
			OAuth2AuthorizationGET(mock.Ctx, rw, r)
		})

		assert.Empty(t, rw.Header().Get(fasthttp.HeaderLocation))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleUnknownClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		values := newTestOIDCAuthorizationValues()
		values.Set(oidc.FormParameterClientID, "not-a-client")

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_client", location.Query().Get("error"))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationFailed, nil)
	})

	t.Run("ShouldHandlePromptNoneWhileAnonymous", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		values := newTestOIDCAuthorizationValues()
		values.Set(oidc.FormParameterPrompt, oidc.PromptNone)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "login_required", location.Query().Get("error"))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationPromptNoneAnonymous, nil)
	})

	t.Run("ShouldRedirectAnonymousUserToConsentFlow", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "login.example.com:8080", location.Host)
		assert.Equal(t, "openid_connect", location.Query().Get("flow"))
		assert.NotEmpty(t, location.Query().Get("flow_id"))
	})

	t.Run("ShouldAcceptAnUnsignedRequestObjectWithoutATypeHeader", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.RequestObjectSigningAlg = oidc.SigningAlgNone

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)

		values := newTestOIDCAuthorizationValues()
		values.Set("request", newTestUnsignedRequestObject(t, values))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Empty(t, location.Query().Get("error"), location.Query().Get("error_debug"))
		assert.Equal(t, "openid_connect", location.Query().Get("flow"))
		assert.NotEmpty(t, location.Query().Get("flow_id"))
	})

	t.Run("ShouldRequireTheTypeHeaderOfASignedRequestObject", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		client := newTestOIDCAuthorizationCodeClient(t)
		client.RequestObjectSigningAlg = oidc.SigningAlgRSAUsingSHA256
		client.JSONWebKeys = []schema.JWK{{KeyID: "client", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: &key.PublicKey}}

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		values := newTestOIDCAuthorizationValues()
		values.Set("request", newTestSignedRequestObject(t, key, "client", values))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "invalid_request_object", location.Query().Get("error"))
		assert.Contains(t, mock.Hook.LastEntry().Message, "expects request objects to be signed with the 'typ' header value")
	})

	t.Run("ShouldHandleInvalidRedirectURI", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		values := newTestOIDCAuthorizationValues()
		values.Set(oidc.FormParameterRedirectURI, "https://not-registered.example.com")

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.Equal(t, "invalid_request", location.Query().Get("error"))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationFailed, nil)
	})

	t.Run("ShouldAuthorizeAuthenticatedUserWithImplicitConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "app.example.com", location.Host)
		assert.Equal(t, "/oidc/callback", location.Path)
		assert.NotEmpty(t, location.Query().Get("code"))
		assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", location.Query().Get("state"))
		assert.Empty(t, location.Query().Get("error"))
	})

	t.Run("ShouldHandleUserDetailsError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		mock.UserProviderMock.EXPECT().
			GetDetailsExtended(testUsername).
			Return(nil, sql.ErrConnDone)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "server_error", location.Query().Get("error"))
	})

	t.Run("ShouldDestroySessionAndRedirectToFlowWhenUserNotFound", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)

		mock.UserProviderMock.EXPECT().
			GetDetailsExtended(testUsername).
			Return(nil, authentication.ErrUserNotFound)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Empty(t, location.Query().Get("error"))
		assert.Equal(t, "login.example.com:8080", location.Host)
		assert.Equal(t, "openid_connect", location.Query().Get("flow"))
		assert.NotEmpty(t, location.Query().Get("flow_id"))

		userSession, err := mock.Ctx.GetSession()

		require.NoError(t, err)

		assert.True(t, userSession.IsAnonymous())
		assert.Equal(t, authentication.NotAuthenticated, userSession.AuthenticationLevel(false))
	})

	t.Run("ShouldDestroySessionAndReturnLoginRequiredWhenUserNotFoundWithPromptNone", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.UserProviderMock.EXPECT().
			GetDetailsExtended(testUsername).
			Return(nil, authentication.ErrUserNotFound)

		values := newTestOIDCAuthorizationValues()
		values.Set(oidc.FormParameterPrompt, oidc.PromptNone)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "login_required", location.Query().Get("error"))

		userSession, err := mock.Ctx.GetSession()

		require.NoError(t, err)

		assert.True(t, userSession.IsAnonymous())
	})

	t.Run("ShouldReturnAccessDeniedWhenUserNotFoundAndSessionSaveFails", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		repository := setupTestFailingSessionRepository(t, mock)

		userSession := newTestOIDCUserSession(1)

		require.NoError(t, mock.Ctx.SaveSession(&userSession))

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.UserProviderMock.EXPECT().
			GetDetailsExtended(testUsername).
			Return(nil, authentication.ErrUserNotFound)

		repository.errDelete = errTestSessionBackend
		repository.errSave = []error{errTestSessionBackend}

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "access_denied", location.Query().Get("error"))

		var destroyed, saved bool

		for _, entry := range mock.Hook.AllEntries() {
			switch {
			case strings.HasSuffix(entry.Message, "had an error while destroying session"):
				destroyed = true

				assert.ErrorIs(t, entry.Data["error"].(error), errTestSessionBackend)
			case strings.HasSuffix(entry.Message, "had an error while saving updated session"):
				saved = true

				assert.ErrorIs(t, entry.Data["error"].(error), errTestSessionBackend)
			}
		}

		assert.True(t, destroyed)
		assert.True(t, saved)
	})

	t.Run("ShouldRedirectUserWithInsufficientAuthenticationLevelToFlow", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.AuthorizationPolicy = "two_factor"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "login.example.com:8080", location.Host)
		assert.Equal(t, "openid_connect", location.Query().Get("flow"))
		assert.NotEmpty(t, location.Query().Get("flow_id"))
	})

	t.Run("ShouldRedirectPublicClientWithoutAssuredIdentityToExplicitConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"
		client.Public = true
		client.Secret = nil
		client.TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
		client.ApplicationType = oidc.ApplicationTypeNative
		client.RedirectURIs = []string{"http://127.0.0.1/oidc/callback"}

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		values := newTestOIDCAuthorizationValues()
		values.Set(oidc.FormParameterRedirectURI, "http://127.0.0.1:43210/oidc/callback")
		values.Set("code_challenge", "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM")
		values.Set("code_challenge_method", oidc.PKCEChallengeMethodSHA256)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "login.example.com:8080", location.Host)
		assert.Equal(t, "openid_connect", location.Query().Get("flow"))
		assert.NotEmpty(t, location.Query().Get("flow_id"))
	})

	t.Run("ShouldAuthorizePublicClientWithAssuredIdentityWithImplicitConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "implicit"
		client.Public = true
		client.Secret = nil
		client.TokenEndpointAuthMethod = oidc.ClientAuthMethodNone

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2RequestReplay(gomock.Any(), gomock.Any()).
			Return(nil)

		values := newTestOIDCAuthorizationValues()
		values.Set("code_challenge", "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM")
		values.Set("code_challenge_method", oidc.PKCEChallengeMethodSHA256)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "app.example.com", location.Host)
		assert.Equal(t, "/oidc/callback", location.Path)
		assert.NotEmpty(t, location.Query().Get("code"))
		assert.Empty(t, location.Query().Get("error"))
	})

	t.Run("ShouldRedirectToExplicitConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "explicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "login.example.com:8080", location.Host)
		assert.Equal(t, "openid_connect", location.Query().Get("flow"))
		assert.NotEmpty(t, location.Query().Get("flow_id"))
	})

	t.Run("ShouldHandleMalformedConsentID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "explicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		mock.Ctx.Request.SetRequestURI(testOIDCAuthorizationEndpoint + "?consent_id=not-a-uuid")

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.NotEmpty(t, location.Query().Get("error"))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpConsentMalformedChallengeID, nil)
	})

	t.Run("ShouldCompleteExplicitConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "explicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		subject := mustGetTestOIDCSubject(t, mock, client.ID)

		consent := newTestOIDCConsentSession(t, mock, subject)
		consent.Authorized = true
		consent.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID}
		consent.RespondedAt = sql.NullTime{Time: mock.Ctx.GetClock().Now(), Valid: true}

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Return(nil)

		mock.Ctx.Request.SetRequestURI(testOIDCAuthorizationEndpoint + "?consent_id=" + consent.ChallengeID.String())

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "app.example.com", location.Host)
		assert.NotEmpty(t, location.Query().Get("code"))
		assert.Empty(t, location.Query().Get("error"))
	})

	t.Run("ShouldNotIssueCodeWhenConsentAlreadyGranted", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "explicit"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		subject := mustGetTestOIDCSubject(t, mock, client.ID)

		consent := newTestOIDCConsentSession(t, mock, subject)
		consent.Authorized = true
		consent.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID}
		consent.RespondedAt = sql.NullTime{Time: mock.Ctx.GetClock().Now(), Valid: true}

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Return(storage.ErrNoRowsAffected)

		mock.StorageMock.EXPECT().
			SaveOAuth2Session(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		mock.Ctx.Request.SetRequestURI(testOIDCAuthorizationEndpoint + "?consent_id=" + consent.ChallengeID.String())

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Empty(t, location.Query().Get("code"))
		assert.NotEmpty(t, location.Query().Get("error"))
	})

	t.Run("ShouldRedirectToConsentWhenNoPreConfigurationMatches", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCAuthorizationCodeClient(t)
		client.ConsentMode = "pre-configured"
		client.ConsentPreConfiguredDuration = &testOIDCPreConfiguredDuration

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentPreConfigurations(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&storage.ConsentPreConfigRows{}, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationGET(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "openid_connect", location.Query().Get("flow"))
		assert.NotEmpty(t, location.Query().Get("flow_id"))
	})
}

func TestOAuth2AuthorizationPOST(t *testing.T) {
	testCases := []struct {
		name    string
		values  url.Values
		body    string
		headers map[string]string
		code    int
		setup   func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect  func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder)
	}{
		{
			name:   "ShouldHandleBadIssuer",
			values: newTestOIDCAuthorizationValues(),
			code:   http.StatusOK,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				clearForwardedHeaders(mock)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				assert.Empty(t, rw.Header().Get(fasthttp.HeaderLocation))

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationIssuerError, "missing required X-Forwarded-Host header")
			},
		},
		{
			name:   "ShouldStoreRequestInBoundConsentSession",
			values: newTestOIDCAuthorizationValues(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setupTestOIDCConsentStore(t, mock)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				consent := assertTestOAuth2AuthorizationConsentReference(t, mock, rw)

				assert.Equal(t, mocks.AuthorizationBinding, consent.Binding)
				assert.Equal(t, mocks.AuthorizationBinding, assertTestOAuth2AuthorizationBindingCookie(t, mock))
			},
		},
		{
			name:    "ShouldStoreRequestInBoundConsentSessionWhenCrossSite",
			values:  newTestOIDCAuthorizationValues(),
			headers: newTestOAuth2AuthorizationBounceHeaders(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setupTestOIDCConsentStore(t, mock)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				consent := assertTestOAuth2AuthorizationConsentReference(t, mock, rw)

				assert.Equal(t, mocks.AuthorizationBinding, consent.Binding)
			},
		},
		{
			name:   "ShouldIssueBindingWhenAbsent",
			values: newTestOIDCAuthorizationValues(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.DelCookie(session.AuthorizationBindingCookieName("authelia_session"))

				setupTestOIDCConsentStore(t, mock)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				consent := assertTestOAuth2AuthorizationConsentReference(t, mock, rw)

				binding := assertTestOAuth2AuthorizationBindingCookie(t, mock)

				assert.NotEqual(t, mocks.AuthorizationBinding, binding)
				assert.Equal(t, binding, consent.Binding)
			},
		},
		{
			name:   "ShouldHandleUnknownClient",
			values: url.Values{oidc.FormParameterClientID: []string{"abc"}},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				location := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation))

				assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
				assert.Equal(t, "invalid_client", location.Query().Get("error"))
				assert.Nil(t, mock.Ctx.Response.Header.PeekCookie(session.AuthorizationBindingCookieName("authelia_session")))
			},
		},
		{
			name: "ShouldHandleMalformedForm",
			body: "client_id=%zz",
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				location := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation))

				assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
				assert.Equal(t, "invalid_request", location.Query().Get("error"))

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationFailed, nil)
			},
		},
		{
			name:   "ShouldHandleSaveError",
			values: newTestOIDCAuthorizationValues(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("bad"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				location := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation))

				assert.Equal(t, testOIDCRedirectURI, (&url.URL{Scheme: location.Scheme, Host: location.Host, Path: location.Path}).String())
				assert.Equal(t, "server_error", location.Query().Get("error"))

				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`could not be processed: error occurred saving consent: bad$`), nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			config := newTestOIDCConfig(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

			setupTestOIDCProvider(t, mock, config)

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCAuthorizationEndpoint, tc.values)

			if tc.body != "" {
				r.Header.Set(fasthttp.HeaderContentType, "application/x-www-form-urlencoded")
				r.Body = io.NopCloser(strings.NewReader(tc.body))
				r.ContentLength = int64(len(tc.body))
			}

			for name, value := range tc.headers {
				r.Header.Set(name, value)
			}

			OAuth2AuthorizationPOST(mock.Ctx, rw, r)

			if tc.code == 0 {
				assert.Equal(t, http.StatusSeeOther, rw.Code)
			} else {
				assert.Equal(t, tc.code, rw.Code)
			}

			if tc.expect != nil {
				tc.expect(t, mock, rw)
			}
		})
	}
}

func mustParseTestURI(t *testing.T, in string) *url.URL {
	t.Helper()

	uri, err := url.Parse(in)
	require.NoError(t, err)

	return uri
}

func assertTestOAuth2AuthorizationConsentReference(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) (consent *model.OAuth2ConsentSession) {
	t.Helper()

	location := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation))

	assert.Equal(t, "https", location.Scheme)
	assert.Equal(t, "login.example.com:8080", location.Host)
	assert.Equal(t, oidc.EndpointPathAuthorization, location.Path)

	query := location.Query()

	require.Len(t, query, 2)
	assert.Equal(t, testOIDCAuthorizationCodeID, query.Get(oidc.FormParameterClientID))

	challengeID, err := uuid.Parse(query.Get(queryArgConsentID))
	require.NoError(t, err)

	consent, err = mock.Ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(mock.Ctx, challengeID)
	require.NoError(t, err)

	assert.Equal(t, testOIDCAuthorizationCodeID, consent.ClientID)
	assert.False(t, consent.Subject.Valid)
	assert.False(t, consent.Responded())

	form, err := consent.GetForm()
	require.NoError(t, err)

	assert.Equal(t, newTestOIDCAuthorizationValues(), form)

	return consent
}

func assertTestOAuth2AuthorizationBindingCookie(t *testing.T, mock *mocks.MockAutheliaCtx) (value string) {
	t.Helper()

	name := session.AuthorizationBindingCookieName("authelia_session")

	raw := mock.Ctx.Response.Header.PeekCookie(name)
	require.NotNil(t, raw)

	cookie := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(cookie)

	require.NoError(t, cookie.ParseBytes(raw))

	assert.Equal(t, "__Host-Http-authelia_session_oauth2", string(cookie.Key()))
	assert.Len(t, cookie.Value(), 43)
	assert.Equal(t, "/", string(cookie.Path()))
	assert.Empty(t, cookie.Domain())
	assert.Equal(t, 3600, cookie.MaxAge())
	assert.True(t, cookie.Secure())
	assert.True(t, cookie.HTTPOnly())
	assert.Equal(t, fasthttp.CookieSameSiteLaxMode, cookie.SameSite())

	return string(cookie.Value())
}

const testOIDCAuthorizationEndpoint = "https://login.example.com:8080/api/oidc/authorization"

func newTestOIDCAuthorizationValues() url.Values {
	return url.Values{
		oidc.FormParameterClientID:     []string{testOIDCAuthorizationCodeID},
		oidc.FormParameterResponseType: []string{oidc.ResponseTypeAuthorizationCodeFlow},
		oidc.FormParameterRedirectURI:  []string{testOIDCRedirectURI},
		oidc.FormParameterScope:        []string{oidc.ScopeOpenID},
		oidc.FormParameterState:        []string{"abcdefghijklmnopqrstuvwxyz"},
	}
}

func newTestUnsignedRequestObject(t *testing.T, values url.Values) string {
	t.Helper()

	claims := map[string]any{}

	for key := range values {
		claims[key] = values.Get(key)
	}

	data, err := json.Marshal(claims)
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)) + "." + base64.RawURLEncoding.EncodeToString(data) + "."
}

func newTestSignedRequestObject(t *testing.T, key *rsa.PrivateKey, kid string, values url.Values) string {
	t.Helper()

	claims := map[string]any{}

	for k := range values {
		claims[k] = values.Get(k)
	}

	data, err := json.Marshal(claims)
	require.NoError(t, err)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: kid, Algorithm: string(jose.RS256), Use: oidc.KeyUseSignature}}, nil)
	require.NoError(t, err)

	signed, err := signer.Sign(data)
	require.NoError(t, err)

	object, err := signed.CompactSerialize()
	require.NoError(t, err)

	return object
}
