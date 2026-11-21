// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"authelia.com/provider/jose"
	josejwt "authelia.com/provider/jose/jwt"
	oauthelia2 "authelia.com/provider/oauth2"
	fjwt "authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

const (
	testIDJAGIssuer       = "https://ras.example.com"
	testIDJAGAlias        = "ras"
	testIDJAGSubjectID    = "ras-sso"
	testIDJAGRASClientID  = "ras-client-id"
	testIDJAGClaimsPolicy = "ras"
	testIDJAGScopeRead    = "ras.read"
	testIDJAGScopeWrite   = "ras.write"
	testIDJAGResource     = "https://api.ras.example.com"
)

func setupTestOIDCIDJAGFlow(t *testing.T, mock *mocks.MockAutheliaCtx, mutators ...testOIDCTokenExchangeMutator) {
	t.Helper()

	setupTestOIDCTokenExchangeFlow(t, mock, append([]testOIDCTokenExchangeMutator{func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		config.ClaimsPolicies[testIDJAGClaimsPolicy] = schema.IdentityProvidersOpenIDConnectClaimsPolicy{
			IDToken: []string{oidc.ClaimEmail},
		}

		config.TokenExchange.TokenTypes = []string{oidc.TokenTypeAccessToken, oidc.TokenTypeRefreshToken, oidc.TokenTypeIDToken, oidc.TokenTypeIDJAG}

		config.IdentityAssertionAuthorizationGrant = schema.IdentityProvidersOpenIDConnectIDJAG{
			ResourceAuthorizationServers: []schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{
				{
					Issuer:          mustParseTestIDJAGURL(testIDJAGIssuer),
					Aliases:         []string{testIDJAGAlias},
					SubjectClientID: testIDJAGSubjectID,
					ClaimsPolicy:    testIDJAGClaimsPolicy,
				},
			},
		}

		subject.SectorIdentifierURI = mustParseTestIDJAGURL(testIDJAGIssuer)

		exchange.GrantTypes = append(exchange.GrantTypes, oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken)
		exchange.ResponseTypes = []string{oidc.ResponseTypeAuthorizationCodeFlow}
		exchange.RedirectURIs = []string{testOIDCRedirectURI}
		exchange.Scopes = append(exchange.Scopes, oidc.ScopeOfflineAccess, oidc.ScopeEmail)
		exchange.SubjectTokenTypesSupported = []string{oidc.TokenTypeIDToken, oidc.TokenTypeRefreshToken}
		exchange.RequestTokenTypesSupported = []string{oidc.TokenTypeIDJAG}
		exchange.IDJAGRelationships = []schema.IdentityProvidersOpenIDConnectClientIDJAGRelationship{
			{
				Issuer:   mustParseTestIDJAGURL(testIDJAGIssuer),
				ClientID: testIDJAGRASClientID,
				Scopes:   []string{testIDJAGScopeRead},
			},
		}
	}}, mutators...)...)
}

func setupTestOIDCIDJAGSubjectClientMutator(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
	subject.ID = testIDJAGSubjectID
}

func mustGetTestOIDCIDJAGOwnIDToken(t *testing.T, mock *mocks.MockAutheliaCtx) (token string) {
	t.Helper()

	token, ok := mustGetTestOIDCClientTokenResponseWithScope(t, mock, testOIDCTokenExchangeID, oidc.ScopeOpenID)["id_token"].(string)

	require.True(t, ok)
	require.NotEmpty(t, token)

	return token
}

func newTestOIDCIDJAGValues(subject, subjectType, audience string) url.Values {
	values := newTestOIDCTokenExchangeValues(subject, subjectType)

	values.Set(testOIDCFormParameterRequestedTokenType, oidc.TokenTypeIDJAG)
	values.Set(testOIDCFormParameterAudience, audience)
	values.Set(oidc.FormParameterScope, strings.Join([]string{testIDJAGScopeRead, testIDJAGScopeWrite}, " "))

	return values
}

func mustGetTestOIDCIDJAGClaims(t *testing.T, response map[string]any) (claims map[string]any) {
	t.Helper()

	assert.Equal(t, oidc.TokenTypeIDJAG, response[testOIDCFormParameterIssuedTokenType], response)
	assert.Equal(t, "N_A", response["token_type"], response)

	token, ok := response["access_token"].(string)

	require.True(t, ok, response)

	header := mustDecodeTestJWTHeader(t, token)

	assert.Equal(t, "oauth-id-jag+jwt", header["typ"])

	return mustDecodeTestJWTClaims(t, token)
}

func mustDecodeTestJWTHeader(t *testing.T, token string) (header map[string]any) {
	t.Helper()

	parts := strings.Split(token, ".")

	require.Len(t, parts, 3, "the token should be a JWT")

	data, err := base64.RawURLEncoding.DecodeString(parts[0])

	require.NoError(t, err)

	header = map[string]any{}

	require.NoError(t, json.Unmarshal(data, &header))

	return header
}

func assertTestOIDCIDJAGClaims(t *testing.T, mock *mocks.MockAutheliaCtx, claims map[string]any) {
	t.Helper()

	assert.Equal(t, testIDJAGIssuer, claims[oidc.ClaimAudience])
	assert.Equal(t, testIDJAGRASClientID, claims[oidc.ClaimClientIdentifier])
	assert.Equal(t, mustGetTestOIDCSubject(t, mock, testIDJAGSubjectID).String(), claims[oidc.ClaimSubject])
	assert.NotEqual(t, mustGetTestOIDCSubject(t, mock, testOIDCTokenExchangeID).String(), claims[oidc.ClaimSubject])
	assert.Equal(t, testIDJAGScopeRead, claims[oidc.ClaimScope])
	assert.Equal(t, testEmail, claims[oidc.ClaimEmail])
	assert.NotContains(t, claims, "cnf")
}

func TestOAuth2IDJAGIssueIDToken(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator)

	code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(mustGetTestOIDCIDJAGOwnIDToken(t, mock), oidc.TokenTypeIDToken, testIDJAGIssuer))

	require.Equal(t, http.StatusOK, code, response)

	assertTestOIDCIDJAGClaims(t, mock, mustGetTestOIDCIDJAGClaims(t, response))
}

func TestOAuth2IDJAGIssueRefreshToken(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator)

	code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(mustGetTestOIDCOwnRefreshToken(t, mock), oidc.TokenTypeRefreshToken, testIDJAGIssuer))

	require.Equal(t, http.StatusOK, code, response)

	assertTestOIDCIDJAGClaims(t, mock, mustGetTestOIDCIDJAGClaims(t, response))
}

func TestOAuth2IDJAGIssueAlias(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator)

	code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(mustGetTestOIDCIDJAGOwnIDToken(t, mock), oidc.TokenTypeIDToken, testIDJAGAlias))

	require.Equal(t, http.StatusOK, code, response)

	assert.Equal(t, testIDJAGIssuer, mustGetTestOIDCIDJAGClaims(t, response)[oidc.ClaimAudience])
}

func TestOAuth2IDJAGIssueResourceRequired(t *testing.T) {
	testCases := []struct {
		name     string
		mutator  testOIDCTokenExchangeMutator
		resource string
		err      string
	}{
		{
			"ShouldRequireResource",
			nil,
			"",
			"invalid_target",
		},
		{
			"ShouldGrantRequestedResource",
			nil,
			testIDJAGResource,
			"",
		},
		{
			"ShouldWaiveResourceWhenOptionalForProvider",
			func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
				config.IdentityAssertionAuthorizationGrant.ResourceOptional = true
			},
			"",
			"",
		},
		{
			"ShouldWaiveResourceWhenOptionalForClient",
			func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
				exchange.IDJAGResourceOptional = true
			},
			"",
			"",
		},
		{
			"ShouldRequireResourceWhenOptionalForOtherClient",
			func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
				subject.IDJAGResourceOptional = true
			},
			"",
			"invalid_target",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			mutators := []testOIDCTokenExchangeMutator{setupTestOIDCIDJAGSubjectClientMutator, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
				exchange.IDJAGRelationships[0].Resources = []string{testIDJAGResource}
			}}

			if tc.mutator != nil {
				mutators = append(mutators, tc.mutator)
			}

			setupTestOIDCIDJAGFlow(t, mock, mutators...)

			values := newTestOIDCIDJAGValues(mustGetTestOIDCIDJAGOwnIDToken(t, mock), oidc.TokenTypeIDToken, testIDJAGIssuer)

			if tc.resource != "" {
				values.Set(testIDJAGClaimResource, tc.resource)
			}

			code, response := doTestOIDCTokenExchange(t, mock, values)

			if tc.err != "" {
				assert.Equal(t, http.StatusBadRequest, code, response)
				assert.Equal(t, tc.err, response["error"], response)

				return
			}

			require.Equal(t, http.StatusOK, code, response)

			claims := mustGetTestOIDCIDJAGClaims(t, response)

			if tc.resource == "" {
				assert.NotContains(t, claims, testIDJAGClaimResource)
			} else {
				assert.Equal(t, tc.resource, claims[testIDJAGClaimResource])
			}
		})
	}
}

func TestOAuth2IDJAGIssueNoRelationship(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		config.IdentityAssertionAuthorizationGrant.ResourceAuthorizationServers = append(config.IdentityAssertionAuthorizationGrant.ResourceAuthorizationServers, schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{
			Issuer:          mustParseTestIDJAGURL("https://other.example.com"),
			SubjectClientID: testIDJAGSubjectID,
		})
	})

	code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(mustGetTestOIDCIDJAGOwnIDToken(t, mock), oidc.TokenTypeIDToken, "https://other.example.com"))

	assert.Equal(t, http.StatusBadRequest, code, response)
	assert.Equal(t, "access_denied", response["error"], response)
}

func TestOAuth2IDJAGIssueSubjectTokenOtherClient(t *testing.T) {
	testCases := []struct {
		name        string
		mutator     testOIDCTokenExchangeMutator
		token       func(t *testing.T, mock *mocks.MockAutheliaCtx) string
		tokenType   string
		description string
	}{
		{
			"ShouldRejectRefreshToken",
			func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
				subject.GrantTypes = append(subject.GrantTypes, oidc.GrantTypeRefreshToken)
				subject.Scopes = append(subject.Scopes, oidc.ScopeOfflineAccess)
				subject.ConsentMode = "explicit"
				subject.RequestedAudienceMode = "explicit"
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) string {
				return mustGetTestOIDCRefreshTokenForClient(t, mock, testOIDCAuthorizationCodeID)
			},
			oidc.TokenTypeRefreshToken,
			"The subject token was not issued to the OAuth 2.0 client requesting a '" + oidc.TokenTypeIDJAG + "'.",
		},
		{
			"ShouldRejectIDToken",
			func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) string {
				token, ok := mustGetTestOIDCClientTokenResponseWithScope(t, mock, testOIDCAuthorizationCodeID, oidc.ScopeOpenID)["id_token"].(string)

				require.True(t, ok)

				return token
			},
			oidc.TokenTypeIDToken,
			"Claim 'aud' from the id_token must include the OAuth 2.0 Client '" + testOIDCTokenExchangeID + "'.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			setupTestOIDCIDJAGFlow(t, mock, tc.mutator, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
				exchange.ImpersonationAllowed = true
				exchange.ConsentMode = "implicit"
			})

			code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(tc.token(t, mock), tc.tokenType, testIDJAGIssuer))

			assert.Equal(t, http.StatusBadRequest, code, response)
			assert.Equal(t, "invalid_request", response["error"], response)
			assert.Contains(t, response["error_description"], tc.description, response)
		})
	}
}

func TestOAuth2IDJAGIssueJWTProfileAccessTokenNotHydrated(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		config.ClaimsPolicies["access"] = schema.IdentityProvidersOpenIDConnectClaimsPolicy{AccessToken: []string{oidc.ClaimPreferredUsername}}
		config.Discovery.JWTResponseAccessTokens = true

		exchange.ClaimsPolicy = "access"
		exchange.AccessTokenSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
		exchange.AccessTokenSignedResponseKeyID = testOIDCKeyID
	})

	token := mustGetTestOIDCIDJAGOwnIDToken(t, mock)

	mock.SetLogLevel(logrus.DebugLevel)

	code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(token, oidc.TokenTypeIDToken, testIDJAGIssuer))

	require.Equal(t, http.StatusOK, code, response)

	assertTestOIDCIDJAGClaims(t, mock, mustGetTestOIDCIDJAGClaims(t, response))

	for _, entry := range mock.Hook.AllEntries() {
		assert.NotEqual(t, "Hydrate JWT Profile Access Token claims", entry.Message)
	}
}

func TestOAuth2IDJAGIssueRASPolicyDenied(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		config.AuthorizationPolicies = map[string]schema.IdentityProvidersOpenIDConnectPolicy{
			"deny-john": {
				DefaultPolicy: "one_factor",
				Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
					{Policy: "deny", Subjects: schema.AccessControlRuleSubjects{{"user:" + testUsername}}},
				},
			},
		}

		subject.AuthorizationPolicy = "deny-john"
	})

	code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(mustGetTestOIDCIDJAGOwnIDToken(t, mock), oidc.TokenTypeIDToken, testIDJAGIssuer))

	assert.Equal(t, http.StatusBadRequest, code, response)
	assert.Equal(t, "invalid_grant", response["error"], response)
	assert.Contains(t, response["error_description"], "The authorization policy of the requested audience does not permit this user.", response)
}

func TestOAuth2IDJAGIssueRASPolicyInsufficient(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		subject.AuthorizationPolicy = "two_factor"
	})

	code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(mustGetTestOIDCIDJAGOwnIDToken(t, mock), oidc.TokenTypeIDToken, testIDJAGIssuer))

	assert.Equal(t, http.StatusBadRequest, code, response)
	assert.Equal(t, "invalid_grant", response["error"], response)
	assert.Contains(t, response["error_description"], "The authentication level of the 'subject_token' is insufficient for the authorization policy of the requested audience.", response)
}

func TestOAuth2IDJAGIssueRequestingClientPolicyDenied(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator)

	token := mustGetTestOIDCIDJAGOwnIDToken(t, mock)

	client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, testOIDCTokenExchangeID)

	require.NoError(t, err)

	registered, ok := client.(*oidc.RegisteredClient)

	require.True(t, ok)

	registered.AuthorizationPolicy = oidc.ClientAuthorizationPolicy{DefaultPolicy: authorization.TwoFactor}

	code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(token, oidc.TokenTypeIDToken, testIDJAGIssuer))

	assert.Equal(t, http.StatusBadRequest, code, response)
	assert.Equal(t, "invalid_grant", response["error"], response)
	assert.Contains(t, response["error_description"], "The authentication level of the 'subject_token' is insufficient for the client's authorization policy.", response)
}

func TestOAuth2IDJAGIssueDPoPRefused(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator, setupTestOIDCIDJAGDPoPMutator(false))

	token := mustGetTestOIDCIDJAGOwnIDToken(t, mock)

	key, _ := newTestOIDCIDJAGDPoPKey(t)

	code, response := doTestOIDCIDJAGWithDPoP(t, mock, key, newTestOIDCIDJAGValues(token, oidc.TokenTypeIDToken, testIDJAGIssuer))

	assert.Equal(t, http.StatusBadRequest, code, response)
	assert.Equal(t, "invalid_request", response["error"], response)
	assert.Contains(t, response["error_description"], "The requested audience does not support DPoP bound grants.", response)
}

func TestOAuth2IDJAGIssueDPoPBound(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator, setupTestOIDCIDJAGDPoPMutator(true))

	token := mustGetTestOIDCIDJAGOwnIDToken(t, mock)

	key, jkt := newTestOIDCIDJAGDPoPKey(t)

	code, response := doTestOIDCIDJAGWithDPoP(t, mock, key, newTestOIDCIDJAGValues(token, oidc.TokenTypeIDToken, testIDJAGIssuer))

	require.Equal(t, http.StatusOK, code, response)

	claims := mustGetTestOIDCIDJAGClaims(t, response)

	assert.Equal(t, map[string]any{"jkt": jkt}, claims["cnf"])
}

func TestOAuth2IDJAGIssueAuditFields(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator)

	token := mustGetTestOIDCIDJAGOwnIDToken(t, mock)

	mock.Hook.Reset()

	code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGValues(token, oidc.TokenTypeIDToken, testIDJAGAlias))

	require.Equal(t, http.StatusOK, code, response)

	grant, ok := response["access_token"].(string)

	require.True(t, ok)

	entries := getTestOIDCTokenExchangeAuditEntries(mock)

	require.Len(t, entries, 1)

	entry := entries[0]

	assert.Equal(t, logrus.InfoLevel, entry.Level)
	assert.Equal(t, "Token Exchange completed", entry.Message)
	assert.Equal(t, oidc.TokenTypeIDJAG, entry.Data["issued_token_type"])
	assert.Equal(t, testIDJAGIssuer, entry.Data["id_jag_audience"])
	assert.Equal(t, testIDJAGRASClientID, entry.Data["id_jag_client_id"])

	for _, e := range mock.Hook.AllEntries() {
		assert.NotContains(t, e.Message, token)
		assert.NotContains(t, e.Message, grant)

		for _, v := range e.Data {
			assert.NotContains(t, fmt.Sprint(v), token)
			assert.NotContains(t, fmt.Sprint(v), grant)
		}
	}
}

func setupTestOIDCIDJAGDPoPMutator(supported bool) testOIDCTokenExchangeMutator {
	return func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		config.DPoP = schema.IdentityProvidersOpenIDConnectDPoP{
			Enabled:       true,
			ClockSkew:     time.Minute,
			NonceLifespan: time.Minute,
		}

		config.IdentityAssertionAuthorizationGrant.ResourceAuthorizationServers[0].JWTDPoPSupported = supported
	}
}

func newTestOIDCIDJAGDPoPKey(t *testing.T) (key *ecdsa.PrivateKey, jkt string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	jkt, err = fjwt.ThumbprintJWK(&jose.JSONWebKey{Key: key.Public()})
	require.NoError(t, err)

	return key, jkt
}

func doTestOIDCIDJAGWithDPoP(t *testing.T, mock *mocks.MockAutheliaCtx, key *ecdsa.PrivateKey, values url.Values) (code int, response map[string]any) {
	t.Helper()

	mock.StorageMock.EXPECT().
		CheckAndSetOAuth2DPoPProofUsed(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(false, nil)

	rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, values)

	r.Header.Set(oidc.HeaderDPoP, newTestOIDCIDJAGDPoPProof(t, key, fasthttp.MethodPost, testOIDCTokenEndpoint))

	OAuth2TokenPOST(mock.Ctx, rw, r)

	return rw.Code, getTestOAuth2ErrorResponse(t, rw)
}

func newTestOIDCIDJAGDPoPProof(t *testing.T, key *ecdsa.PrivateKey, method, target string) (proof string) {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: key},
		(&jose.SignerOptions{EmbedJWK: true}).WithType(fjwt.JSONWebTokenTypeDPoP),
	)
	require.NoError(t, err)

	payload, err := json.Marshal(map[string]any{
		"jti": uuid.Must(uuid.NewRandom()).String(),
		"htm": method,
		"htu": target,
		"iat": time.Now().Unix(),
	})
	require.NoError(t, err)

	object, err := signer.Sign(payload)
	require.NoError(t, err)

	proof, err = object.CompactSerialize()
	require.NoError(t, err)

	return proof
}

func mustParseTestIDJAGURL(raw string) *url.URL {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		panic(err)
	}

	return parsed
}

func TestHandleOAuth2IDJAGIssuePolicy(t *testing.T) {
	testCases := []struct {
		name     string
		audience string
		subject  bool
		details  bool
		expected string
	}{
		{"ShouldRejectNilDetails", testIDJAGIssuer, true, false, "invalid_grant"},
		{"ShouldErrorUnresolvableResourceAuthorizationServer", "https://unknown.example.com", true, true, "server_error"},
		{"ShouldErrorUnregisteredSubjectClient", testIDJAGIssuer, false, true, "server_error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			mutators := []testOIDCTokenExchangeMutator{}

			if tc.subject {
				mutators = append(mutators, setupTestOIDCIDJAGSubjectClientMutator)
			}

			setupTestOIDCIDJAGFlow(t, mock, mutators...)

			requester := &oauthelia2.AccessRequest{Request: oauthelia2.Request{GrantedAudience: oauthelia2.Arguments{tc.audience}}}

			var details *authentication.UserDetailsExtended

			if tc.details {
				details = &authentication.UserDetailsExtended{UserDetails: newTestOIDCUserDetails()}
			}

			err := handleOAuth2IDJAGIssuePolicy(mock.Ctx, requester, oidc.NewSession(), details)

			require.Error(t, err)
			assert.Equal(t, tc.expected, oauthelia2.ErrorToRFC6749Error(err).ErrorField)
		})
	}
}

func TestOAuth2IDJAGIssueDPoPDisabledProofIgnored(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGFlow(t, mock, setupTestOIDCIDJAGSubjectClientMutator)

	token := mustGetTestOIDCIDJAGOwnIDToken(t, mock)

	key, _ := newTestOIDCIDJAGDPoPKey(t)

	rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, newTestOIDCIDJAGValues(token, oidc.TokenTypeIDToken, testIDJAGIssuer))

	r.Header.Set(oidc.HeaderDPoP, newTestOIDCIDJAGDPoPProof(t, key, fasthttp.MethodPost, testOIDCTokenEndpoint))

	OAuth2TokenPOST(mock.Ctx, rw, r)

	response := getTestOAuth2ErrorResponse(t, rw)

	require.Equal(t, http.StatusOK, rw.Code, response)

	assert.NotContains(t, mustGetTestOIDCIDJAGClaims(t, response), "cnf")
}

const (
	testIDJAGTrustedIssuer    = "https://idp.example.com"
	testIDJAGTrustedKeyID     = "idp-key"
	testIDJAGExternalSubject  = "external-subject"
	testIDJAGUsernameClaim    = "preferred_username"
	testIDJAGGovernedResource = "https://app.example.com/"
	testIDJAGCELTwoFactor     = "has(claims.amr) && 'mfa' in claims.amr ? ['pwd', 'otp'] : ['pwd']"
	testIDJAGCELNoHas         = "'mfa' in claims.amr ? ['pwd', 'otp'] : ['pwd']"
	testIDJAGHintMapping      = "The authentication methods of the Identity Assertion JWT Authorization Grant could not be mapped."
	testIDJAGHintInsufficient = "The authentication level of the Identity Assertion JWT Authorization Grant is insufficient for the client's authorization policy."
	testIDJAGHintExplicit     = "The client requires explicit user consent which cannot be obtained when redeeming an Identity Assertion JWT Authorization Grant."
	testIDJAGHintNoConsent    = "The user has not consented to this client for the requested scopes."
	testIDJAGAssertion        = "assertion"
	testIDJAGClaimResource    = "resource"
)

var testIDJAGTrustedKey *rsa.PrivateKey

func mustGetTestIDJAGTrustedKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	if testIDJAGTrustedKey == nil {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		testIDJAGTrustedKey = key
	}

	return testIDJAGTrustedKey
}

func setupTestOIDCIDJAGRedeemFlow(t *testing.T, mock *mocks.MockAutheliaCtx, mutators ...testOIDCTokenExchangeMutator) {
	t.Helper()

	key := mustGetTestIDJAGTrustedKey(t)

	setupTestOIDCTokenExchangeFlow(t, mock, append([]testOIDCTokenExchangeMutator{func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		config.IdentityAssertionAuthorizationGrant = schema.IdentityProvidersOpenIDConnectIDJAG{
			TrustedIssuers: []schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer{
				{
					Issuer: mustParseTestIDJAGURL(testIDJAGTrustedIssuer),
					JSONWebKeys: []schema.JWK{
						{
							KeyID:     testIDJAGTrustedKeyID,
							Use:       oidc.KeyUseSignature,
							Algorithm: oidc.SigningAlgRSAUsingSHA256,
							Key:       &key.PublicKey,
						},
					},
					Clients:       []string{testOIDCTokenExchangeID},
					Resources:     []string{testIDJAGGovernedResource},
					UsernameClaim: testIDJAGUsernameClaim,
				},
			},
		}

		exchange.GrantTypes = append(exchange.GrantTypes, oidc.GrantTypeJWTBearer)
		exchange.Scopes = append(exchange.Scopes, testOIDCScopeBearerAuthz)
		exchange.Resource = []string{testIDJAGGovernedResource}
	}}, mutators...)...)

	setupTestOIDCIDJAGUsedStore(t, mock)
}

func setupTestOIDCIDJAGUsedStore(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	used := map[string]model.OAuth2BlacklistedIssuerJTI{}

	mock.StorageMock.EXPECT().
		LoadOAuth2BlacklistedIssuerJTI(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, issuer, signature string) (*model.OAuth2BlacklistedIssuerJTI, error) {
			if value, ok := used[issuer+":"+signature]; ok {
				return &value, nil
			}

			return nil, sql.ErrNoRows
		})

	mock.StorageMock.EXPECT().
		SaveOAuth2BlacklistedIssuerJTI(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, value model.OAuth2BlacklistedIssuerJTI) error {
			key := value.Issuer + ":" + value.Signature

			if _, ok := used[key]; ok {
				return storage.ErrNoRowsAffected
			}

			used[key] = value

			return nil
		})
}

func setupTestOIDCIDJAGRedeemAMRMutator(expression string) testOIDCTokenExchangeMutator {
	return func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		config.IdentityAssertionAuthorizationGrant.TrustedIssuers[0].AuthenticationMethodsReferences = expression
	}
}

func setupTestOIDCIDJAGRedeemTwoFactorMutator(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
	exchange.AuthorizationPolicy = "two_factor"
}

func setupTestOIDCIDJAGRedeemDPoPMutator(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
	config.DPoP = schema.IdentityProvidersOpenIDConnectDPoP{
		Enabled:       true,
		ClockSkew:     time.Minute,
		NonceLifespan: time.Minute,
	}

	exchange.GrantTypes = append(exchange.GrantTypes, oidc.GrantTypeJWTDPoP)
}

func newTestOIDCIDJAGRedeemClaims(t *testing.T, mock *mocks.MockAutheliaCtx) (claims map[string]any) {
	t.Helper()

	issuer, err := mock.Ctx.IssuerURL()
	require.NoError(t, err)

	now := time.Now()

	return map[string]any{
		oidc.ClaimIssuer:           testIDJAGTrustedIssuer,
		oidc.ClaimSubject:          testIDJAGExternalSubject,
		oidc.ClaimAudience:         issuer.String(),
		oidc.ClaimClientIdentifier: testOIDCTokenExchangeID,
		oidc.ClaimJWTID:            uuid.Must(uuid.NewRandom()).String(),
		oidc.ClaimExpirationTime:   now.Add(time.Minute * 5).Unix(),
		oidc.ClaimIssuedAt:         now.Unix(),
		testIDJAGUsernameClaim:     testUsername,
		oidc.ClaimScope:            strings.Join([]string{testOIDCScopeBearerAuthz, oidc.ScopeProfile}, " "),
		testIDJAGClaimResource:     testIDJAGGovernedResource,
	}
}

func mustSignTestOIDCIDJAGRedeemGrant(t *testing.T, claims map[string]any) (grant string) {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: mustGetTestIDJAGTrustedKey(t), KeyID: testIDJAGTrustedKeyID}},
		(&jose.SignerOptions{}).WithType("oauth-id-jag+jwt"),
	)
	require.NoError(t, err)

	grant, err = josejwt.Signed(signer).Claims(claims).Serialize()
	require.NoError(t, err)

	return grant
}

func newTestOIDCIDJAGRedeemValues(grantType, grant string) url.Values {
	return url.Values{
		testOIDCFormParameterGrantType:    []string{grantType},
		oidc.FormParameterClientID:        []string{testOIDCTokenExchangeID},
		testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		testIDJAGAssertion:                []string{grant},
	}
}

func doTestOIDCIDJAGRedeem(t *testing.T, mock *mocks.MockAutheliaCtx, claims map[string]any) (code int, response map[string]any) {
	t.Helper()

	return doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTBearer, mustSignTestOIDCIDJAGRedeemGrant(t, claims)))
}

func mustGetTestOIDCIDJAGRedeemStoredLevel(t *testing.T, mock *mocks.MockAutheliaCtx, token string) (level authentication.Level) {
	t.Helper()

	authn := &Authn{Header: HeaderAuthorization{Authorization: model.NewAuthorization()}}

	require.NoError(t, authn.Header.Authorization.ParseBearer(token))

	target, err := url.Parse(testIDJAGGovernedResource)
	require.NoError(t, err)

	username, _, ccs, level, err := handleVerifyGETAuthorizationBearerIntrospection(mock.Ctx, mock.Ctx.Providers.OpenIDConnect, authn, &authorization.Object{URL: target, Method: fasthttp.MethodGet}, nil)

	require.NoError(t, err)
	assert.False(t, ccs)
	assert.Equal(t, testUsername, username)

	return level
}

func assertTestOIDCIDJAGRedeemError(t *testing.T, code int, response map[string]any, expected, hint string) {
	t.Helper()

	assert.Equal(t, http.StatusBadRequest, code, response)
	assert.Equal(t, expected, response["error"], response)

	if hint != "" {
		assert.Contains(t, response["error_description"], hint, response)
	}
}

func mustGetTestOIDCIDJAGRedeemAccessToken(t *testing.T, code int, response map[string]any) (token string) {
	t.Helper()

	require.Equal(t, http.StatusOK, code, response)

	token, ok := response["access_token"].(string)

	require.True(t, ok, response)
	require.NotEmpty(t, token)

	return token
}

func TestOAuth2IDJAGRedeem(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock)

	code, response := doTestOIDCIDJAGRedeem(t, mock, newTestOIDCIDJAGRedeemClaims(t, mock))

	token := mustGetTestOIDCIDJAGRedeemAccessToken(t, code, response)

	introspection := mustGetTestOIDCIntrospection(t, mock, token)

	assert.Equal(t, true, introspection["active"], introspection)
	assert.Equal(t, mustGetTestOIDCSubject(t, mock, testOIDCTokenExchangeID).String(), introspection[oidc.ClaimSubject], introspection)
	assert.Equal(t, testUsername, introspection["username"], introspection)
	assert.Equal(t, testOIDCTokenExchangeID, introspection[oidc.ClaimClientIdentifier], introspection)

	assert.Equal(t, authentication.OneFactor, mustGetTestOIDCIDJAGRedeemStoredLevel(t, mock, token))
}

func TestOAuth2IDJAGRedeemJWTProfileAccessTokenGrantedScopes(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		config.ClaimsPolicies = map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
			"access": {AccessToken: []string{oidc.ClaimPreferredUsername}},
		}

		config.Discovery.JWTResponseAccessTokens = true

		exchange.ClaimsPolicy = "access"
		exchange.AccessTokenSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
		exchange.AccessTokenSignedResponseKeyID = testOIDCKeyID
	})

	mock.SetLogLevel(logrus.DebugLevel)

	code, response := doTestOIDCIDJAGRedeem(t, mock, newTestOIDCIDJAGRedeemClaims(t, mock))

	claims := mustDecodeTestJWTClaims(t, mustGetTestOIDCIDJAGRedeemAccessToken(t, code, response))

	assert.Equal(t, testUsername, claims[oidc.ClaimPreferredUsername], claims)

	var hydrated bool

	for _, entry := range mock.Hook.AllEntries() {
		if entry.Message == "Hydrate JWT Profile Access Token claims" {
			hydrated = true

			assert.ElementsMatch(t, oauthelia2.Arguments{testOIDCScopeBearerAuthz, oidc.ScopeProfile}, entry.Data["scope"])
		}
	}

	assert.True(t, hydrated)
}

func TestOAuth2IDJAGRedeemCELTwoFactor(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock, setupTestOIDCIDJAGRedeemAMRMutator(testIDJAGCELTwoFactor), setupTestOIDCIDJAGRedeemTwoFactorMutator)

	claims := newTestOIDCIDJAGRedeemClaims(t, mock)
	claims[oidc.ClaimAuthenticationMethodsReference] = []string{"mfa"}

	authTime := time.Now().Add(-time.Minute).Unix()

	claims[oidc.ClaimAuthenticationTime] = authTime

	code, response := doTestOIDCIDJAGRedeem(t, mock, claims)

	token := mustGetTestOIDCIDJAGRedeemAccessToken(t, code, response)

	assert.Equal(t, authentication.TwoFactor, mustGetTestOIDCIDJAGRedeemStoredLevel(t, mock, token))

	_, requester, err := mock.Ctx.Providers.OpenIDConnect.IntrospectToken(mock.Ctx, token, oauthelia2.AccessToken, oidc.NewSession())

	require.NoError(t, err)

	session, ok := requester.GetSession().(*oidc.Session)

	require.True(t, ok)
	require.NotNil(t, session.Claims)
	require.NotNil(t, session.Claims.AuthTime)

	assert.Equal(t, []string{"pwd", "otp"}, session.Claims.AuthenticationMethodsReferences)
	assert.Equal(t, authTime, session.Claims.AuthTime.Unix())
}

func TestOAuth2IDJAGRedeemCELOneFactor(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock, setupTestOIDCIDJAGRedeemAMRMutator(testIDJAGCELTwoFactor), setupTestOIDCIDJAGRedeemTwoFactorMutator)

	claims := newTestOIDCIDJAGRedeemClaims(t, mock)
	claims[oidc.ClaimAuthenticationMethodsReference] = []string{"pwd"}

	code, response := doTestOIDCIDJAGRedeem(t, mock, claims)

	assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", testIDJAGHintInsufficient)
}

func TestOAuth2IDJAGRedeemCELErrorRejects(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock, setupTestOIDCIDJAGRedeemAMRMutator(testIDJAGCELNoHas))

	code, response := doTestOIDCIDJAGRedeem(t, mock, newTestOIDCIDJAGRedeemClaims(t, mock))

	assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", testIDJAGHintMapping)

	var logged bool

	for _, entry := range mock.Hook.AllEntries() {
		if entry.Level == logrus.ErrorLevel && entry.Data["error"] != nil {
			logged = logged || strings.Contains(fmt.Sprint(entry.Data["error"]), "error evaluating the expression")
		}
	}

	assert.True(t, logged, "the evaluation error should be logged at the error level")
}

func TestOAuth2IDJAGRedeemNoExpression(t *testing.T) {
	t.Run("ShouldPermitOneFactor", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock)

		claims := newTestOIDCIDJAGRedeemClaims(t, mock)
		claims[oidc.ClaimAuthenticationMethodsReference] = []string{"mfa"}

		code, response := doTestOIDCIDJAGRedeem(t, mock, claims)

		token := mustGetTestOIDCIDJAGRedeemAccessToken(t, code, response)

		assert.Equal(t, authentication.OneFactor, mustGetTestOIDCIDJAGRedeemStoredLevel(t, mock, token))
	})

	t.Run("ShouldRejectTwoFactor", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock, setupTestOIDCIDJAGRedeemTwoFactorMutator)

		claims := newTestOIDCIDJAGRedeemClaims(t, mock)
		claims[oidc.ClaimAuthenticationMethodsReference] = []string{"mfa"}

		code, response := doTestOIDCIDJAGRedeem(t, mock, claims)

		assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", testIDJAGHintInsufficient)
	})
}

func TestOAuth2IDJAGRedeemUntrustedIssuer(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock)

	claims := newTestOIDCIDJAGRedeemClaims(t, mock)
	claims[oidc.ClaimIssuer] = "https://untrusted.example.com"

	code, response := doTestOIDCIDJAGRedeem(t, mock, claims)

	assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", "The issuer 'https://untrusted.example.com' is not trusted.")
}

func TestOAuth2IDJAGRedeemClientNotListed(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
		config.IdentityAssertionAuthorizationGrant.TrustedIssuers[0].Clients = []string{testOIDCActorID}
	})

	code, response := doTestOIDCIDJAGRedeem(t, mock, newTestOIDCIDJAGRedeemClaims(t, mock))

	assertTestOIDCIDJAGRedeemError(t, code, response, "access_denied", "The OAuth 2.0 Client is not permitted to present grants from the issuer")
}

func TestOAuth2IDJAGRedeemResourceNotGoverned(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock)

	claims := newTestOIDCIDJAGRedeemClaims(t, mock)
	claims[testIDJAGClaimResource] = "https://other.example.com/"

	code, response := doTestOIDCIDJAGRedeem(t, mock, claims)

	assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", "The resource 'https://other.example.com/' of the Identity Assertion JWT Authorization Grant is not governed by this authorization server.")
}

func TestOAuth2IDJAGRedeemReplay(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock)

	values := newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTBearer, mustSignTestOIDCIDJAGRedeemGrant(t, newTestOIDCIDJAGRedeemClaims(t, mock)))

	code, response := doTestOIDCTokenExchange(t, mock, values)

	require.Equal(t, http.StatusOK, code, response)

	code, response = doTestOIDCTokenExchange(t, mock, values)

	assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", "")
}

func TestOAuth2IDJAGRedeemConsentModes(t *testing.T) {
	consentModeMutator := func(mode string) testOIDCTokenExchangeMutator {
		return func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
			exchange.ConsentMode = mode

			if mode == "pre-configured" {
				duration := time.Hour

				exchange.ConsentPreConfiguredDuration = &duration
			}
		}
	}

	t.Run("ShouldPermitImplicit", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock, consentModeMutator("implicit"))

		code, response := doTestOIDCIDJAGRedeem(t, mock, newTestOIDCIDJAGRedeemClaims(t, mock))

		assert.Equal(t, http.StatusOK, code, response)
	})

	t.Run("ShouldRejectExplicit", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock, consentModeMutator("explicit"))

		code, response := doTestOIDCIDJAGRedeem(t, mock, newTestOIDCIDJAGRedeemClaims(t, mock))

		assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", testIDJAGHintExplicit)
	})

	governed := model.StringSlicePipeDelimited{testIDJAGGovernedResource}
	covering := model.StringSlicePipeDelimited{oidc.ScopeProfile, testOIDCScopeBearerAuthz}

	preConfiguredCases := []struct {
		name      string
		scopes    model.StringSlicePipeDelimited
		resource  model.StringSlicePipeDelimited
		revoked   bool
		omitScope bool
		expected  int
	}{
		{"ShouldPermitPreConfiguredCovering", covering, governed, false, false, http.StatusOK},
		{"ShouldRejectPreConfiguredNotCovering", model.StringSlicePipeDelimited{oidc.ScopeOpenID}, governed, false, false, http.StatusBadRequest},
		{"ShouldRejectPreConfiguredRevoked", covering, governed, true, false, http.StatusBadRequest},
		{"ShouldPermitPreConfiguredCoveringWithoutScopeParameter", covering, governed, false, true, http.StatusOK},
		{"ShouldRejectPreConfiguredNotCoveringWithoutScopeParameter", model.StringSlicePipeDelimited{oidc.ScopeOpenID}, governed, false, true, http.StatusBadRequest},
		{"ShouldRejectPreConfiguredNotCoveringResource", covering, model.StringSlicePipeDelimited{"https://other.example.com/"}, false, true, http.StatusBadRequest},
	}

	for _, tc := range preConfiguredCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			setupTestOIDCIDJAGRedeemFlow(t, mock, consentModeMutator("pre-configured"))

			subject := mustGetTestOIDCSubject(t, mock, testOIDCTokenExchangeID)

			rows, closer := mocks.NewConsentPreConfigRows(t, model.OAuth2ConsentPreConfig{
				ID:        1,
				ClientID:  testOIDCTokenExchangeID,
				Subject:   subject,
				CreatedAt: mock.Ctx.GetClock().Now(),
				Revoked:   tc.revoked,
				Scopes:    tc.scopes,
				Resource:  tc.resource,
			})
			defer closer()

			mock.StorageMock.EXPECT().
				LoadOAuth2ConsentPreConfigurations(gomock.Any(), gomock.Eq(testOIDCTokenExchangeID), gomock.Eq(subject), gomock.Any()).
				Return(rows, nil)

			values := newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTBearer, mustSignTestOIDCIDJAGRedeemGrant(t, newTestOIDCIDJAGRedeemClaims(t, mock)))
			if !tc.omitScope {
				values.Set(oidc.FormParameterScope, strings.Join([]string{oidc.ScopeProfile, testOIDCScopeBearerAuthz}, " "))
			}

			code, response := doTestOIDCTokenExchange(t, mock, values)

			require.Equal(t, tc.expected, code, response)

			if tc.expected != http.StatusOK {
				assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", testIDJAGHintNoConsent)
			}
		})
	}
}

func TestOAuth2IDJAGRedeemJWTDPoP(t *testing.T) {
	newBoundClaims := func(t *testing.T, mock *mocks.MockAutheliaCtx, jkt string) map[string]any {
		claims := newTestOIDCIDJAGRedeemClaims(t, mock)
		claims["cnf"] = map[string]any{"jkt": jkt}

		return claims
	}

	t.Run("ShouldIssueBoundTokenWithProof", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock, setupTestOIDCIDJAGRedeemDPoPMutator)

		key, jkt := newTestOIDCIDJAGDPoPKey(t)

		code, response := doTestOIDCIDJAGWithDPoP(t, mock, key, newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTDPoP, mustSignTestOIDCIDJAGRedeemGrant(t, newBoundClaims(t, mock, jkt))))

		token := mustGetTestOIDCIDJAGRedeemAccessToken(t, code, response)

		assert.Equal(t, "DPoP", response["token_type"], response)

		introspection := mustGetTestOIDCIntrospection(t, mock, token)

		assert.Equal(t, map[string]any{"jkt": jkt}, introspection["cnf"], introspection)
	})

	t.Run("ShouldRefuseBoundGrantAsJWTBearer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock, setupTestOIDCIDJAGRedeemDPoPMutator)

		key, jkt := newTestOIDCIDJAGDPoPKey(t)

		mock.StorageMock.EXPECT().
			CheckAndSetOAuth2DPoPProofUsed(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(false, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTBearer, mustSignTestOIDCIDJAGRedeemGrant(t, newBoundClaims(t, mock, jkt))))

		r.Header.Set(oidc.HeaderDPoP, newTestOIDCIDJAGDPoPProof(t, key, fasthttp.MethodPost, testOIDCTokenEndpoint))

		OAuth2TokenPOST(mock.Ctx, rw, r)

		assertTestOIDCIDJAGRedeemError(t, rw.Code, getTestOAuth2ErrorResponse(t, rw), "invalid_grant", "requires the 'urn:ietf:params:oauth:grant-type:jwt-dpop' authorization grant")
	})

	t.Run("ShouldRejectWithoutProof", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock, setupTestOIDCIDJAGRedeemDPoPMutator)

		_, jkt := newTestOIDCIDJAGDPoPKey(t)

		code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTDPoP, mustSignTestOIDCIDJAGRedeemGrant(t, newBoundClaims(t, mock, jkt))))

		assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", "")
	})
}

func getTestOIDCIDJAGRedeemAuditEntries(mock *mocks.MockAutheliaCtx) (entries []*logrus.Entry) {
	for _, entry := range mock.Hook.AllEntries() {
		if strings.HasPrefix(entry.Message, "Identity Assertion JWT Authorization Grant ") {
			entries = append(entries, entry)
		}
	}

	return entries
}

func assertTestOIDCIDJAGRedeemNotLogged(t *testing.T, mock *mocks.MockAutheliaCtx, values ...string) {
	t.Helper()

	for _, e := range mock.Hook.AllEntries() {
		for _, value := range values {
			assert.NotContains(t, e.Message, value)

			for _, v := range e.Data {
				assert.NotContains(t, fmt.Sprint(v), value)
			}
		}
	}
}

func TestOAuth2IDJAGRedeemAuditLog(t *testing.T) {
	t.Run("ShouldLogRedeemed", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock)

		claims := newTestOIDCIDJAGRedeemClaims(t, mock)
		grant := mustSignTestOIDCIDJAGRedeemGrant(t, claims)

		mock.Hook.Reset()

		code, response := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTBearer, grant))

		token := mustGetTestOIDCIDJAGRedeemAccessToken(t, code, response)

		entries := getTestOIDCIDJAGRedeemAuditEntries(mock)

		require.Len(t, entries, 1)

		entry := entries[0]

		assert.Equal(t, logrus.InfoLevel, entry.Level)
		assert.Equal(t, "Identity Assertion JWT Authorization Grant redeemed", entry.Message)
		assert.Equal(t, testOIDCTokenExchangeID, entry.Data["client_id"])
		assert.Equal(t, oidc.GrantTypeJWTBearer, entry.Data["grant_type"])
		assert.Equal(t, testIDJAGTrustedIssuer, entry.Data["iss"])
		assert.Equal(t, claims[oidc.ClaimJWTID], entry.Data["jti"])
		assert.Equal(t, testIDJAGExternalSubject, entry.Data["sub"])
		assert.Equal(t, testUsername, entry.Data["username"])
		assert.Equal(t, mustGetTestOIDCSubject(t, mock, testOIDCTokenExchangeID).String(), entry.Data["subject"])
		assert.Equal(t, []string{testOIDCScopeBearerAuthz, oidc.ScopeProfile}, entry.Data["granted_scope"])
		assert.Equal(t, []string{testIDJAGGovernedResource}, entry.Data["granted_resource"])
		assert.Equal(t, false, entry.Data["dpop_bound"])
		assert.NotEmpty(t, entry.Data["remote_ip"])

		assertTestOIDCIDJAGRedeemNotLogged(t, mock, grant, token)
	})

	t.Run("ShouldLogDeniedByPolicy", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock, setupTestOIDCIDJAGRedeemTwoFactorMutator)

		claims := newTestOIDCIDJAGRedeemClaims(t, mock)
		grant := mustSignTestOIDCIDJAGRedeemGrant(t, claims)

		mock.Hook.Reset()

		code, _ := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTBearer, grant))

		require.Equal(t, http.StatusBadRequest, code)

		entries := getTestOIDCIDJAGRedeemAuditEntries(mock)

		require.Len(t, entries, 1)

		assert.Equal(t, logrus.WarnLevel, entries[0].Level)
		assert.Equal(t, "Identity Assertion JWT Authorization Grant denied", entries[0].Message)
		assert.Equal(t, testIDJAGTrustedIssuer, entries[0].Data["iss"])
		assert.Equal(t, claims[oidc.ClaimJWTID], entries[0].Data["jti"])
		assert.Contains(t, fmt.Sprint(entries[0].Data["error"]), testIDJAGHintInsufficient)

		assertTestOIDCIDJAGRedeemNotLogged(t, mock, grant)
	})

	t.Run("ShouldLogDeniedAccessRequest", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock, func(config *schema.IdentityProvidersOpenIDConnect, subject, exchange *schema.IdentityProvidersOpenIDConnectClient) {
			config.IdentityAssertionAuthorizationGrant.TrustedIssuers[0].Clients = []string{testOIDCActorID}
		})

		grant := mustSignTestOIDCIDJAGRedeemGrant(t, newTestOIDCIDJAGRedeemClaims(t, mock))

		mock.Hook.Reset()

		code, _ := doTestOIDCTokenExchange(t, mock, newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTBearer, grant))

		require.Equal(t, http.StatusBadRequest, code)

		entries := getTestOIDCIDJAGRedeemAuditEntries(mock)

		require.Len(t, entries, 1)

		assert.Equal(t, logrus.WarnLevel, entries[0].Level)
		assert.Equal(t, "Identity Assertion JWT Authorization Grant denied", entries[0].Message)
		assert.Equal(t, testOIDCTokenExchangeID, entries[0].Data["client_id"])
		assert.Equal(t, oidc.GrantTypeJWTBearer, entries[0].Data["grant_type"])
		assert.NotContains(t, entries[0].Data, "iss")
		assert.NotContains(t, entries[0].Data, "jti")
		assert.NotContains(t, entries[0].Data, "sub")

		assertTestOIDCIDJAGRedeemNotLogged(t, mock, grant)
	})

	t.Run("ShouldNotLogUnauthenticatedClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCIDJAGRedeemFlow(t, mock)

		values := newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTBearer, mustSignTestOIDCIDJAGRedeemGrant(t, newTestOIDCIDJAGRedeemClaims(t, mock)))
		values.Set(testOIDCFormParameterClientSecret, "wrong")

		mock.Hook.Reset()

		code, _ := doTestOIDCTokenExchange(t, mock, values)

		require.Equal(t, http.StatusUnauthorized, code)

		assert.Empty(t, getTestOIDCIDJAGRedeemAuditEntries(mock))
	})
}

func TestOAuth2IDJAGRedeemDeniedDoesNotConsume(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCIDJAGRedeemFlow(t, mock, setupTestOIDCIDJAGRedeemTwoFactorMutator)

	values := newTestOIDCIDJAGRedeemValues(oidc.GrantTypeJWTBearer, mustSignTestOIDCIDJAGRedeemGrant(t, newTestOIDCIDJAGRedeemClaims(t, mock)))

	code, response := doTestOIDCTokenExchange(t, mock, values)

	assertTestOIDCIDJAGRedeemError(t, code, response, "invalid_grant", testIDJAGHintInsufficient)

	client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, testOIDCTokenExchangeID)

	require.NoError(t, err)

	registered, ok := client.(*oidc.RegisteredClient)

	require.True(t, ok)

	registered.AuthorizationPolicy = oidc.ClientAuthorizationPolicy{DefaultPolicy: authorization.OneFactor}

	code, response = doTestOIDCTokenExchange(t, mock, values)

	assert.Equal(t, http.StatusOK, code, response)
}
