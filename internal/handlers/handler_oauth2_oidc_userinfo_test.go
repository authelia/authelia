// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"authelia.com/provider/jose"
	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	fjwt "authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

const testOIDCUserinfoEndpoint = "https://login.example.com:8080/api/oidc/userinfo"

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
		assert.Contains(t, rw.Header().Get(fasthttp.HeaderWWWAuthenticate), `error="request_unauthorized"`)

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

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCClientCredentialsClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSessionStore(t, mock)

		token := mustGetTestOIDCClientCredentialsToken(t, mock)

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example2.com")

		rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, "https://auth.example2.com/api/oidc/userinfo", nil)

		r.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+token)

		OpenIDConnectUserinfo(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpUserInfoIssuerMismatch, nil)
	})
}

func TestOpenIDConnectUserinfoDPoP(t *testing.T) {
	testCases := []struct {
		Name           string
		JKT            string
		Scheme         string
		Proof          bool
		Enforced       bool
		NonceEnforced  bool
		NonceStateless bool
		ProofNonce     bool
		StorageError   bool
		RequestHost    string
		ExpectedStatus int
		ExpectedError  string
		ExpectedNonce  bool
	}{
		{
			Name:           "ShouldRejectBoundTokenPresentedAsBearer",
			JKT:            "D6Nq0uHi1xL9fbLBu6xVGvKtOsBqiOxfHy_hOZlLzHM",
			Scheme:         "Bearer",
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_token",
		},
		{
			Name:           "ShouldRejectBoundTokenPresentedAsDPoPWithoutProof",
			JKT:            "D6Nq0uHi1xL9fbLBu6xVGvKtOsBqiOxfHy_hOZlLzHM",
			Scheme:         oidc.SchemeDPoP,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_dpop_proof",
		},
		{
			Name:           "ShouldRejectUnboundTokenPresentedAsDPoP",
			Scheme:         oidc.SchemeDPoP,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_token",
		},
		{
			Name:           "ShouldAllowUnboundTokenPresentedAsBearer",
			Scheme:         "Bearer",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "ShouldRejectUnboundTokenPresentedAsBearerWhenEnforced",
			Scheme:         "Bearer",
			Enforced:       true,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_token",
		},
		{
			Name:           "ShouldReturnServerErrorWhenReplayStorageFails",
			Scheme:         oidc.SchemeDPoP,
			Proof:          true,
			StorageError:   true,
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedError:  "server_error",
		},
		{
			Name:           "ShouldAllowBoundTokenPresentedAsDPoPWithValidProof",
			Scheme:         oidc.SchemeDPoP,
			Proof:          true,
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "ShouldAllowProofForConfiguredEndpointURLWhenRequestHostDiffers",
			Scheme:         oidc.SchemeDPoP,
			Proof:          true,
			RequestHost:    "authelia:9091",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "ShouldChallengeBoundTokenPresentedAsDPoPWithProofMissingNonce",
			Scheme:         oidc.SchemeDPoP,
			Proof:          true,
			NonceEnforced:  true,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "use_dpop_nonce",
			ExpectedNonce:  true,
		},
		{
			Name:           "ShouldChallengeWithStatelessNonceWithoutStoringIt",
			Scheme:         oidc.SchemeDPoP,
			Proof:          true,
			NonceEnforced:  true,
			NonceStateless: true,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "use_dpop_nonce",
			ExpectedNonce:  true,
		},
		{
			Name:           "ShouldAllowProofWithStatelessNonceWithoutLoadingIt",
			Scheme:         oidc.SchemeDPoP,
			Proof:          true,
			NonceEnforced:  true,
			NonceStateless: true,
			ProofNonce:     true,
			ExpectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			setUpMockClock(mock)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "login.example.com:8080")

			mock.Ctx.Configuration.IdentityProviders = schema.IdentityProviders{
				OIDC: &schema.IdentityProvidersOpenIDConnect{
					HMACSecret: "abcdefghijklmnopqrstuvwxyz123456",
					DPoP: schema.IdentityProvidersOpenIDConnectDPoP{
						Enabled:        true,
						Enforced:       tc.Enforced,
						ClockSkew:      time.Minute,
						NonceEnforced:  tc.NonceEnforced,
						NonceStateless: tc.NonceStateless,
						NonceLifespan:  time.Minute,
					},
					Clients: []schema.IdentityProvidersOpenIDConnectClient{
						{
							ID:                  "test-userinfo-client",
							Scopes:              []string{oidc.ScopeOpenID},
							GrantTypes:          []string{oidc.GrantTypeClientCredentials},
							AuthorizationPolicy: "one_factor",
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(&mock.Ctx.Configuration, mock.StorageMock, mock.Ctx.Providers.Templates)

			client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, "test-userinfo-client")
			require.NoError(t, err)

			var (
				key *ecdsa.PrivateKey
				jkt = tc.JKT
			)

			if tc.Proof {
				key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)

				jkt, err = fjwt.ThumbprintJWK(&jose.JSONWebKey{Key: key.Public()})
				require.NoError(t, err)
			}

			now := mock.Ctx.Providers.Clock.Now()

			session := &oidc.Session{
				ClientID:          "test-userinfo-client",
				ClientCredentials: true,
				DefaultSession: &openid.DefaultSession{
					JWKThumbprint: jkt,
					Headers:       &fjwt.Headers{Extra: map[string]any{}},
					Claims: &fjwt.IDTokenClaims{
						Issuer:   "https://login.example.com:8080",
						Subject:  "test-userinfo-client",
						IssuedAt: fjwt.NewNumericDate(now),
						Extra:    map[string]any{},
					},
					RequestedAt: now,
				},
			}

			requester := &oauthelia2.AccessRequest{
				GrantTypes: oauthelia2.Arguments{oidc.GrantTypeClientCredentials},
				Request: oauthelia2.Request{
					ID:             "request-userinfo",
					RequestedAt:    now,
					Client:         client,
					RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID},
					GrantedScope:   oauthelia2.Arguments{oidc.ScopeOpenID},
					Session:        session,
					Form:           url.Values{},
				},
			}

			token, signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.GenerateAccessToken(mock.Ctx, requester)
			require.NoError(t, err)

			oauthSession, err := model.NewOAuth2SessionFromRequest(signature, requester)
			require.NoError(t, err)

			mock.StorageMock.EXPECT().
				LoadOAuth2Session(gomock.Any(), gomock.Eq(storage.OAuth2SessionTypeAccessToken), gomock.Eq(signature)).
				Return(oauthSession, nil)

			r := httptest.NewRequest(fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)
			r.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			r.Header.Set(fasthttp.HeaderAuthorization, tc.Scheme+" "+token)

			if tc.RequestHost != "" {
				r.Host = tc.RequestHost
			}

			if tc.Proof {
				var nonce string

				if tc.ProofNonce {
					nonce, err = mock.Ctx.Providers.OpenIDConnect.GetDPoPStrategy(mock.Ctx).NewDPoPNonce(mock.Ctx)
					require.NoError(t, err)
				}

				proof := newTestDPoPProofWithNonce(t, key, fasthttp.MethodGet, testOIDCUserinfoEndpoint, token, nonce)

				r.Header.Set(oidc.HeaderDPoP, proof)

				switch {
				case tc.NonceEnforced && !tc.ProofNonce && tc.NonceStateless:
				case tc.NonceEnforced && !tc.ProofNonce:
					mock.StorageMock.EXPECT().
						SaveOAuth2DPoPNonce(gomock.Any(), gomock.Any()).
						Return(nil)
				case tc.StorageError:
					mock.StorageMock.EXPECT().
						CheckAndSetOAuth2DPoPProofUsed(gomock.Any(), newTestDPoPProofMatcher(t, proof, jkt, fasthttp.MethodGet, testOIDCUserinfoEndpoint), gomock.Any()).
						Return(false, fmt.Errorf("failed to check proof"))
				default:
					mock.StorageMock.EXPECT().
						CheckAndSetOAuth2DPoPProofUsed(gomock.Any(), newTestDPoPProofMatcher(t, proof, jkt, fasthttp.MethodGet, testOIDCUserinfoEndpoint), gomock.Any()).
						Return(false, nil)
				}
			}

			rw := httptest.NewRecorder()

			OpenIDConnectUserinfo(mock.Ctx, rw, r)

			assert.Equal(t, tc.ExpectedStatus, rw.Code)

			body := map[string]any{}

			require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))

			if tc.ExpectedNonce {
				assert.NotEmpty(t, rw.Header().Get(oidc.HeaderDPoPNonce))
			} else {
				assert.Empty(t, rw.Header().Get(oidc.HeaderDPoPNonce))
			}

			if tc.ExpectedError == "" {
				assert.Empty(t, rw.Header().Get(fasthttp.HeaderWWWAuthenticate))
				assert.Equal(t, "test-userinfo-client", body[oidc.ClaimSubject])

				return
			}

			assert.Equal(t, tc.ExpectedError, body["error"])

			challenge := rw.Header().Get(fasthttp.HeaderWWWAuthenticate)

			assert.Contains(t, challenge, fmt.Sprintf(`%s error="%s"`, tc.Scheme, tc.ExpectedError))
			assert.Contains(t, challenge, `DPoP `)
			assert.Contains(t, challenge, `algs="RS256 RS384 RS512 ES256 ES384 ES512 PS256 PS384 PS512 Ed25519 EdDSA"`)
		})
	}
}

func TestOpenIDConnectUserinfoChallengeScheme(t *testing.T) {
	testCases := []struct {
		Name           string
		Scheme         string
		RefreshToken   bool
		ExpectedStatus int
		ExpectedScheme string
	}{
		{
			Name:           "ShouldChallengeWithBearerWhenIntrospectionFailsUnderBearerScheme",
			Scheme:         "Bearer",
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedScheme: "Bearer",
		},
		{
			Name:           "ShouldChallengeWithDPoPWhenIntrospectionFailsUnderDPoPScheme",
			Scheme:         oidc.SchemeDPoP,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedScheme: oidc.SchemeDPoP,
		},
		{
			Name:           "ShouldChallengeWithBearerWhenTokenIsNotAnAccessTokenUnderBearerScheme",
			Scheme:         "Bearer",
			RefreshToken:   true,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedScheme: "Bearer",
		},
		{
			Name:           "ShouldChallengeWithDPoPWhenTokenIsNotAnAccessTokenUnderDPoPScheme",
			Scheme:         oidc.SchemeDPoP,
			RefreshToken:   true,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedScheme: oidc.SchemeDPoP,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			setUpMockClock(mock)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "login.example.com:8080")

			mock.Ctx.Configuration.IdentityProviders = schema.IdentityProviders{
				OIDC: &schema.IdentityProvidersOpenIDConnect{
					HMACSecret: "abcdefghijklmnopqrstuvwxyz123456",
					DPoP: schema.IdentityProvidersOpenIDConnectDPoP{
						Enabled:   true,
						ClockSkew: time.Minute,
					},
					Clients: []schema.IdentityProvidersOpenIDConnectClient{
						{
							ID:                  "test-userinfo-client",
							Scopes:              []string{oidc.ScopeOpenID},
							GrantTypes:          []string{oidc.GrantTypeClientCredentials, oidc.GrantTypeRefreshToken},
							AuthorizationPolicy: "one_factor",
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(&mock.Ctx.Configuration, mock.StorageMock, mock.Ctx.Providers.Templates)

			token := "authelia_at_not-a-valid-token"

			if tc.RefreshToken {
				client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, "test-userinfo-client")
				require.NoError(t, err)

				now := mock.Ctx.Providers.Clock.Now()

				requester := &oauthelia2.AccessRequest{
					GrantTypes: oauthelia2.Arguments{oidc.GrantTypeClientCredentials},
					Request: oauthelia2.Request{
						ID:             "request-userinfo",
						RequestedAt:    now,
						Client:         client,
						RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID},
						GrantedScope:   oauthelia2.Arguments{oidc.ScopeOpenID},
						Form:           url.Values{},
						Session: &oidc.Session{
							ClientID:          "test-userinfo-client",
							ClientCredentials: true,
							DefaultSession: &openid.DefaultSession{
								Headers: &fjwt.Headers{Extra: map[string]any{}},
								Claims: &fjwt.IDTokenClaims{
									Issuer:   "https://login.example.com:8080",
									Subject:  "test-userinfo-client",
									IssuedAt: fjwt.NewNumericDate(now),
									Extra:    map[string]any{},
								},
								RequestedAt: now,
							},
						},
					},
				}

				var signature string

				token, signature, err = mock.Ctx.Providers.OpenIDConnect.Strategy.Core.GenerateRefreshToken(mock.Ctx, requester)
				require.NoError(t, err)

				oauthSession, err := model.NewOAuth2SessionFromRequest(signature, requester)
				require.NoError(t, err)

				mock.StorageMock.EXPECT().
					LoadOAuth2Session(gomock.Any(), gomock.Eq(storage.OAuth2SessionTypeRefreshToken), gomock.Eq(signature)).
					Return(oauthSession, nil)
			}

			r := httptest.NewRequest(fasthttp.MethodGet, testOIDCUserinfoEndpoint, nil)
			r.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			r.Header.Set(fasthttp.HeaderAuthorization, tc.Scheme+" "+token)

			rw := httptest.NewRecorder()

			OpenIDConnectUserinfo(mock.Ctx, rw, r)

			assert.Equal(t, tc.ExpectedStatus, rw.Code)
			challenge := rw.Header().Get(fasthttp.HeaderWWWAuthenticate)

			if tc.ExpectedScheme == oidc.SchemeDPoP {
				assert.Contains(t, challenge, `DPoP error="`)
			} else {
				assert.True(t, strings.HasPrefix(challenge, "Bearer "), challenge)
				assert.Contains(t, challenge, `error="`)
			}
		})
	}
}

func newTestDPoPProof(t *testing.T, key *ecdsa.PrivateKey, method, target, token string) (proof string) {
	t.Helper()

	return newTestDPoPProofWithNonce(t, key, method, target, token, "")
}

func newTestDPoPProofWithNonce(t *testing.T, key *ecdsa.PrivateKey, method, target, token, nonce string) (proof string) {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: key},
		(&jose.SignerOptions{EmbedJWK: true}).WithType(fjwt.JSONWebTokenTypeDPoP),
	)
	require.NoError(t, err)

	sum := sha256.Sum256([]byte(token))

	claims := map[string]any{
		"jti": uuid.Must(uuid.NewRandom()).String(),
		"htm": method,
		"htu": target,
		"iat": time.Now().Unix(),
		"ath": base64.RawURLEncoding.EncodeToString(sum[:]),
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}

	payload, err := json.Marshal(claims)
	require.NoError(t, err)

	object, err := signer.Sign(payload)
	require.NoError(t, err)

	proof, err = object.CompactSerialize()
	require.NoError(t, err)

	return proof
}

func newTestDPoPProofMatcher(t *testing.T, proof, jkt, htm, htu string) gomock.Matcher {
	t.Helper()

	parts := strings.Split(proof, ".")
	require.Len(t, parts, 3)

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)

	claims := map[string]any{}

	require.NoError(t, json.Unmarshal(payload, &claims))

	jti, ok := claims["jti"].(string)
	require.True(t, ok)

	signature := model.NewOAuth2DPoPProofSignature(jti, jkt, htm, htu)

	return gomock.Cond(func(x any) bool {
		value, ok := x.(model.OAuth2DPoPProof)

		return ok && value.Signature == signature && !value.ExpiresAt.IsZero()
	})
}
