// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

const testOIDCTokenExchangeID = "token-exchange"

//nolint:gosec // These are form parameter names, not credentials.
const (
	testOIDCFormParameterSubjectToken       = "subject_token"
	testOIDCFormParameterSubjectTokenType   = "subject_token_type"
	testOIDCFormParameterRequestedTokenType = "requested_token_type"
	testOIDCFormParameterIssuedTokenType    = "issued_token_type"
)

// testOIDCTokenExchangeMutator adjusts the provider and client configurations of a Token Exchange test before the
// provider is created.
type testOIDCTokenExchangeMutator func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient)

// TestOAuth2TokenPOSTTokenExchange drives a real RFC 8693 Token Exchange through the provider: the subject token is
// obtained via the Authorization Code Flow as one client, then exchanged by a second client.
func TestOAuth2TokenPOSTTokenExchange(t *testing.T) {
	t.Run("ShouldAttributeExchangedIDTokenToRequestingClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCTokenExchangeFlow(t, mock)

		subject := mustGetTestOIDCSubjectAccessToken(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:          []string{oidc.GrantTypeTokenExchange},
			oidc.FormParameterClientID:              []string{testOIDCTokenExchangeID},
			testOIDCFormParameterClientSecret:       []string{testOIDCClientSecretValue},
			testOIDCFormParameterSubjectToken:       []string{subject},
			testOIDCFormParameterSubjectTokenType:   []string{oidc.TokenTypeAccessToken},
			testOIDCFormParameterRequestedTokenType: []string{oidc.TokenTypeIDToken},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		response := getTestOAuth2ErrorResponse(t, rw)

		require.Equal(t, http.StatusOK, rw.Code, response)

		assert.Equal(t, oidc.TokenTypeIDToken, response[testOIDCFormParameterIssuedTokenType])

		token, ok := response["access_token"].(string)

		require.True(t, ok)

		claims := mustDecodeTestJWTClaims(t, token)

		assert.Equal(t, testOIDCTokenExchangeID, claims[oidc.ClaimAuthorizedParty], "the 'azp' claim must identify the client which requested the exchange")
		assert.Contains(t, mustGetTestJWTAudience(t, claims), testOIDCTokenExchangeID, "the 'aud' claim must include the client which requested the exchange")
		assert.NotContains(t, mustGetTestJWTAudience(t, claims), testOIDCAuthorizationCodeID, "the 'aud' claim must not include the client the subject token was issued to")
		assert.NotEmpty(t, claims[oidc.ClaimSubject])
	})

	t.Run("ShouldAttributeExchangedJWTProfileAccessTokenToRequestingClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCTokenExchangeFlow(t, mock, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
			config.Discovery.JWTResponseAccessTokens = true

			exchange.AccessTokenSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
			exchange.AccessTokenSignedResponseKeyID = testOIDCKeyID
		})

		subject := mustGetTestOIDCSubjectAccessToken(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:          []string{oidc.GrantTypeTokenExchange},
			oidc.FormParameterClientID:              []string{testOIDCTokenExchangeID},
			testOIDCFormParameterClientSecret:       []string{testOIDCClientSecretValue},
			testOIDCFormParameterSubjectToken:       []string{subject},
			testOIDCFormParameterSubjectTokenType:   []string{oidc.TokenTypeAccessToken},
			testOIDCFormParameterRequestedTokenType: []string{oidc.TokenTypeAccessToken},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		response := getTestOAuth2ErrorResponse(t, rw)

		require.Equal(t, http.StatusOK, rw.Code, response)

		token, ok := response["access_token"].(string)

		require.True(t, ok)

		claims := mustDecodeTestJWTClaims(t, token)

		assert.Equal(t, testOIDCTokenExchangeID, claims[oidc.ClaimClientIdentifier], "the 'client_id' claim must identify the client which requested the exchange")
	})

	t.Run("ShouldExchangeAccessTokenForUsableAccessToken", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCTokenExchangeFlow(t, mock)

		subject := mustGetTestOIDCSubjectAccessToken(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:          []string{oidc.GrantTypeTokenExchange},
			oidc.FormParameterClientID:              []string{testOIDCTokenExchangeID},
			testOIDCFormParameterClientSecret:       []string{testOIDCClientSecretValue},
			testOIDCFormParameterSubjectToken:       []string{subject},
			testOIDCFormParameterSubjectTokenType:   []string{oidc.TokenTypeAccessToken},
			testOIDCFormParameterRequestedTokenType: []string{oidc.TokenTypeAccessToken},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		response := getTestOAuth2ErrorResponse(t, rw)

		require.Equal(t, http.StatusOK, rw.Code, response)

		assert.Equal(t, oidc.TokenTypeAccessToken, response[testOIDCFormParameterIssuedTokenType])
		assert.Equal(t, "bearer", response["token_type"])

		token, ok := response["access_token"].(string)

		require.True(t, ok)
		require.NotEmpty(t, token)

		assert.NotEqual(t, subject, token)

		rwi, ri := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCIntrospectionEndpoint, url.Values{
			testOIDCFormParameterToken:        []string{token},
			oidc.FormParameterClientID:        []string{testOIDCTokenExchangeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		})

		OAuth2IntrospectionPOST(mock.Ctx, rwi, ri)

		require.Equal(t, http.StatusOK, rwi.Code)

		introspection := getTestOAuth2ErrorResponse(t, rwi)

		assert.Equal(t, true, introspection["active"], introspection)
	})

	t.Run("ShouldDenyExchangeWhenSubjectClientDoesNotPermitTheRequestingClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCTokenExchangeFlow(t, mock, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
			subject.SubjectTokenClientsSupported = nil
		})

		token := mustGetTestOIDCSubjectAccessToken(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:          []string{oidc.GrantTypeTokenExchange},
			oidc.FormParameterClientID:              []string{testOIDCTokenExchangeID},
			testOIDCFormParameterClientSecret:       []string{testOIDCClientSecretValue},
			testOIDCFormParameterSubjectToken:       []string{token},
			testOIDCFormParameterSubjectTokenType:   []string{oidc.TokenTypeAccessToken},
			testOIDCFormParameterRequestedTokenType: []string{oidc.TokenTypeAccessToken},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		assert.Equal(t, "invalid_grant", getTestOAuth2ErrorResponse(t, rw)["error"])
	})
}

// setupTestOIDCTokenExchangeFlow registers the Authorization Code Flow client whose tokens are exchanged and the
// confidential client which performs the exchange, then wires up the storage mocks. Each mutator is applied to the
// provider and client configurations before the provider is created.
func setupTestOIDCTokenExchangeFlow(t *testing.T, mock *mocks.MockAutheliaCtx, mutators ...testOIDCTokenExchangeMutator) {
	t.Helper()

	subject := newTestOIDCAuthorizationCodeClient(t)
	subject.ConsentMode = "implicit"
	subject.SubjectTokenClientsSupported = []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
		{ClientID: testOIDCTokenExchangeID},
	}

	exchange := newTestOIDCTokenExchangeClient(t)

	config := newTestOIDCConfig(t)

	// The validator applies these defaults to every deployment, so apply them here too; the RFC 8693 token type
	// handlers take their lifespans from these values directly.
	config.Lifespans = schema.DefaultOpenIDConnectConfiguration.Lifespans

	for _, mutator := range mutators {
		mutator(config, &subject, &exchange)
	}

	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{subject, exchange}

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSessionStore(t, mock)
	setupTestOIDCConsentStore(t, mock)
	setupTestOIDCSubjectStore(t, mock)
	setupTestOIDCUserDetails(t, mock)
}

func newTestOIDCTokenExchangeClient(t *testing.T) schema.IdentityProvidersOpenIDConnectClient {
	t.Helper()

	// A Token Exchange client is confidential and authenticates at the token endpoint just as a Client Credentials
	// Flow client does, so the base configuration is shared.
	client := newTestOIDCClientCredentialsClient(t)

	client.ID = testOIDCTokenExchangeID
	client.GrantTypes = []string{oidc.GrantTypeTokenExchange}
	client.Scopes = []string{oidc.ScopeOpenID, oidc.ScopeProfile}
	client.Audience = nil

	return client
}

// mustGetTestOIDCSubjectAccessToken performs the Authorization Code Flow as the subject client and returns the issued
// access token, which is the 'subject_token' of the exchange.
func mustGetTestOIDCSubjectAccessToken(t *testing.T, mock *mocks.MockAutheliaCtx) (token string) {
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

	response := getTestOAuth2ErrorResponse(t, rw)

	require.Equal(t, http.StatusOK, rw.Code, response)

	token, ok := response["access_token"].(string)

	require.True(t, ok)
	require.NotEmpty(t, token)

	return token
}

func mustDecodeTestJWTClaims(t *testing.T, token string) (claims map[string]any) {
	t.Helper()

	parts := strings.Split(token, ".")

	require.Len(t, parts, 3, "the token should be a JWT")

	data, err := base64.RawURLEncoding.DecodeString(parts[1])

	require.NoError(t, err)

	claims = map[string]any{}

	require.NoError(t, json.Unmarshal(data, &claims))

	return claims
}

func mustGetTestJWTAudience(t *testing.T, claims map[string]any) (audience []string) {
	t.Helper()

	switch value := claims[oidc.ClaimAudience].(type) {
	case string:
		return []string{value}
	case []any:
		for _, item := range value {
			str, ok := item.(string)

			require.True(t, ok)

			audience = append(audience, str)
		}

		return audience
	case nil:
		return nil
	default:
		t.Fatalf("unexpected 'aud' claim type %T", value)

		return nil
	}
}
