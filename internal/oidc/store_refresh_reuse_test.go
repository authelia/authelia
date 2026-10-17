// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestRefreshTokenReuseRevokesGrant(t *testing.T) {
	testCases := []struct {
		name   string
		replay func(t *testing.T, provider *oidc.OpenIDConnectProvider, original, rotated refreshTokenGrant)
	}{
		{
			name: "ShouldRevokeGrantOnReplay",
			replay: func(t *testing.T, provider *oidc.OpenIDConnectProvider, original, rotated refreshTokenGrant) {
				_, err := doRefreshTokenGrant(provider, original.refresh)

				assert.ErrorIs(t, err, oauthelia2.ErrInvalidGrant)
			},
		},
		{
			name: "ShouldRevokeGrantOnReplayWhenAccessTokenAlreadyRevoked",
			replay: func(t *testing.T, provider *oidc.OpenIDConnectProvider, original, rotated refreshTokenGrant) {
				require.NoError(t, provider.RevokeAccessToken(context.Background(), rotated.requestID))

				_, err := doRefreshTokenGrant(provider, original.refresh)

				assert.ErrorIs(t, err, oauthelia2.ErrInvalidGrant)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newRefreshTokenReuseProvider(t)

			original := seedRefreshTokenGrant(t, provider)

			rotated, err := doRefreshTokenGrant(provider, original.refresh)
			require.NoError(t, err)
			require.NotEmpty(t, rotated.refresh)
			require.NotEqual(t, original.refresh, rotated.refresh)

			tc.replay(t, provider, original, rotated)

			_, err = doRefreshTokenGrant(provider, rotated.refresh)
			assert.ErrorIs(t, err, oauthelia2.ErrInvalidGrant)

			assertAccessTokenRevoked(t, provider, rotated.access)
		})
	}
}

func TestRefreshTokenReplayCaughtAtPopulationRevokesGrant(t *testing.T) {
	provider := newRefreshTokenReuseProvider(t)

	original := seedRefreshTokenGrant(t, provider)

	first, err := newRefreshTokenAccessRequest(provider, original.refresh)
	require.NoError(t, err)

	second, err := newRefreshTokenAccessRequest(provider, original.refresh)
	require.NoError(t, err)

	response, err := provider.NewAccessResponse(context.Background(), first)
	require.NoError(t, err)

	rotated := refreshTokenGrantFromResponse(first, response)

	_, err = provider.NewAccessResponse(context.Background(), second)
	assert.ErrorIs(t, err, oauthelia2.ErrInvalidGrant)

	_, err = doRefreshTokenGrant(provider, rotated.refresh)
	assert.ErrorIs(t, err, oauthelia2.ErrInvalidGrant)

	assertAccessTokenRevoked(t, provider, rotated.access)
}

type refreshTokenGrant struct {
	requestID string
	access    string
	refresh   string
}

const (
	refreshReuseClientID = "refresh-reuse-client"
	refreshReuseSubject  = "b7c3a4f2-3c1e-4a8e-9f6d-0a2b1c3d4e5f"
)

func newRefreshTokenReuseProvider(t *testing.T) *oidc.OpenIDConnectProvider {
	t.Helper()

	store, err := storage.NewSQLiteProvider(&schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	})

	require.NoError(t, err)
	require.NoError(t, store.StartupCheck())

	provider := oidc.NewOpenIDConnectProvider(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				HMACSecret:             badhmac,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                      refreshReuseClientID,
						Secret:                  tOpenIDConnectPlainTextClientSecret,
						AuthorizationPolicy:     onefactor,
						TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
						GrantTypes:              []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
						ResponseTypes:           []string{oidc.ResponseTypeAuthorizationCodeFlow},
						Scopes:                  []string{oidc.ScopeOfflineAccess},
						RedirectURIs:            []string{examplecom},
					},
				},
			},
		},
	}, store, nil)

	require.NotNil(t, provider)

	return provider
}

func seedRefreshTokenGrant(t *testing.T, provider *oidc.OpenIDConnectProvider) refreshTokenGrant {
	t.Helper()

	ctx := context.Background()

	client, err := provider.GetClient(ctx, refreshReuseClientID)
	require.NoError(t, err)

	session := oidc.NewSession()
	session.Subject = refreshReuseSubject
	session.ChallengeID = model.MustNullUUID(model.NewRandomNullUUID())
	session.SetExpiresAt(oauthelia2.AccessToken, time.Now().UTC().Add(time.Hour))
	session.SetExpiresAt(oauthelia2.RefreshToken, time.Now().UTC().Add(time.Hour))

	request := oauthelia2.NewRequest()
	request.SetID("f4b6d1b0-2f9a-4a7e-9b2c-6d8e0f1a2b3c")
	request.Client = client
	request.Session = session
	request.SetRequestedScopes(oauthelia2.Arguments{oidc.ScopeOfflineAccess})
	request.GrantScope(oidc.ScopeOfflineAccess)

	access, accessSignature, err := provider.Strategy.Core.GenerateAccessToken(ctx, request)
	require.NoError(t, err)

	refresh, refreshSignature, err := provider.Strategy.Core.GenerateRefreshToken(ctx, request)
	require.NoError(t, err)

	require.NoError(t, provider.CreateAccessTokenSession(ctx, accessSignature, request))
	require.NoError(t, provider.CreateRefreshTokenSession(ctx, refreshSignature, accessSignature, request))

	return refreshTokenGrant{requestID: request.GetID(), access: access, refresh: refresh}
}

func newRefreshTokenAccessRequest(provider *oidc.OpenIDConnectProvider, refresh string) (oauthelia2.AccessRequester, error) {
	form := url.Values{
		"grant_type":               []string{oidc.GrantTypeRefreshToken},
		"refresh_token":            []string{refresh},
		oidc.FormParameterClientID: []string{refreshReuseClientID},
		"client_secret":            []string{"client-secret"},
	}

	r := httptest.NewRequest(http.MethodPost, "https://auth.example.com/api/oidc/token", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return provider.NewAccessRequest(context.Background(), r, oidc.NewSession())
}

func doRefreshTokenGrant(provider *oidc.OpenIDConnectProvider, refresh string) (grant refreshTokenGrant, err error) {
	var (
		requester oauthelia2.AccessRequester
		response  oauthelia2.AccessResponder
	)

	if requester, err = newRefreshTokenAccessRequest(provider, refresh); err != nil {
		return grant, err
	}

	if response, err = provider.NewAccessResponse(context.Background(), requester); err != nil {
		return grant, err
	}

	return refreshTokenGrantFromResponse(requester, response), nil
}

func refreshTokenGrantFromResponse(requester oauthelia2.AccessRequester, response oauthelia2.AccessResponder) refreshTokenGrant {
	refresh, _ := response.GetExtra("refresh_token").(string)

	return refreshTokenGrant{requestID: requester.GetID(), access: response.GetAccessToken(), refresh: refresh}
}

func assertAccessTokenRevoked(t *testing.T, provider *oidc.OpenIDConnectProvider, access string) {
	t.Helper()

	signature := provider.Strategy.Core.AccessTokenSignature(context.Background(), access)

	_, err := provider.GetAccessTokenSession(context.Background(), signature, oidc.NewSession())

	assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
}
