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

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	"authelia.com/provider/oauth2/token/jose"
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
		assert.Contains(t, rw.Header().Get(fasthttp.HeaderWWWAuthenticate), `error="request_unauthorized"`)

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
		NonceEnforced  bool
		ExpectedStatus int
		ExpectedError  string
		ExpectedNonce  bool
	}{
		{
			Name:           "ShouldRejectBoundTokenPresentedAsBearer",
			JKT:            "D6Nq0uHi1xL9fbLBu6xVGvKtOsBqiOxfHy_hOZlLzHM",
			Scheme:         schemeOpenIDConnectUserinfoBearer,
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
			Scheme:         schemeOpenIDConnectUserinfoBearer,
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "ShouldAllowBoundTokenPresentedAsDPoPWithValidProof",
			Scheme:         oidc.SchemeDPoP,
			Proof:          true,
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
						Enabled:       true,
						ClockSkew:     time.Minute,
						NonceEnforced: tc.NonceEnforced,
						NonceLifespan: time.Minute,
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

			if tc.Proof {
				r.Header.Set(oidc.HeaderDPoP, newTestDPoPProof(t, key, fasthttp.MethodGet, testOIDCUserinfoEndpoint, token))

				if tc.NonceEnforced {
					mock.StorageMock.EXPECT().
						SaveOAuth2DPoPNonce(gomock.Any(), gomock.Any()).
						Return(nil)
				} else {
					mock.StorageMock.EXPECT().
						CheckAndSetOAuth2DPoPProofUsed(gomock.Any(), gomock.Any(), gomock.Eq(fasthttp.MethodGet), gomock.Eq(testOIDCUserinfoEndpoint), gomock.Any(), gomock.Any()).
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
			assert.True(t, strings.HasPrefix(rw.Header().Get(fasthttp.HeaderWWWAuthenticate), oidc.SchemeDPoP+" "), rw.Header().Get(fasthttp.HeaderWWWAuthenticate))
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
			Scheme:         schemeOpenIDConnectUserinfoBearer,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedScheme: schemeOpenIDConnectUserinfoBearer,
		},
		{
			Name:           "ShouldChallengeWithDPoPWhenIntrospectionFailsUnderDPoPScheme",
			Scheme:         oidc.SchemeDPoP,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedScheme: oidc.SchemeDPoP,
		},
		{
			Name:           "ShouldChallengeWithBearerWhenTokenIsNotAnAccessTokenUnderBearerScheme",
			Scheme:         schemeOpenIDConnectUserinfoBearer,
			RefreshToken:   true,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedScheme: schemeOpenIDConnectUserinfoBearer,
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
			assert.True(t, strings.HasPrefix(rw.Header().Get(fasthttp.HeaderWWWAuthenticate), tc.ExpectedScheme+" "), rw.Header().Get(fasthttp.HeaderWWWAuthenticate))
		})
	}
}

func newTestDPoPProof(t *testing.T, key *ecdsa.PrivateKey, method, target, token string) (proof string) {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: key},
		(&jose.SignerOptions{EmbedJWK: true}).WithType(fjwt.JSONWebTokenTypeDPoP),
	)
	require.NoError(t, err)

	sum := sha256.Sum256([]byte(token))

	payload, err := json.Marshal(map[string]any{
		"jti": uuid.Must(uuid.NewRandom()).String(),
		"htm": method,
		"htu": target,
		"iat": time.Now().Unix(),
		"ath": base64.RawURLEncoding.EncodeToString(sum[:]),
	})
	require.NoError(t, err)

	object, err := signer.Sign(payload)
	require.NoError(t, err)

	proof, err = object.CompactSerialize()
	require.NoError(t, err)

	return proof
}
