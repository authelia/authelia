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

	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	fjwt "authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestOpenIDConnectUserinfoDPoP(t *testing.T) {
	const (
		schemeBearer = "Bearer"
		targetURI    = "https://login.example.com:8080/api/oidc/userinfo"
	)

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
			Scheme:         schemeBearer,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_dpop_proof",
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
			Scheme:         schemeBearer,
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
				DPoPJWKThumbprint: jkt,
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

			r := httptest.NewRequest(fasthttp.MethodGet, targetURI, nil)
			r.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			r.Header.Set(fasthttp.HeaderAuthorization, tc.Scheme+" "+token)

			if tc.Proof {
				r.Header.Set(oidc.HeaderDPoP, newTestDPoPProof(t, key, fasthttp.MethodGet, targetURI, token))

				if tc.NonceEnforced {
					mock.StorageMock.EXPECT().
						SaveOAuth2DPoPNonce(gomock.Any(), gomock.Any()).
						Return(nil)
				} else {
					mock.StorageMock.EXPECT().
						CheckAndSetOAuth2DPoPProofUsed(gomock.Any(), gomock.Any(), gomock.Eq(fasthttp.MethodGet), gomock.Eq(targetURI), gomock.Any(), gomock.Any()).
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
	const (
		schemeBearer = "Bearer"
		targetURI    = "https://login.example.com:8080/api/oidc/userinfo"
	)

	testCases := []struct {
		Name           string
		Scheme         string
		RefreshToken   bool
		ExpectedStatus int
		ExpectedScheme string
	}{
		{
			Name:           "ShouldChallengeWithBearerWhenIntrospectionFailsUnderBearerScheme",
			Scheme:         schemeBearer,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedScheme: schemeBearer,
		},
		{
			Name:           "ShouldChallengeWithDPoPWhenIntrospectionFailsUnderDPoPScheme",
			Scheme:         oidc.SchemeDPoP,
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedScheme: oidc.SchemeDPoP,
		},
		{
			Name:           "ShouldChallengeWithBearerWhenTokenIsNotAnAccessTokenUnderBearerScheme",
			Scheme:         schemeBearer,
			RefreshToken:   true,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedScheme: schemeBearer,
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

			r := httptest.NewRequest(fasthttp.MethodGet, targetURI, nil)
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
