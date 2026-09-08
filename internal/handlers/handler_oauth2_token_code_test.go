package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2TokenPOSTAuthorizationCode(t *testing.T) {
	t.Run("ShouldExchangeAuthorizationCode", func(t *testing.T) {
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

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.NotEmpty(t, response["access_token"])
		assert.Equal(t, "bearer", response["token_type"])

		idToken, ok := response["id_token"].(string)

		require.True(t, ok)
		assert.Len(t, strings.Split(idToken, "."), 3)
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

func setupTestOIDCAuthorizationCodeFlow(t *testing.T, mock *mocks.MockAutheliaCtx, client schema.IdentityProvidersOpenIDConnectClient) {
	t.Helper()

	config := newTestOIDCConfig(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSessionStore(t, mock)
	setupTestOIDCConsentStore(t, mock)
	setupTestOIDCSubjectStore(t, mock)
	setupTestOIDCUserDetails(t, mock)
}
