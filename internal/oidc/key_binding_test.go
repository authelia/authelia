// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"authelia.com/provider/jose"
	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/oidckb"
	fjwt "authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOpenIDConnectKeyBindingIDTokenIssuance(t *testing.T) {
	const (
		code   = "authelia_ac_example-authorization-code"
		target = "https://auth.example.com/api/oidc/token"
	)

	testCases := []struct {
		name        string
		disabled    bool
		refresh     bool
		requested   bool
		granted     bool
		otherKey    bool
		codeHash    string
		expectedErr string
		expectedCnf bool
	}{
		{name: "ShouldIssueKeyBoundIDToken", requested: true, granted: true, codeHash: oidckb.CodeHash(code), expectedCnf: true},
		{name: "ShouldNotBindWithoutDPoPJKT", granted: false},
		{name: "ShouldNotBindWhenBoundKeyNotGranted", requested: true, codeHash: oidckb.CodeHash(code)},
		{name: "ShouldNotBindWhenKeyBindingDisabled", disabled: true, requested: true, granted: true, codeHash: oidckb.CodeHash(code)},
		{name: "ShouldRejectMissingCodeHash", requested: true, granted: true, expectedErr: "The DPoP proof is missing the 'c_s256' claim, which is required because the 'bound_key' scope was granted."},
		{name: "ShouldRejectMismatchedCodeHash", requested: true, granted: true, codeHash: oidckb.CodeHash("other"), expectedErr: "The DPoP proof 'c_s256' claim does not match the presented code."},
		{name: "ShouldRejectProofForAnotherKey", requested: true, granted: true, otherKey: true, codeHash: oidckb.CodeHash(code), expectedErr: "The DPoP proof key does not match the key the grant is bound to."},
		{name: "ShouldKeepConfirmationOnRefresh", refresh: true, requested: true, granted: true, expectedCnf: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mocks.NewMockStorage(ctrl)

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret: "abcdefghijklmnopqrstuvwxyz123456",
						JSONWebKeys: []schema.JWK{
							{KeyID: "rs256", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: x509PrivateKeyRSA2048},
						},
						DPoP: schema.IdentityProvidersOpenIDConnectDPoP{
							Enabled:       true,
							KeyBinding:    !tc.disabled,
							ClockSkew:     time.Minute,
							ProofLifespan: time.Minute,
						},
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:           "kb-client",
								Secret:       tOpenIDConnectPlainTextClientSecret,
								RedirectURIs: []string{"https://app.example.com/callback"},
								Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeBoundKey, oidc.ScopeOfflineAccess},
								GrantTypes:   []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
							},
						},
					},
				},
			}

			provider := oidc.NewOpenIDConnectProvider(config, store, nil)

			ctx := context.Background()

			handlers := provider.GetTokenEndpointBindingHandlers(ctx)

			if tc.disabled {
				require.Len(t, handlers, 1)
			} else {
				require.Len(t, handlers, 2)
			}

			client, err := provider.GetRegisteredClient(ctx, "kb-client")
			require.NoError(t, err)

			key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			require.NoError(t, err)

			jkt, err := fjwt.ThumbprintJWK(&jose.JSONWebKey{Key: key.Public()})
			require.NoError(t, err)

			presented := key

			if tc.otherKey {
				presented, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)
			}

			session := oidc.NewSessionWithRequestedAt(time.Now())
			session.Claims.Subject = "john"
			session.Claims.Issuer = "https://auth.example.com"
			session.Claims.AuthTime = fjwt.NewNumericDate(time.Now().Add(-time.Minute))
			session.SetOIDCKeyBindingGranted(tc.granted)

			if tc.requested {
				session.SetRequestedDPoPJWKThumbprint(jkt)
			}

			form := url.Values{}

			grant := oidc.GrantTypeAuthorizationCode

			if tc.refresh {
				grant = oidc.GrantTypeRefreshToken

				raw, err := (&jose.JSONWebKey{Key: key.Public()}).MarshalJSON()
				require.NoError(t, err)

				session.SetDPoPJWKThumbprint(jkt)
				session.SetDPoPPublicKeyJWK(raw)
			} else {
				form.Set("code", code)
			}

			request := &oauthelia2.AccessRequest{
				GrantTypes: oauthelia2.Arguments{grant},
				Request: oauthelia2.Request{
					RequestedAt:  time.Now(),
					Client:       client,
					Session:      session,
					Form:         form,
					GrantedScope: oauthelia2.Arguments{oidc.ScopeOpenID, oidc.ScopeBoundKey},
				},
			}

			r, err := http.NewRequest(http.MethodPost, target, nil)
			require.NoError(t, err)

			r.Header.Set("X-Forwarded-Proto", "https")
			r.Header.Set(oidc.HeaderDPoP, newTestDPoPTokenProofWithCodeHash(t, presented, http.MethodPost, target, tc.codeHash))

			store.EXPECT().CheckAndSetOAuth2DPoPProofUsed(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

			ctx = context.WithValue(ctx, oauthelia2.RequestContextKey, r)
			ctx = context.WithValue(ctx, oauthelia2.AccessRequestContextKey, request)
			ctx = context.WithValue(ctx, oauthelia2.DPoPProofContextKey, &oauthelia2.DPoPProofHolder{})

			for _, handler := range handlers {
				if err = handler.BindAccessRequest(ctx, request); err != nil {
					break
				}
			}

			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr, oauthelia2.ErrorToRFC6749Error(err).HintField)

				return
			}

			require.NoError(t, err)

			token, err := provider.Strategy.OpenID.GenerateIDToken(ctx, time.Hour, request)
			require.NoError(t, err)

			header, claims := decodeTestJWT(t, token)

			if !tc.expectedCnf {
				assert.NotContains(t, claims, "cnf")
				assert.NotEqual(t, fjwt.JSONWebTokenTypeDPoPIDToken, header["typ"])

				return
			}

			assert.Equal(t, fjwt.JSONWebTokenTypeDPoPIDToken, header["typ"])

			cnf, ok := claims["cnf"].(map[string]any)
			require.True(t, ok)

			raw, err := json.Marshal(cnf["jwk"])
			require.NoError(t, err)

			var confirmed jose.JSONWebKey

			require.NoError(t, confirmed.UnmarshalJSON(raw))

			thumbprint, err := fjwt.ThumbprintJWK(&confirmed)
			require.NoError(t, err)

			assert.Equal(t, jkt, thumbprint)
		})
	}
}

func decodeTestJWT(t *testing.T, token string) (header, claims map[string]any) {
	t.Helper()

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)

	for i, target := range []*map[string]any{&header, &claims} {
		raw, err := base64.RawURLEncoding.DecodeString(parts[i])
		require.NoError(t, err)

		require.NoError(t, json.Unmarshal(raw, target))
	}

	return header, claims
}

func TestDPoPBoundJWTAccessTokenConfirmation(t *testing.T) {
	testCases := []struct {
		name     string
		jkt      string
		expected bool
	}{
		{"ShouldIncludeConfirmationForBoundToken", "0ZcOCORZNYy-DWpqq30jZyJGHTN0d2HglBV3uiguA4I", true},
		{"ShouldOmitConfirmationForUnboundToken", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret: "abcdefghijklmnopqrstuvwxyz123456",
						JSONWebKeys: []schema.JWK{
							{KeyID: "rs256", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: x509PrivateKeyRSA2048},
						},
						DPoP: schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true},
						Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
							JWTResponseAccessTokens: true,
						},
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:                             "jwt-client",
								Secret:                         tOpenIDConnectPlainTextClientSecret,
								GrantTypes:                     []string{oidc.GrantTypeClientCredentials},
								AccessTokenSignedResponseAlg:   oidc.SigningAlgRSAUsingSHA256,
								AccessTokenSignedResponseKeyID: "rs256",
							},
						},
					},
				},
			}

			provider := oidc.NewOpenIDConnectProvider(config, mocks.NewMockStorage(ctrl), nil)

			ctx := context.Background()

			client, err := provider.GetRegisteredClient(ctx, "jwt-client")
			require.NoError(t, err)

			session := oidc.NewSessionWithRequestedAt(time.Now())
			session.ClientID = "jwt-client"
			session.ClientCredentials = true
			session.Claims.Subject = "jwt-client"
			session.Claims.Issuer = "https://auth.example.com"
			session.SetExpiresAt(oauthelia2.AccessToken, time.Now().Add(time.Hour))
			session.SetDPoPJWKThumbprint(tc.jkt)

			request := &oauthelia2.AccessRequest{
				GrantTypes: oauthelia2.Arguments{oidc.GrantTypeClientCredentials},
				Request: oauthelia2.Request{
					RequestedAt:     time.Now(),
					Client:          client,
					Session:         session,
					Form:            url.Values{},
					GrantedAudience: oauthelia2.Arguments{"https://api.example.com"},
				},
			}

			token, _, err := provider.Strategy.Core.GenerateAccessToken(ctx, request)
			require.NoError(t, err)

			header, claims := decodeTestJWT(t, token)

			assert.Equal(t, "at+jwt", header["typ"])

			if !tc.expected {
				assert.NotContains(t, claims, "cnf")

				return
			}

			assert.Equal(t, map[string]any{"jkt": tc.jkt}, claims["cnf"])
		})
	}
}
