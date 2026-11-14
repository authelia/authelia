// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"authelia.com/provider/jose"
	oauthelia2 "authelia.com/provider/oauth2"
	fjwt "authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewDPoPResourceRequest(t *testing.T) {
	testCases := []struct {
		Name     string
		Method   string
		Target   string
		Token    string
		Proof    string
		Expected string
	}{
		{
			"ShouldReconstructHTTPSTarget",
			http.MethodGet,
			"https://app.example.com/resource?query=value",
			"token",
			"proof",
			"https://app.example.com/resource",
		},
		{
			"ShouldReconstructHTTPTarget",
			http.MethodPost,
			"http://app.example.com/api/v1",
			"token",
			"proof",
			"http://app.example.com/api/v1",
		},
		{
			"ShouldPreserveEscapedPath",
			http.MethodGet,
			"https://app.example.com/resource%3Fx=1",
			"token",
			"proof",
			"https://app.example.com/resource%3Fx=1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			target, err := url.Parse(tc.Target)
			require.NoError(t, err)

			r := oidc.NewDPoPResourceRequest(tc.Method, target, tc.Token, tc.Proof)

			assert.Equal(t, tc.Method, r.Method)
			assert.Equal(t, target.Host, r.Host)
			assert.Equal(t, tc.Proof, r.Header.Get(oidc.HeaderDPoP))
			assert.Equal(t, oidc.SchemeDPoP+" "+tc.Token, r.Header.Get("Authorization"))
			assert.Equal(t, tc.Expected, oauthelia2.RequestURL(r))
		})
	}
}

func TestValidateDPoPResourceAccess(t *testing.T) {
	testCases := []struct {
		Name       string
		Strategy   oauthelia2.DPoPStrategy
		JKT        string
		Alternates []string
		Expected   string
	}{
		{
			"ShouldRejectUnboundToken",
			&testDPoPStrategy{},
			"",
			nil,
			"The access token is not bound to a DPoP proof-of-possession key.",
		},
		{
			"ShouldRejectBoundTokenWithoutStrategy",
			nil,
			"abc",
			nil,
			"The access token is bound to a DPoP proof-of-possession key but this issuer is not configured to validate DPoP proofs.",
		},
		{
			"ShouldReturnValidationError",
			&testDPoPStrategy{err: oauthelia2.ErrInvalidDPoPProof.WithHint("The DPoP proof has already been used.")},
			"abc",
			nil,
			"The DPoP proof has already been used.",
		},
		{
			"ShouldReturnNonceChallengeWithoutMintingNonce",
			&testDPoPStrategy{err: oauthelia2.ErrUseDPoPNonce.WithHint("The DPoP proof is missing the required 'nonce' claim."), nonce: "fresh-nonce"},
			"abc",
			nil,
			"The DPoP proof is missing the required 'nonce' claim.",
		},
		{
			"ShouldSucceed",
			&testDPoPStrategy{},
			"abc",
			nil,
			"",
		},
		{
			"ShouldSucceedPassingAlternates",
			&testDPoPStrategy{},
			"abc",
			[]string{"https://auth.example.com/api/oidc/userinfo"},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			target, err := url.Parse("https://app.example.com/resource")
			require.NoError(t, err)

			r := oidc.NewDPoPResourceRequest(http.MethodGet, target, "token", "proof")

			err = oidc.ValidateDPoPResourceAccess(context.TODO(), tc.Strategy, r, "token", tc.JKT, false, tc.Alternates...)

			if strategy, ok := tc.Strategy.(*testDPoPStrategy); ok {
				assert.Equal(t, 0, strategy.minted)

				if tc.JKT != "" {
					assert.Equal(t, tc.Alternates, strategy.alternates)
				}
			}

			if tc.Expected == "" {
				assert.NoError(t, err)

				return
			}

			require.Error(t, err)
			assert.Equal(t, tc.Expected, oauthelia2.ErrorToRFC6749Error(err).HintField)
		})
	}
}

func TestNewDPoPNonceChallenge(t *testing.T) {
	testCases := []struct {
		Name     string
		Strategy oauthelia2.DPoPStrategy
		Err      error
		Expected string
		Minted   int
	}{
		{"ShouldNotMintWithoutError", &testDPoPStrategy{nonce: "fresh-nonce"}, nil, "", 0},
		{"ShouldNotMintWithoutStrategy", nil, oauthelia2.ErrUseDPoPNonce, "", 0},
		{"ShouldNotMintForOtherErrors", &testDPoPStrategy{nonce: "fresh-nonce"}, oauthelia2.ErrInvalidDPoPProof, "", 0},
		{"ShouldMintForNonceChallenge", &testDPoPStrategy{nonce: "fresh-nonce"}, oauthelia2.ErrUseDPoPNonce.WithHint("The DPoP proof is missing the required 'nonce' claim."), "fresh-nonce", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			nonce, err := oidc.NewDPoPNonceChallenge(context.TODO(), tc.Strategy, tc.Err)

			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, nonce)

			if strategy, ok := tc.Strategy.(*testDPoPStrategy); ok {
				assert.Equal(t, tc.Minted, strategy.minted)
			}
		})
	}
}

type testDPoPStrategy struct {
	err        error
	nonce      string
	minted     int
	alternates []string
}

func (s *testDPoPStrategy) ValidateDPoPProof(ctx context.Context, method, url, proof string, requireNonce bool, _ ...string) (parsed *oauthelia2.DPoPProof, err error) {
	return nil, s.err
}

func (s *testDPoPStrategy) NewDPoPNonce(ctx context.Context) (nonce string, err error) {
	s.minted++

	return s.nonce, nil
}

func (s *testDPoPStrategy) ValidateDPoPNonce(ctx context.Context, nonce string) (err error) {
	return nil
}

func (s *testDPoPStrategy) ValidateResourceAccess(ctx context.Context, r *http.Request, accessToken, boundJKT string, requireNonce bool, alternates ...string) (parsed *oauthelia2.DPoPProof, err error) {
	s.alternates = alternates

	if s.err != nil {
		return nil, s.err
	}

	return &oauthelia2.DPoPProof{Thumbprint: boundJKT}, nil
}

func TestRFC9449HandlerRefreshTokenBinding(t *testing.T) {
	testCases := []struct {
		name         string
		strict       bool
		clientStrict bool
		public       bool
		sameKey      bool
		keyBound     bool
		expected     string
	}{
		{name: "ShouldRebindConfidentialClientByDefault"},
		{name: "ShouldKeepBindingWithSameKey", strict: true, sameKey: true},
		{name: "ShouldNotRebindWhenProviderStrict", strict: true, expected: "The DPoP proof key does not match the key the grant is bound to."},
		{name: "ShouldNotRebindWhenClientStrict", clientStrict: true, expected: "The DPoP proof key does not match the key the grant is bound to."},
		{name: "ShouldNotRebindPublicClient", public: true, expected: "The DPoP proof key does not match the key the grant is bound to."},
		{name: "ShouldNotRebindKeyBoundGrant", keyBound: true, expected: "The DPoP proof key does not match the key the grant is bound to."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mocks.NewMockStorage(ctrl)

			client := schema.IdentityProvidersOpenIDConnectClient{
				ID:                            "dpop-client",
				Public:                        tc.public,
				DPoPStrictRefreshTokenBinding: tc.clientStrict,
				RedirectURIs:                  []string{"https://app.example.com/callback"},
				GrantTypes:                    []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
			}

			if !tc.public {
				client.Secret = tOpenIDConnectPlainTextClientSecret
			}

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret: "abcdefghijklmnopqrstuvwxyz123456",
						DPoP: schema.IdentityProvidersOpenIDConnectDPoP{
							Enabled:                   true,
							ClockSkew:                 time.Minute,
							ProofLifespan:             time.Minute,
							StrictRefreshTokenBinding: tc.strict,
						},
						Clients: []schema.IdentityProvidersOpenIDConnectClient{client},
					},
				},
			}

			provider := oidc.NewOpenIDConnectProvider(config, store, nil)

			ctx := context.Background()

			handlers := provider.GetTokenEndpointBindingHandlers(ctx)
			require.Len(t, handlers, 1)

			registered, err := provider.GetRegisteredClient(ctx, "dpop-client")
			require.NoError(t, err)

			original, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			require.NoError(t, err)

			presented := original

			if !tc.sameKey {
				presented, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)
			}

			bound, err := fjwt.ThumbprintJWK(&jose.JSONWebKey{Key: original.Public()})
			require.NoError(t, err)

			thumbprint, err := fjwt.ThumbprintJWK(&jose.JSONWebKey{Key: presented.Public()})
			require.NoError(t, err)

			session := oidc.NewSession()
			session.SetDPoPJWKThumbprint(bound)
			session.SetOIDCKeyBindingGranted(tc.keyBound)

			request := &oauthelia2.AccessRequest{
				GrantTypes: oauthelia2.Arguments{oidc.GrantTypeRefreshToken},
				Request: oauthelia2.Request{
					Client:  registered,
					Session: session,
					Form:    url.Values{},
				},
			}

			r, err := http.NewRequest(http.MethodPost, "https://auth.example.com/api/oidc/token", nil)
			require.NoError(t, err)

			r.Header.Set("X-Forwarded-Proto", "https")
			r.Header.Set(oidc.HeaderDPoP, newTestDPoPTokenProof(t, presented, http.MethodPost, "https://auth.example.com/api/oidc/token"))

			store.EXPECT().CheckAndSetOAuth2DPoPProofUsed(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

			err = handlers[0].BindAccessRequest(context.WithValue(ctx, oauthelia2.RequestContextKey, r), request)

			if tc.expected == "" {
				assert.NoError(t, err)
				assert.Equal(t, thumbprint, session.GetDPoPJWKThumbprint())

				return
			}

			require.Error(t, err)
			assert.Equal(t, tc.expected, oauthelia2.ErrorToRFC6749Error(err).HintField)
			assert.Equal(t, bound, session.GetDPoPJWKThumbprint())
		})
	}
}

func newTestDPoPTokenProof(t *testing.T, key *ecdsa.PrivateKey, method, target string) (proof string) {
	t.Helper()

	return newTestDPoPTokenProofWithCodeHash(t, key, method, target, "")
}

func newTestDPoPTokenProofWithCodeHash(t *testing.T, key *ecdsa.PrivateKey, method, target, codeHash string) (proof string) {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: key},
		(&jose.SignerOptions{EmbedJWK: true}).WithType(fjwt.JSONWebTokenTypeDPoP),
	)
	require.NoError(t, err)

	claims := map[string]any{
		"jti": uuid.Must(uuid.NewRandom()).String(),
		"htm": method,
		"htu": target,
		"iat": time.Now().Unix(),
	}

	if codeHash != "" {
		claims["c_s256"] = codeHash
	}

	payload, err := json.Marshal(claims)
	require.NoError(t, err)

	object, err := signer.Sign(payload)
	require.NoError(t, err)

	proof, err = object.CompactSerialize()
	require.NoError(t, err)

	return proof
}

func TestDPoPDisabledTokenBindingHandler(t *testing.T) {
	const jkt = "D6Nq0uHi1xL9fbLBu6xVGvKtOsBqiOxfHy_hOZlLzHM"

	testCases := []struct {
		name      string
		bound     string
		requested string
		err       string
	}{
		{
			"ShouldAllowUnboundGrant",
			"",
			"",
			"",
		},
		{
			"ShouldRejectBoundGrant",
			jkt,
			"",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The grant is bound to a DPoP proof-of-possession key which can't be verified as DPoP is disabled. The grant records a DPoP binding but DPoP is disabled, so possession of the bound key can't be proven and the grant must not be redeemed.",
		},
		{
			"ShouldRejectRequestedBindingGrant",
			"",
			jkt,
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The grant is bound to a DPoP proof-of-possession key which can't be verified as DPoP is disabled. The grant records a DPoP binding but DPoP is disabled, so possession of the bound key can't be proven and the grant must not be redeemed.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			config := &oidc.Config{}

			config.LoadHandlers(&oidc.Store{})

			handlers := config.GetTokenEndpointBindingHandlers(ctx)
			require.Len(t, handlers, 1)

			session := oidc.NewSession()
			session.SetDPoPJWKThumbprint(tc.bound)
			session.SetRequestedDPoPJWKThumbprint(tc.requested)

			request := oauthelia2.NewAccessRequest(session)

			err := handlers[0].BindAccessRequest(ctx, request)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, oauthelia2.ErrorToDebugRFC6749Error(err), tc.err)
			}

			assert.NoError(t, handlers[0].PopulateBoundTokenEndpointResponse(ctx, request, oauthelia2.NewAccessResponse()))
		})
	}
}
