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
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"authelia.com/provider/oauth2/token/jose"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
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
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCAuthorizationEndpoint, newTestOIDCAuthorizationValues())

		OAuth2AuthorizationPOST(mock.Ctx, rw, r)

		assert.Empty(t, rw.Header().Get(fasthttp.HeaderLocation))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationRequestIDIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldRedirectToGET", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		values := newTestOIDCAuthorizationValues()

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCAuthorizationEndpoint, values)

		OAuth2AuthorizationPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, "https", location.Scheme)
		assert.Equal(t, "login.example.com:8080", location.Host)
		assert.Equal(t, oidc.EndpointPathAuthorization, location.Path)

		query := location.Query()

		assert.Equal(t, testOIDCAuthorizationCodeID, query.Get(oidc.FormParameterClientID))
		assert.Equal(t, testOIDCRedirectURI, query.Get(oidc.FormParameterRedirectURI))
		assert.Equal(t, oidc.ScopeOpenID, query.Get(oidc.FormParameterScope))
	})

	t.Run("ShouldHandleMalformedMultipartForm", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCAuthorizationEndpoint, nil)

		r.Header.Set(fasthttp.HeaderContentType, "multipart/form-data; boundary=abc")
		r.Body = io.NopCloser(strings.NewReader("not a valid multipart body"))

		OAuth2AuthorizationPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusSeeOther, rw.Code)

		location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

		require.NoError(t, err)

		assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
		assert.NotEmpty(t, location.Query().Get("error"))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationMultipartError, regexpAnyError)
	})
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
