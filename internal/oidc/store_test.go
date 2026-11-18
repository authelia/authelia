// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"
	hoauth2 "authelia.com/provider/oauth2/handler/oauth2"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestOpenIDConnectStore_GetInternalClient(t *testing.T) {
	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  myclient,
						Name:                myclientname,
						AuthorizationPolicy: onefactor,
						Scopes:              []string{oidc.ScopeOpenID, oidc.ScopeProfile},
						Secret:              tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}, nil, nil)

	client, err := s.GetClient(context.Background(), "myinvalidclient")
	assert.EqualError(t, err, "invalid_client")
	assert.Nil(t, client)

	client, err = s.GetClient(context.Background(), myclient)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, myclient, client.GetID())
}

func TestOpenIDConnectStore_GetInternalClient_ValidClient(t *testing.T) {
	ctx := context.Background()

	id := myclient

	c1 := schema.IdentityProvidersOpenIDConnectClient{
		ID:                  id,
		Name:                myclientname,
		AuthorizationPolicy: onefactor,
		Scopes:              []string{oidc.ScopeOpenID, oidc.ScopeProfile},
		Secret:              tOpenIDConnectPlainTextClientSecret,
	}

	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients:                []schema.IdentityProvidersOpenIDConnectClient{c1},
			},
		},
	}, nil, nil)

	client, err := s.GetRegisteredClient(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, id, client.GetID())
	assert.Equal(t, myclientname, client.GetName())
	assert.Equal(t, oauthelia2.Arguments(c1.Scopes), client.GetScopes())
	assert.Equal(t, oauthelia2.Arguments([]string{oidc.GrantTypeAuthorizationCode}), client.GetGrantTypes())
	assert.Equal(t, oauthelia2.Arguments([]string{oidc.ResponseTypeAuthorizationCodeFlow}), client.GetResponseTypes())
	assert.Equal(t, []string(nil), client.GetRedirectURIs())
	assert.Equal(t, authorization.OneFactor, client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{}))
	assert.Equal(t, "$plaintext$client-secret", client.GetClientSecret().(*oidc.ClientSecretDigest).Encode())
}

func TestOpenIDConnectStore_GetInternalClient_InvalidClient(t *testing.T) {
	ctx := context.Background()

	c1 := schema.IdentityProvidersOpenIDConnectClient{
		ID:                  myclient,
		Name:                myclientname,
		AuthorizationPolicy: onefactor,
		Scopes:              []string{oidc.ScopeOpenID, oidc.ScopeProfile},
		Secret:              tOpenIDConnectPlainTextClientSecret,
	}

	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients:                []schema.IdentityProvidersOpenIDConnectClient{c1},
			},
		},
	}, nil, nil)

	client, err := s.GetRegisteredClient(ctx, "another-client")
	assert.Nil(t, client)
	assert.EqualError(t, err, "invalid_client")
}

func TestOpenIDConnectStore_IsValidClientID(t *testing.T) {
	ctx := context.Background()

	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  myclient,
						Name:                myclientname,
						AuthorizationPolicy: onefactor,
						Scopes:              []string{oidc.ScopeOpenID, oidc.ScopeProfile},
						Secret:              tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}, nil, nil)

	validClient := s.IsValidClientID(ctx, myclient)
	invalidClient := s.IsValidClientID(ctx, "myinvalidclient")

	assert.True(t, validClient)
	assert.False(t, invalidClient)
}

func TestOAuth2SessionGrantedResourceRoundTripsThroughStore(t *testing.T) {
	ctx := context.Background()

	provider, err := storage.NewSQLiteProvider(&schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	})

	require.NoError(t, err)
	require.NoError(t, provider.StartupCheck())

	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  myclient,
						Name:                myclientname,
						AuthorizationPolicy: onefactor,
						Scopes:              []string{oidc.ScopeOpenID, oidc.ScopeProfile},
						Secret:              tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}, provider, nil)

	requester := &oauthelia2.Request{
		ID:                "req-resource-e2e",
		Client:            &oidc.RegisteredClient{ID: myclient},
		RequestedScope:    oauthelia2.Arguments{oidc.ScopeOpenID},
		RequestedAudience: oauthelia2.Arguments{"aud-a"},
		Session:           &oidc.Session{},
	}

	consent := &model.OAuth2ConsentSession{
		GrantedScopes:   model.StringSlicePipeDelimited{oidc.ScopeOpenID},
		GrantedAudience: model.StringSlicePipeDelimited{"aud-a"},
		GrantedResource: model.StringSlicePipeDelimited{"https://api.example.com"},
	}

	oidc.GrantScopeAudienceConsent(requester, consent)

	require.NoError(t, s.CreateAccessTokenSession(ctx, "sig-resource-e2e", requester))

	loaded, err := s.GetAccessTokenSession(ctx, "sig-resource-e2e", &oidc.Session{})

	require.NoError(t, err)
	assert.Equal(t, oauthelia2.Arguments{oidc.ScopeOpenID}, loaded.GetGrantedScopes())
	assert.Equal(t, oauthelia2.Arguments{"aud-a"}, loaded.GetGrantedAudience())
	assert.Equal(t, oauthelia2.Arguments{"https://api.example.com"}, loaded.GetGrantedResource())
}

func TestOAuth2RefreshTokenSessionExpiryDeadlineRoundTripsThroughStore(t *testing.T) {
	ctx := context.Background()

	provider, err := storage.NewSQLiteProvider(&schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	})

	require.NoError(t, err)
	require.NoError(t, provider.StartupCheck())

	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  myclient,
						Name:                myclientname,
						AuthorizationPolicy: onefactor,
						Scopes:              []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
						Secret:              tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}, provider, nil)

	deadline := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)

	subject, err := s.GetSubject(ctx, "", "john")

	require.NoError(t, err)

	session := oidc.NewSession()
	session.ChallengeID = model.MustNullUUID(model.NewRandomNullUUID())
	session.Subject = subject.String()

	var setter oauthelia2.ExpiryDeadlineSetterSession = session

	setter.SetExpiryDeadline(deadline)

	requester := &oauthelia2.Request{
		ID:             "req-deadline",
		Client:         &oidc.RegisteredClient{ID: myclient},
		RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
		GrantedScope:   oauthelia2.Arguments{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
		Session:        session,
	}

	require.NoError(t, s.CreateRefreshTokenSession(ctx, "sig-deadline", "", requester))

	loaded, err := s.GetRefreshTokenSession(ctx, "sig-deadline", oidc.NewSession())

	require.NoError(t, err)

	bounded, ok := loaded.GetSession().(oauthelia2.ExpiryDeadlineSession)

	require.True(t, ok)
	assert.True(t, deadline.Equal(bounded.GetExpiryDeadline()))

	cloned, ok := loaded.GetSession().Clone().(oauthelia2.ExpiryDeadlineSession)

	require.True(t, ok)
	assert.True(t, deadline.Equal(cloned.GetExpiryDeadline()))
}

func TestOAuth2RefreshTokenGracePeriodRotationThroughStore(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)

	provider, err := storage.NewSQLiteProvider(&schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	})

	require.NoError(t, err)
	require.NoError(t, provider.StartupCheck())

	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  myclient,
						Name:                myclientname,
						AuthorizationPolicy: onefactor,
						Scopes:              []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
						Secret:              tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}, provider, nil)

	fixed := clock.NewFixed(now)
	ctx := &TestContext{Context: context.Background(), Clock: fixed}

	subject, err := s.GetSubject(ctx, "", "john")

	require.NoError(t, err)

	newRequester := func() oauthelia2.Requester {
		session := oidc.NewSession()
		session.ChallengeID = model.MustNullUUID(model.NewRandomNullUUID())
		session.Subject = subject.String()

		return &oauthelia2.Request{
			ID:             "req-grace",
			Client:         &oidc.RegisteredClient{ID: myclient},
			RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
			GrantedScope:   oauthelia2.Arguments{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
			Session:        session,
		}
	}

	require.NoError(t, s.CreateAccessTokenSession(ctx, "at-1", newRequester()))
	require.NoError(t, s.CreateRefreshTokenSession(ctx, "rt-1", "at-1", newRequester()))

	deadline := now.Add(10 * time.Second)

	require.NoError(t, s.RotateRefreshTokenWithGracePeriod(ctx, "req-grace", "rt-1", deadline))

	_, err = s.GetAccessTokenSession(ctx, "at-1", oidc.NewSession())
	assert.ErrorIs(t, err, oauthelia2.ErrNotFound)

	_, err = s.GetRefreshTokenSession(ctx, "rt-1", oidc.NewSession())
	require.NoError(t, err)

	require.NoError(t, s.CreateAccessTokenSession(ctx, "at-2", newRequester()))
	require.NoError(t, s.CreateRefreshTokenSession(ctx, "rt-2", "at-2", newRequester()))

	fixed.Set(now.Add(5 * time.Second))

	require.NoError(t, s.RotateRefreshTokenWithGracePeriod(ctx, "req-grace", "rt-1", now.Add(15*time.Second)))

	state, err := provider.LoadOAuth2RefreshTokenSessionState(ctx, "rt-1")
	require.NoError(t, err)
	require.True(t, state.GraceDeadline.Valid)
	assert.True(t, deadline.Equal(state.GraceDeadline.Time))
	assert.True(t, state.Active)

	_, err = s.GetAccessTokenSession(ctx, "at-2", oidc.NewSession())
	require.NoError(t, err)

	_, err = s.GetRefreshTokenSession(ctx, "rt-2", oidc.NewSession())
	require.NoError(t, err)

	fixed.Set(deadline)

	r, err := s.GetRefreshTokenSession(ctx, "rt-1", oidc.NewSession())
	assert.ErrorIs(t, err, oauthelia2.ErrInactiveToken)
	require.NotNil(t, r)
	assert.Equal(t, "req-grace", r.GetID())

	assert.ErrorIs(t, s.RotateRefreshTokenWithGracePeriod(ctx, "req-grace", "rt-1", deadline.Add(10*time.Second)), oauthelia2.ErrInactiveToken)
	assert.ErrorIs(t, s.RotateRefreshToken(ctx, "req-grace", "rt-1"), oauthelia2.ErrInactiveToken)

	require.NoError(t, s.RotateRefreshTokenWithGracePeriod(ctx, "req-grace", "rt-2", deadline.Add(10*time.Second)))

	require.NoError(t, s.RevokeRefreshToken(ctx, "req-grace"))

	_, err = s.GetRefreshTokenSession(ctx, "rt-2", oidc.NewSession())
	assert.ErrorIs(t, err, oauthelia2.ErrInactiveToken)

	assert.ErrorIs(t, s.RotateRefreshTokenWithGracePeriod(ctx, "req-grace", "rt-2", deadline.Add(10*time.Second)), oauthelia2.ErrInactiveToken)
	assert.ErrorIs(t, s.RotateRefreshTokenWithGracePeriod(ctx, "req-grace", "rt-missing", deadline), oauthelia2.ErrNotFound)
}

func TestOAuth2RefreshTokenGracePeriodForkRevocation(t *testing.T) {
	const (
		tokenAT0 = "at0"
		tokenAT1 = "at1"
		tokenAT2 = "at2"
		tokenRT0 = "rt0"
		tokenRT1 = "rt1"
		tokenRT2 = "rt2"
	)

	testCases := []struct {
		name     string
		explicit bool
		expired  bool
		token    string
		hint     oauthelia2.TokenType
		inactive []string
		active   []string
	}{
		{
			name:     "ShouldRevokeBothForksWhenForkAccessTokenRevoked",
			token:    tokenAT1,
			hint:     oauthelia2.AccessToken,
			inactive: []string{tokenAT0, tokenAT1, tokenAT2, tokenRT0, tokenRT1, tokenRT2},
		},
		{
			name:     "ShouldRevokeBothForksWhenSiblingForkAccessTokenRevoked",
			token:    tokenAT2,
			hint:     oauthelia2.AccessToken,
			inactive: []string{tokenAT0, tokenAT1, tokenAT2, tokenRT0, tokenRT1, tokenRT2},
		},
		{
			name:     "ShouldRevokeOnlyAccessTokensOfBothForksWhenAccessTokenRevokedExplicitly",
			explicit: true,
			token:    tokenAT1,
			hint:     oauthelia2.AccessToken,
			inactive: []string{tokenAT0, tokenAT1, tokenAT2},
			active:   []string{tokenRT0, tokenRT1, tokenRT2},
		},
		{
			name:     "ShouldRevokeBothForksWhenForkRefreshTokenRevoked",
			token:    tokenRT1,
			hint:     oauthelia2.RefreshToken,
			inactive: []string{tokenAT0, tokenAT1, tokenAT2, tokenRT0, tokenRT1, tokenRT2},
		},
		{
			name:     "ShouldRevokeBothForksWhenSiblingForkRefreshTokenRevoked",
			token:    tokenRT2,
			hint:     oauthelia2.RefreshToken,
			inactive: []string{tokenAT0, tokenAT1, tokenAT2, tokenRT0, tokenRT1, tokenRT2},
		},
		{
			name:     "ShouldRevokeBothForksWhenForkRefreshTokenRevokedExplicitly",
			explicit: true,
			token:    tokenRT1,
			hint:     oauthelia2.RefreshToken,
			inactive: []string{tokenAT0, tokenAT1, tokenAT2, tokenRT0, tokenRT1, tokenRT2},
		},
		{
			name:     "ShouldRevokeBothForksWhenGraceRefreshTokenRevoked",
			token:    tokenRT0,
			hint:     oauthelia2.RefreshToken,
			inactive: []string{tokenAT0, tokenAT1, tokenAT2, tokenRT0, tokenRT1, tokenRT2},
		},
		{
			name:     "ShouldRevokeBothForksWhenGraceRefreshTokenRevokedAfterDeadline",
			expired:  true,
			token:    tokenRT0,
			hint:     oauthelia2.RefreshToken,
			inactive: []string{tokenAT0, tokenAT1, tokenAT2, tokenRT0, tokenRT1, tokenRT2},
		},
		{
			name:     "ShouldRevokeBothForksWhenGraceRefreshTokenRevokedWithAccessTokenHint",
			token:    tokenRT0,
			hint:     oauthelia2.AccessToken,
			inactive: []string{tokenAT0, tokenAT1, tokenAT2, tokenRT0, tokenRT1, tokenRT2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now().UTC()

			provider, err := storage.NewSQLiteProvider(&schema.Configuration{
				Storage: schema.Storage{
					EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
					Local: &schema.StorageLocal{
						Path: filepath.Join(t.TempDir(), "db.sqlite3"),
					},
				},
			})

			require.NoError(t, err)
			require.NoError(t, provider.StartupCheck())

			s := oidc.NewStore(&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						IssuerCertificateChain: schema.X509CertificateChain{},
						IssuerPrivateKey:       x509PrivateKeyRSA2048,
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:                  myclient,
								Name:                myclientname,
								AuthorizationPolicy: onefactor,
								Scopes:              []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
								GrantTypes:          []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
								Secret:              tOpenIDConnectPlainTextClientSecret,
							},
						},
					},
				},
			}, provider, nil)

			fixed := clock.NewFixed(now)
			ctx := &TestContext{Context: context.Background(), Clock: fixed}

			config := &oauthelia2.Config{
				GlobalSecret:                []byte("a-test-global-secret-of-32-bytes"),
				AccessTokenLifespan:         time.Hour,
				RefreshTokenLifespan:        time.Hour,
				RefreshTokenGracePeriod:     time.Minute,
				RefreshTokenScopes:          []string{oidc.ScopeOfflineAccess},
				ScopeStrategy:               oauthelia2.ExactScopeStrategy,
				AudienceStrategy:            oauthelia2.DefaultAudienceStrategy,
				RevokeRefreshTokensExplicit: tc.explicit,
			}

			strategy := hoauth2.NewHMACCoreStrategy(config, "authelia_%s_")

			grant := &hoauth2.RefreshTokenGrantHandler{
				AccessTokenStrategy:    strategy,
				RefreshTokenStrategy:   strategy,
				TokenRevocationStorage: s,
				Config:                 config,
			}

			revocation := &hoauth2.TokenRevocationHandler{
				TokenRevocationStorage: s,
				AccessTokenStrategy:    strategy,
				RefreshTokenStrategy:   strategy,
				Config:                 config,
			}

			client, err := s.GetClient(ctx, myclient)
			require.NoError(t, err)

			subject, err := s.GetSubject(ctx, "", "john")
			require.NoError(t, err)

			session := oidc.NewSession()
			session.ChallengeID = model.MustNullUUID(model.NewRandomNullUUID())
			session.Subject = subject.String()
			session.SetExpiresAt(oauthelia2.AccessToken, now.Add(time.Hour))
			session.SetExpiresAt(oauthelia2.RefreshToken, now.Add(time.Hour))

			original := &oauthelia2.Request{
				ID:             "req-fork",
				Client:         client,
				RequestedAt:    now,
				RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
				GrantedScope:   oauthelia2.Arguments{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
				Session:        session,
			}

			tokens := map[string]string{}
			signatures := map[string]string{}

			var accessSignature, refreshSignature string

			tokens[tokenAT0], accessSignature, err = strategy.GenerateAccessToken(ctx, original)
			require.NoError(t, err)

			tokens[tokenRT0], refreshSignature, err = strategy.GenerateRefreshToken(ctx, original)
			require.NoError(t, err)

			signatures[tokenAT0], signatures[tokenRT0] = accessSignature, refreshSignature

			require.NoError(t, s.CreateAccessTokenSession(ctx, accessSignature, original))
			require.NoError(t, s.CreateRefreshTokenSession(ctx, refreshSignature, accessSignature, original))

			refresh := func(at, rt string) {
				request := oauthelia2.NewAccessRequest(oidc.NewSession())
				request.Client = client
				request.GrantTypes = oauthelia2.Arguments{oidc.GrantTypeRefreshToken}
				request.Form = url.Values{oidc.GrantTypeRefreshToken: {tokens[tokenRT0]}}

				require.NoError(t, grant.HandleTokenEndpointRequest(ctx, request))

				response := oauthelia2.NewAccessResponse()

				require.NoError(t, grant.PopulateTokenEndpointResponse(ctx, request, response))

				tokens[at] = response.GetAccessToken()
				tokens[rt] = response.GetExtra(oidc.GrantTypeRefreshToken).(string)
				signatures[at] = strategy.AccessTokenSignature(ctx, tokens[at])
				signatures[rt] = strategy.RefreshTokenSignature(ctx, tokens[rt])
			}

			refresh(tokenAT1, tokenRT1)
			refresh(tokenAT2, tokenRT2)

			isActive := func(name string) bool {
				if name[:2] == "at" {
					_, err := s.GetAccessTokenSession(ctx, signatures[name], oidc.NewSession())

					return err == nil
				}

				_, err := s.GetRefreshTokenSession(ctx, signatures[name], oidc.NewSession())

				return err == nil
			}

			assert.False(t, isActive(tokenAT0))

			for _, name := range []string{tokenAT1, tokenAT2, tokenRT0, tokenRT1, tokenRT2} {
				assert.True(t, isActive(name), name)
			}

			if tc.expired {
				fixed.Set(now.Add(2 * time.Minute))

				assert.False(t, isActive(tokenRT0))
			}

			require.NoError(t, revocation.RevokeToken(ctx, tokens[tc.token], tc.hint, client))

			for _, name := range tc.inactive {
				assert.False(t, isActive(name), name)
			}

			for _, name := range tc.active {
				assert.True(t, isActive(name), name)
			}
		})
	}
}

func TestOAuth2DeviceCodeSessionGrantedResourceRoundTripsThroughStore(t *testing.T) {
	ctx := context.Background()

	provider, err := storage.NewSQLiteProvider(&schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	})

	require.NoError(t, err)
	require.NoError(t, provider.StartupCheck())

	s := oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				IssuerCertificateChain: schema.X509CertificateChain{},
				IssuerPrivateKey:       x509PrivateKeyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  myclient,
						Name:                myclientname,
						AuthorizationPolicy: onefactor,
						Scopes:              []string{oidc.ScopeOpenID, oidc.ScopeProfile},
						Secret:              tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}, provider, nil)

	requester := &oauthelia2.DeviceAuthorizeRequest{
		Request: oauthelia2.Request{
			ID:                "req-resource-e2e-device",
			Client:            &oidc.RegisteredClient{ID: myclient},
			RequestedScope:    oauthelia2.Arguments{oidc.ScopeOpenID},
			RequestedAudience: oauthelia2.Arguments{"aud-a"},
			Session:           &oidc.Session{},
		},
		DeviceCodeSignature: "sig-resource-e2e-device",
	}

	consent := &model.OAuth2ConsentSession{
		GrantedScopes:   model.StringSlicePipeDelimited{oidc.ScopeOpenID},
		GrantedAudience: model.StringSlicePipeDelimited{"aud-a"},
		GrantedResource: model.StringSlicePipeDelimited{"https://api.example.com"},
	}

	oidc.GrantScopeAudienceConsent(requester, consent)

	require.NoError(t, s.CreateDeviceCodeSession(ctx, requester.DeviceCodeSignature, requester))

	loaded, err := s.GetDeviceCodeSession(ctx, requester.DeviceCodeSignature, &oidc.Session{})

	require.NoError(t, err)
	assert.Equal(t, oauthelia2.Arguments{oidc.ScopeOpenID}, loaded.GetGrantedScopes())
	assert.Equal(t, oauthelia2.Arguments{"aud-a"}, loaded.GetGrantedAudience())
	assert.Equal(t, oauthelia2.Arguments{"https://api.example.com"}, loaded.GetGrantedResource())
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, &StoreSuite{})
}

type StoreSuite struct {
	suite.Suite

	ctx   context.Context
	ctrl  *gomock.Controller
	mock  *mocks.MockStorage
	store *oidc.Store
}

func (s *StoreSuite) SetupTest() {
	s.ctx = context.Background()
	s.ctrl = gomock.NewController(s.T())
	s.mock = mocks.NewMockStorage(s.ctrl)
	s.store = oidc.NewStore(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  "hs256",
						Secret:              tOpenIDConnectPBKDF2ClientSecret,
						AuthorizationPolicy: authorization.OneFactor.String(),
						RedirectURIs: []string{
							"https://client.example.com",
						},
						TokenEndpointAuthMethod:     oidc.ClientAuthMethodClientSecretJWT,
						TokenEndpointAuthSigningAlg: oidc.SigningAlgHMACUsingSHA256,
					},
				},
			},
		},
	}, s.mock, nil)
}

func (s *StoreSuite) TestGetSubject() {
	s.T().Run("GenerateNew", func(t *testing.T) {
		s.mock.
			EXPECT().
			LoadUserOpaqueIdentifierBySignature(s.ctx, "openid", "", "john").
			Return(nil, nil)

		s.mock.
			EXPECT().
			SaveUserOpaqueIdentifier(s.ctx, gomock.Any()).
			Return(nil)

		opaqueID, err := s.store.GetSubject(s.ctx, "", "john")

		assert.NoError(t, err)
		assert.NotEqual(t, uint32(0), opaqueID)
	})

	s.T().Run("ReturnDatabaseErrorOnLoad", func(t *testing.T) {
		s.mock.
			EXPECT().
			LoadUserOpaqueIdentifierBySignature(s.ctx, "openid", "", "john").
			Return(nil, fmt.Errorf("failed to load"))

		opaqueID, err := s.store.GetSubject(s.ctx, "", "john")

		assert.EqualError(t, err, "failed to load")
		assert.Equal(t, uuid.Nil, opaqueID)
	})

	s.T().Run("ReturnDatabaseErrorOnSave", func(t *testing.T) {
		s.mock.
			EXPECT().
			LoadUserOpaqueIdentifierBySignature(s.ctx, "openid", "", "john").
			Return(nil, nil)

		s.mock.
			EXPECT().
			SaveUserOpaqueIdentifier(s.ctx, gomock.Any()).
			Return(fmt.Errorf("failed to save"))

		opaqueID, err := s.store.GetSubject(s.ctx, "", "john")

		assert.EqualError(t, err, "failed to save")
		assert.Equal(t, uuid.Nil, opaqueID)
	})

	s.T().Run("ReturnConcurrentlySaved", func(t *testing.T) {
		existing := &model.UserOpaqueIdentifier{Service: "openid", Username: "john", Identifier: uuid.MustParse("ef95d1a4-3d8e-4f07-a4ba-3a3e7e0c4a2e")}

		gomock.InOrder(
			s.mock.
				EXPECT().
				LoadUserOpaqueIdentifierBySignature(s.ctx, "openid", "", "john").
				Return(nil, nil),
			s.mock.
				EXPECT().
				SaveUserOpaqueIdentifier(s.ctx, gomock.Any()).
				Return(fmt.Errorf("error inserting user opaque id for user 'john': %w", sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique})),
			s.mock.
				EXPECT().
				LoadUserOpaqueIdentifierBySignature(s.ctx, "openid", "", "john").
				Return(existing, nil),
		)

		opaqueID, err := s.store.GetSubject(s.ctx, "", "john")

		assert.NoError(t, err)
		assert.Equal(t, existing.Identifier, opaqueID)
	})

	s.T().Run("ReturnDatabaseErrorOnLoadConcurrentlySaved", func(t *testing.T) {
		gomock.InOrder(
			s.mock.
				EXPECT().
				LoadUserOpaqueIdentifierBySignature(s.ctx, "openid", "", "john").
				Return(nil, nil),
			s.mock.
				EXPECT().
				SaveUserOpaqueIdentifier(s.ctx, gomock.Any()).
				Return(sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique}),
			s.mock.
				EXPECT().
				LoadUserOpaqueIdentifierBySignature(s.ctx, "openid", "", "john").
				Return(nil, fmt.Errorf("failed to load")),
		)

		opaqueID, err := s.store.GetSubject(s.ctx, "", "john")

		assert.EqualError(t, err, "failed to load")
		assert.Equal(t, uuid.Nil, opaqueID)
	})

	s.T().Run("ReturnErrorOnConcurrentlySavedNotFound", func(t *testing.T) {
		gomock.InOrder(
			s.mock.
				EXPECT().
				LoadUserOpaqueIdentifierBySignature(s.ctx, "openid", "", "john").
				Return(nil, nil),
			s.mock.
				EXPECT().
				SaveUserOpaqueIdentifier(s.ctx, gomock.Any()).
				Return(sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique}),
			s.mock.
				EXPECT().
				LoadUserOpaqueIdentifierBySignature(s.ctx, "openid", "", "john").
				Return(nil, nil),
		)

		opaqueID, err := s.store.GetSubject(s.ctx, "", "john")

		assert.EqualError(t, err, "error loading the user opaque id which was saved concurrently for user 'john': the user opaque id was not found")
		assert.Equal(t, uuid.Nil, opaqueID)
	})
}

func (s *StoreSuite) TestTx() {
	gomock.InOrder(
		s.mock.EXPECT().BeginTX(s.ctx).Return(s.ctx, nil),
		s.mock.EXPECT().Commit(s.ctx).Return(nil),
		s.mock.EXPECT().Rollback(s.ctx).Return(nil),
		s.mock.EXPECT().BeginTX(s.ctx).Return(nil, fmt.Errorf("failed to begin")),
		s.mock.EXPECT().Commit(s.ctx).Return(fmt.Errorf("failed to commit")),
		s.mock.EXPECT().Rollback(s.ctx).Return(fmt.Errorf("failed to rollback")),
	)

	x, err := s.store.BeginTX(s.ctx)
	s.Equal(s.ctx, x)
	s.NoError(err)
	s.NoError(s.store.Commit(s.ctx))
	s.NoError(s.store.Rollback(s.ctx))

	x, err = s.store.BeginTX(s.ctx)
	s.Equal(nil, x)
	s.EqualError(err, "failed to begin")
	s.EqualError(s.store.Commit(s.ctx), "failed to commit")
	s.EqualError(s.store.Rollback(s.ctx), "failed to rollback")
}

func (s *StoreSuite) TestClientAssertionJWTValid() {
	gomock.InOrder(
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "client", "3a240379e8286a7a8ff5e99d68567e0e5e34e80168b8feffa89d3d33dea95b63").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "3a240379e8286a7a8ff5e99d68567e0e5e34e80168b8feffa89d3d33dea95b63",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "client", "e7f67ad76c80d57d34b19598462817932aec21d2806a08a786a8d4b9dd476068").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "e7f67ad76c80d57d34b19598462817932aec21d2806a08a786a8d4b9dd476068",
				ExpiresAt: time.Now().Add(-time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "client", "f29ef0d85303a09411b76001c579980f1b1b7fc9deb1fa647875a724f4f231c6").
			Return(nil, fmt.Errorf("failed to load")),
	)

	s.EqualError(s.store.ClientAssertionJWTValid(s.ctx, "client", "066ee771-e156-4886-b99f-ee09b0d3edf4"), "jti_known")
	s.NoError(s.store.ClientAssertionJWTValid(s.ctx, "client", "5dad3ff7-e4f2-41b6-98a3-b73d872076ce"))
	s.EqualError(s.store.ClientAssertionJWTValid(s.ctx, "client", "65471ccb-d650-4006-a95f-cb4f4e3d7202"), "failed to load")
}

func (s *StoreSuite) TestCreateSessions() {
	challenge := model.MustNullUUID(model.NewRandomNullUUID())
	session := &oidc.Session{
		ChallengeID: challenge,
	}
	sessionData, _ := json.Marshal(session)

	gomock.InOrder(
		s.mock.
			EXPECT().
			SaveOAuth2Session(s.ctx, storage.OAuth2SessionTypeAuthorizeCode, model.OAuth2Session{ChallengeID: challenge, RequestID: abc, ClientID: "example", Signature: abc, Active: true, Session: sessionData, RequestedScopes: model.StringSlicePipeDelimited{}, GrantedScopes: model.StringSlicePipeDelimited{}}).
			Return(nil),
		s.mock.
			EXPECT().
			SaveOAuth2Session(s.ctx, storage.OAuth2SessionTypeAuthorizeCode, model.OAuth2Session{ChallengeID: challenge, RequestID: abc, ClientID: "example", Signature: abc, Active: true, Session: sessionData, RequestedScopes: model.StringSlicePipeDelimited{}, GrantedScopes: model.StringSlicePipeDelimited{}}).
			Return(fmt.Errorf("duplicate key")),
		s.mock.
			EXPECT().
			SaveOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, model.OAuth2Session{ChallengeID: challenge, RequestID: abc, ClientID: "example", Signature: abc, Active: true, Session: sessionData, RequestedScopes: model.StringSlicePipeDelimited{}, GrantedScopes: model.StringSlicePipeDelimited{}}).
			Return(nil),
		s.mock.
			EXPECT().
			SaveOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, model.OAuth2Session{ChallengeID: challenge, RequestID: abc, ClientID: "example", Signature: abc, Active: true, Session: sessionData, RequestedScopes: model.StringSlicePipeDelimited{}, GrantedScopes: model.StringSlicePipeDelimited{}}).
			Return(nil),
		s.mock.
			EXPECT().
			SaveOAuth2Session(s.ctx, storage.OAuth2SessionTypeOpenIDConnect, model.OAuth2Session{ChallengeID: challenge, RequestID: abc, ClientID: "example", Signature: abc, Active: true, Session: sessionData, RequestedScopes: model.StringSlicePipeDelimited{}, GrantedScopes: model.StringSlicePipeDelimited{}}).
			Return(nil),
		s.mock.
			EXPECT().
			SaveOAuth2Session(s.ctx, storage.OAuth2SessionTypePKCEChallenge, model.OAuth2Session{ChallengeID: challenge, RequestID: abc, ClientID: "example", Signature: abc, Active: true, Session: sessionData, RequestedScopes: model.StringSlicePipeDelimited{}, GrantedScopes: model.StringSlicePipeDelimited{}}).
			Return(nil),
		s.mock.
			EXPECT().
			SaveOAuth2PushedAuthorizationSession(s.ctx, model.OAuth2PushedAuthorizationSession{Signature: abc, RequestID: abc, ClientID: "example", Session: sessionData}).
			Return(nil),
	)

	s.NoError(s.store.CreateAuthorizeCodeSession(s.ctx, abc, &oauthelia2.Request{
		ID: abc,
		Client: &oidc.RegisteredClient{
			ID: "example",
		},
		Session: session,
	}))

	s.EqualError(s.store.CreateAuthorizeCodeSession(s.ctx, abc, &oauthelia2.Request{
		ID: abc,
		Client: &oidc.RegisteredClient{
			ID: "example",
		},
		Session: session,
	}), "duplicate key")

	s.EqualError(s.store.CreateAuthorizeCodeSession(s.ctx, abc, &oauthelia2.Request{
		ID: abc,
		Client: &oidc.RegisteredClient{
			ID: "example",
		},
		Session: nil,
	}), "failed to create new *model.OAuth2Session: the session type OpenIDSession was expected but the type '<nil>' was used")

	s.NoError(s.store.CreateAccessTokenSession(s.ctx, abc, &oauthelia2.Request{
		ID: abc,
		Client: &oidc.RegisteredClient{
			ID: "example",
		},
		Session: session,
	}))

	s.NoError(s.store.CreateRefreshTokenSession(s.ctx, abc, "", &oauthelia2.Request{
		ID: abc,
		Client: &oidc.RegisteredClient{
			ID: "example",
		},
		Session: session,
	}))

	s.NoError(s.store.CreateOpenIDConnectSession(s.ctx, abc, &oauthelia2.Request{
		ID: abc,
		Client: &oidc.RegisteredClient{
			ID: "example",
		},
		Session: session,
	}))

	s.NoError(s.store.CreatePKCERequestSession(s.ctx, abc, &oauthelia2.Request{
		ID: abc,
		Client: &oidc.RegisteredClient{
			ID: "example",
		},
		Session: session,
	}))

	s.NoError(s.store.CreatePARSession(s.ctx, abc, &oauthelia2.AuthorizeRequest{
		Request: oauthelia2.Request{
			ID: abc,
			Client: &oidc.RegisteredClient{
				ID: "example",
			},
			Session: session,
		}}))

	s.EqualError(s.store.CreatePARSession(s.ctx, abc, &oauthelia2.AuthorizeRequest{
		Request: oauthelia2.Request{
			ID: abc,
			Client: &oidc.RegisteredClient{
				ID: "example",
			},
			Session: nil,
		}}), "failed to create new PAR context: can't assert type '<nil>' to an *OAuth2Session")
}

func (s *StoreSuite) TestRevokeSessions() {
	gomock.InOrder(
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeAuthorizeCode, "abc1").
			Return(nil),
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeAuthorizeCode, "abc2").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, "at_example1").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, "at_example2").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			RevokeOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeAccessToken, "65471ccb-d650-4006-a95f-cb4f4e3d7200").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeAccessToken, "65471ccb-d650-4006-a95f-cb4f4e3d7201").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			RevokeOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeAccessToken, "65471ccb-d650-4006-a95f-cb4f4e3d7202").
			Return(sql.ErrNoRows),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "rt_example1").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "rt_example2").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			DeactivateOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeRefreshToken, "65471ccb-d650-4006-a95f-cb4f4e3d7200").
			Return(nil),
		s.mock.
			EXPECT().
			DeactivateOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeRefreshToken, "65471ccb-d650-4006-a95f-cb4f4e3d7201").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			DeactivateOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeRefreshToken, "65471ccb-d650-4006-a95f-cb4f4e3d7202").
			Return(sql.ErrNoRows),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionState(s.ctx, "1").
			Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at_paired_1", Active: true}, nil),
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "1").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, "at_paired_1").
			Return(nil),
		s.mock.
			EXPECT().
			DeactivateOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeRefreshToken, "65471ccb-d650-4006-a95f-cb4f4e3d7200").
			Return(nil),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionState(s.ctx, "2").
			Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "", Active: true}, nil),
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "2").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeAccessToken, "65471ccb-d650-4006-a95f-cb4f4e3d7201").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionState(s.ctx, "3").
			Return(nil, fmt.Errorf("no refresh token session")),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionState(s.ctx, "4").
			Return(nil, sql.ErrNoRows),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionState(s.ctx, "5").
			Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at_paired_5", Active: true}, nil),
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "5").
			Return(storage.ErrNoRowsAffected),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionState(s.ctx, "6").
			Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at_paired_6", Active: true}, nil),
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "6").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, "at_paired_6").
			Return(storage.ErrNoRowsAffected),
		s.mock.
			EXPECT().
			DeactivateOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeRefreshToken, "65471ccb-d650-4006-a95f-cb4f4e3d7206").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypePKCEChallenge, "pkce1").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypePKCEChallenge, "pkce2").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeOpenIDConnect, "ac_1").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeOpenIDConnect, "ac_2").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			RevokeOAuth2PushedAuthorizationSession(s.ctx, "urn:par1").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2PushedAuthorizationSession(s.ctx, "urn:par2").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			RevokeOAuth2PushedAuthorizationSession(s.ctx, "urn:par3").
			Return(storage.ErrNoRowsAffected),
	)

	s.NoError(s.store.InvalidateAuthorizeCodeSession(s.ctx, "abc1"))
	s.EqualError(s.store.InvalidateAuthorizeCodeSession(s.ctx, "abc2"), "not found")

	s.NoError(s.store.DeleteAccessTokenSession(s.ctx, "at_example1"))
	s.EqualError(s.store.DeleteAccessTokenSession(s.ctx, "at_example2"), "not found")

	s.NoError(s.store.RevokeAccessToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7200"))
	s.EqualError(s.store.RevokeAccessToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7201"), "not found")
	s.EqualError(s.store.RevokeAccessToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7202"), "not_found")

	s.NoError(s.store.DeleteRefreshTokenSession(s.ctx, "rt_example1"))
	s.EqualError(s.store.DeleteRefreshTokenSession(s.ctx, "rt_example2"), "not found")

	s.NoError(s.store.RevokeRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7200"))
	s.EqualError(s.store.RevokeRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7201"), "not found")
	s.EqualError(s.store.RevokeRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7202"), "sql: no rows in result set")

	s.NoError(s.store.RotateRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7200", "1"))
	s.EqualError(s.store.RotateRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7201", "2"), "not found")
	s.EqualError(s.store.RotateRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7202", "3"), "no refresh token session")
	s.ErrorIs(s.store.RotateRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7204", "4"), oauthelia2.ErrNotFound)
	s.ErrorIs(s.store.RotateRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7205", "5"), oauthelia2.ErrInactiveToken)
	s.NoError(s.store.RotateRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7206", "6"))

	s.NoError(s.store.DeletePKCERequestSession(s.ctx, "pkce1"))
	s.EqualError(s.store.DeletePKCERequestSession(s.ctx, "pkce2"), "not found")

	s.NoError(s.store.DeleteOpenIDConnectSession(s.ctx, "ac_1"))
	s.EqualError(s.store.DeleteOpenIDConnectSession(s.ctx, "ac_2"), "not found")

	s.NoError(s.store.DeletePARSession(s.ctx, "urn:par1"))
	s.EqualError(s.store.DeletePARSession(s.ctx, "urn:par2"), "not found")
	s.ErrorIs(s.store.DeletePARSession(s.ctx, "urn:par3"), oauthelia2.ErrNotFound)
}

func (s *StoreSuite) TestDeleteSessionsShouldReturnNotFound() {
	testCases := []struct {
		name        string
		sessionType storage.OAuth2SessionType
		do          func(signature string) (err error)
	}{
		{"ShouldMapDeleteAccessTokenSession", storage.OAuth2SessionTypeAccessToken, func(signature string) error { return s.store.DeleteAccessTokenSession(s.ctx, signature) }},
		{"ShouldMapDeleteRefreshTokenSession", storage.OAuth2SessionTypeRefreshToken, func(signature string) error { return s.store.DeleteRefreshTokenSession(s.ctx, signature) }},
		{"ShouldMapDeletePKCERequestSession", storage.OAuth2SessionTypePKCEChallenge, func(signature string) error { return s.store.DeletePKCERequestSession(s.ctx, signature) }},
		{"ShouldMapDeleteOpenIDConnectSession", storage.OAuth2SessionTypeOpenIDConnect, func(signature string) error { return s.store.DeleteOpenIDConnectSession(s.ctx, signature) }},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.mock.EXPECT().RevokeOAuth2Session(s.ctx, tc.sessionType, abc).Return(storage.ErrNoRowsAffected)

			assert.ErrorIs(t, tc.do(abc), oauthelia2.ErrNotFound)
		})
	}
}

func (s *StoreSuite) TestGetRefreshTokenSessionShouldReturnInactiveWhenRevoked() {
	sessionData, err := json.Marshal(&oidc.Session{ClientID: "hs256"})
	s.Require().NoError(err)

	testCases := []struct {
		name   string
		active bool
	}{
		{"ShouldReturnInactiveWhenRevokedAndActive", true},
		{"ShouldReturnInactiveWhenRevokedAndInactive", false},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, abc).
				Return(&model.OAuth2Session{RequestID: abc, ClientID: "hs256", Session: sessionData, Active: tc.active, Revoked: true}, nil)

			r, err := s.store.GetRefreshTokenSession(s.ctx, abc, &oidc.Session{})

			require.NotNil(t, r)
			assert.Equal(t, abc, r.GetID())
			assert.ErrorIs(t, err, oauthelia2.ErrInactiveToken)
		})
	}
}

func (s *StoreSuite) TestGetRefreshTokenSessionShouldReturnInactiveWhenGracePeriodEnded() {
	sessionData, err := json.Marshal(&oidc.Session{ClientID: "hs256"})
	s.Require().NoError(err)

	testCases := []struct {
		name     string
		deadline sql.NullTime
		err      error
	}{
		{"ShouldReturnActiveWithoutDeadline", sql.NullTime{}, nil},
		{"ShouldReturnActiveWithinDeadline", sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}, nil},
		{"ShouldReturnInactiveAfterDeadline", sql.NullTime{Time: time.Now().Add(-time.Second), Valid: true}, oauthelia2.ErrInactiveToken},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, abc).
				Return(&model.OAuth2Session{RequestID: abc, ClientID: "hs256", Session: sessionData, Active: true, GraceDeadline: tc.deadline}, nil)

			r, err := s.store.GetRefreshTokenSession(s.ctx, abc, &oidc.Session{})

			require.NotNil(t, r)
			assert.Equal(t, abc, r.GetID())

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.err)
			}
		})
	}
}

func (s *StoreSuite) TestRotateRefreshTokenShouldReturnInactiveWhenGracePeriodEnded() {
	s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).
		Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true, GraceDeadline: sql.NullTime{Time: time.Now().Add(-time.Second), Valid: true}}, nil)

	assert.ErrorIs(s.T(), s.store.RotateRefreshToken(s.ctx, abc, abc), oauthelia2.ErrInactiveToken)
}

func (s *StoreSuite) TestRotateRefreshTokenWithGracePeriod() {
	deadline := time.Now().Add(time.Minute)
	future := sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}
	past := sql.NullTime{Time: time.Now().Add(-time.Second), Valid: true}

	testCases := []struct {
		name  string
		setup func()
		err   error
		msg   string
	}{
		{
			name: "ShouldRecordDeadlineAndRevokePairedAccessToken",
			setup: func() {
				gomock.InOrder(
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true}, nil),
					s.mock.EXPECT().UpdateOAuth2RefreshTokenSessionGraceDeadline(s.ctx, abc, deadline).Return(nil),
					s.mock.EXPECT().RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, "at").Return(nil),
				)
			},
		},
		{
			name: "ShouldIgnoreAlreadyRevokedPairedAccessToken",
			setup: func() {
				gomock.InOrder(
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true}, nil),
					s.mock.EXPECT().UpdateOAuth2RefreshTokenSessionGraceDeadline(s.ctx, abc, deadline).Return(nil),
					s.mock.EXPECT().RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, "at").Return(storage.ErrNoRowsAffected),
				)
			},
		},
		{
			name: "ShouldRevokeAccessTokensByRequestIDWithoutPairedSignature",
			setup: func() {
				gomock.InOrder(
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{Active: true}, nil),
					s.mock.EXPECT().UpdateOAuth2RefreshTokenSessionGraceDeadline(s.ctx, abc, deadline).Return(nil),
					s.mock.EXPECT().RevokeOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeAccessToken, "req").Return(nil),
				)
			},
		},
		{
			name: "ShouldReturnErrorRevokingPairedAccessToken",
			setup: func() {
				gomock.InOrder(
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true}, nil),
					s.mock.EXPECT().UpdateOAuth2RefreshTokenSessionGraceDeadline(s.ctx, abc, deadline).Return(nil),
					s.mock.EXPECT().RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, "at").Return(fmt.Errorf("revoke error")),
				)
			},
			msg: "revoke error",
		},
		{
			name: "ShouldSucceedWithinRecordedDeadline",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true, GraceDeadline: future}, nil)
			},
		},
		{
			name: "ShouldReturnInactiveAfterRecordedDeadline",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true, GraceDeadline: past}, nil)
			},
			err: oauthelia2.ErrInactiveToken,
		},
		{
			name: "ShouldReturnInactiveWhenDeactivated",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", GraceDeadline: future}, nil)
			},
			err: oauthelia2.ErrInactiveToken,
		},
		{
			name: "ShouldReturnInactiveWhenDeactivatedWithoutDeadline",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at"}, nil)
			},
			err: oauthelia2.ErrInactiveToken,
		},
		{
			name: "ShouldReturnInactiveWhenRevoked",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true, Revoked: true}, nil)
			},
			err: oauthelia2.ErrInactiveToken,
		},
		{
			name: "ShouldSucceedWhenDeadlineRecordedConcurrently",
			setup: func() {
				gomock.InOrder(
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true}, nil),
					s.mock.EXPECT().UpdateOAuth2RefreshTokenSessionGraceDeadline(s.ctx, abc, deadline).Return(storage.ErrNoRowsAffected),
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true, GraceDeadline: future}, nil),
				)
			},
		},
		{
			name: "ShouldReturnInactiveWhenDeactivatedConcurrently",
			setup: func() {
				gomock.InOrder(
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true}, nil),
					s.mock.EXPECT().UpdateOAuth2RefreshTokenSessionGraceDeadline(s.ctx, abc, deadline).Return(storage.ErrNoRowsAffected),
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at"}, nil),
				)
			},
			err: oauthelia2.ErrInactiveToken,
		},
		{
			name: "ShouldReturnNotFoundWhenDeletedConcurrently",
			setup: func() {
				gomock.InOrder(
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true}, nil),
					s.mock.EXPECT().UpdateOAuth2RefreshTokenSessionGraceDeadline(s.ctx, abc, deadline).Return(storage.ErrNoRowsAffected),
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(nil, sql.ErrNoRows),
				)
			},
			err: oauthelia2.ErrNotFound,
		},
		{
			name: "ShouldReturnNotFound",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(nil, sql.ErrNoRows)
			},
			err: oauthelia2.ErrNotFound,
		},
		{
			name: "ShouldReturnErrorLoadingState",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(nil, fmt.Errorf("load error"))
			},
			msg: "load error",
		},
		{
			name: "ShouldReturnErrorRecordingDeadline",
			setup: func() {
				gomock.InOrder(
					s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(&model.OAuth2RefreshTokenSessionState{AccessSignature: "at", Active: true}, nil),
					s.mock.EXPECT().UpdateOAuth2RefreshTokenSessionGraceDeadline(s.ctx, abc, deadline).Return(fmt.Errorf("update error")),
				)
			},
			msg: "update error",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			err := s.store.RotateRefreshTokenWithGracePeriod(s.ctx, "req", abc, deadline)

			switch {
			case tc.err != nil:
				assert.ErrorIs(t, err, tc.err)
			case tc.msg != "":
				assert.EqualError(t, err, tc.msg)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

func (s *StoreSuite) TestGetSessions() {
	challenge := model.MustNullUUID(model.NewRandomNullUUID())
	session := &oidc.Session{
		ChallengeID: challenge,
		ClientID:    "hs256",
	}
	sessionData, _ := json.Marshal(session)

	sessionb := &oidc.Session{
		ChallengeID: challenge,
		ClientID:    "hs256",
	}
	sessionDatab, _ := json.Marshal(sessionb)

	gomock.InOrder(
		s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeAuthorizeCode, "ac_123").
			Return(&model.OAuth2Session{ClientID: "hs256", Session: sessionData, Active: true}, nil),
		s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeAuthorizeCode, "ac_456").
			Return(&model.OAuth2Session{ClientID: "hs256", Session: sessionData, Active: false}, nil),
		s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeAuthorizeCode, "ac_aaa").
			Return(nil, sql.ErrNoRows),
		s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeAuthorizeCode, "ac_130").
			Return(nil, fmt.Errorf("timeout")),
		s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeAuthorizeCode, "ac_badclient").
			Return(&model.OAuth2Session{ClientID: "no-client", Session: sessionDatab, Active: true}, nil),
		s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, "at").
			Return(&model.OAuth2Session{ClientID: "hs256", Session: sessionData, Active: true}, nil),
		s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "rt").
			Return(&model.OAuth2Session{ClientID: "hs256", Session: sessionData, Active: true}, nil),
		s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypePKCEChallenge, "pkce").
			Return(&model.OAuth2Session{ClientID: "hs256", Session: sessionData, Active: true}, nil),
		s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeOpenIDConnect, "ot").
			Return(&model.OAuth2Session{ClientID: "hs256", Session: sessionData, Active: true}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2PushedAuthorizationSession(s.ctx, "urn:par").
			Return(&model.OAuth2PushedAuthorizationSession{Signature: abc, RequestID: abc, ClientID: "hs256", Session: sessionData}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2PushedAuthorizationSession(s.ctx, "urn:par").
			Return(nil, sql.ErrNoRows),
		s.mock.
			EXPECT().
			LoadOAuth2PushedAuthorizationSession(s.ctx, "urn:par").
			Return(nil, fmt.Errorf("connection refused")),
		s.mock.
			EXPECT().
			LoadOAuth2PushedAuthorizationSession(s.ctx, "urn:par").
			Return(&model.OAuth2PushedAuthorizationSession{Signature: abc, RequestID: abc, ClientID: "hs256", Session: sessionData, Revoked: true}, nil),
	)

	var (
		r   oauthelia2.Requester
		err error
	)

	r, err = s.store.GetAuthorizeCodeSession(s.ctx, "ac_123", &oidc.Session{})
	s.NotNil(r)
	s.NoError(err)

	r, err = s.store.GetAuthorizeCodeSession(s.ctx, "ac_456", &oidc.Session{})
	s.NotNil(r)
	s.EqualError(err, "Authorization code has ben invalidated")

	r, err = s.store.GetAuthorizeCodeSession(s.ctx, "ac_aaa", &oidc.Session{})
	s.Nil(r)
	s.EqualError(err, "not_found")

	r, err = s.store.GetAuthorizeCodeSession(s.ctx, "ac_130", &oidc.Session{})
	s.Nil(r)
	s.EqualError(err, "timeout")

	r, err = s.store.GetAuthorizeCodeSession(s.ctx, "ac_badclient", &oidc.Session{})
	s.Nil(r)
	s.EqualError(err, "error occurred while mapping OAuth 2.0 Session back to a Request while trying to lookup the registered client: invalid_client")

	r, err = s.store.GetAccessTokenSession(s.ctx, "at", &oidc.Session{})
	s.NotNil(r)
	s.NoError(err)

	r, err = s.store.GetRefreshTokenSession(s.ctx, "rt", &oidc.Session{})
	s.NotNil(r)
	s.NoError(err)

	r, err = s.store.GetPKCERequestSession(s.ctx, "pkce", &oidc.Session{})
	s.NotNil(r)
	s.NoError(err)

	r, err = s.store.GetOpenIDConnectSession(s.ctx, "ot", &oauthelia2.Request{
		ID: abc,
		Client: &oidc.RegisteredClient{
			ID: "example",
		},
		Session: session,
	})
	s.NotNil(r)
	s.NoError(err)

	r, err = s.store.GetPARSession(s.ctx, "urn:par")
	s.NotNil(r)
	s.NoError(err)

	r, err = s.store.GetPARSession(s.ctx, "urn:par")
	s.Nil(r)
	s.EqualError(oauthelia2.ErrorToDebugRFC6749Error(err), "Could not find the requested resource(s). The requested PAR session was not found, was expired, or was otherwise invalid.")

	r, err = s.store.GetPARSession(s.ctx, "urn:par")
	s.Nil(r)
	s.EqualError(oauthelia2.ErrorToDebugRFC6749Error(err), "The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Error occurred retrieving the PAR session from storage: connection refused")

	r, err = s.store.GetPARSession(s.ctx, "urn:par")
	s.Nil(r)
	s.EqualError(oauthelia2.ErrorToDebugRFC6749Error(err), "Could not find the requested resource(s). The requested PAR session was not found, was expired, or was otherwise invalid.")
}

func (s *StoreSuite) TestSetClientAssertionJWT() {
	gomock.InOrder(
		s.mock.
			EXPECT().
			SaveOAuth2BlacklistedJTI(s.ctx, model.OAuth2BlacklistedJTI{ClientID: "client", Signature: "f29ef0d85303a09411b76001c579980f1b1b7fc9deb1fa647875a724f4f231c6", ExpiresAt: time.Unix(160000000, 0)}).
			Return(nil),
		s.mock.
			EXPECT().
			SaveOAuth2BlacklistedJTI(s.ctx, model.OAuth2BlacklistedJTI{ClientID: "client", Signature: "0dab0de97ed4e05da82763497448daf4f6b555c99218100e3ef5a81f36232940", ExpiresAt: time.Unix(160000000, 0)}).
			Return(storage.ErrNoRowsAffected),
		s.mock.
			EXPECT().
			SaveOAuth2BlacklistedJTI(s.ctx, model.OAuth2BlacklistedJTI{ClientID: "client", Signature: "0dab0de97ed4e05da82763497448daf4f6b555c99218100e3ef5a81f36232940", ExpiresAt: time.Unix(160000000, 0)}).
			Return(fmt.Errorf("failed to save")),
	)

	s.NoError(s.store.SetClientAssertionJWT(s.ctx, "client", "65471ccb-d650-4006-a95f-cb4f4e3d7202", time.Unix(160000000, 0)))
	s.ErrorIs(s.store.SetClientAssertionJWT(s.ctx, "client", "65471ccb-d650-4006-a95f-cb4f4e3d7201", time.Unix(160000000, 0)), oauthelia2.ErrJTIKnown)
	s.EqualError(s.store.SetClientAssertionJWT(s.ctx, "client", "65471ccb-d650-4006-a95f-cb4f4e3d7201", time.Unix(160000000, 0)), "failed to save")
}

func (s *StoreSuite) TestSetRequestReplay() {
	gomock.InOrder(
		s.mock.
			EXPECT().
			SaveOAuth2RequestReplay(s.ctx, model.NewOAuth2RequestReplay(model.OAuth2RequestReplayTypePKCEChallenge, "client", "challenge", time.Unix(160000000, 0))).
			Return(nil),
		s.mock.
			EXPECT().
			SaveOAuth2RequestReplay(s.ctx, model.NewOAuth2RequestReplay(model.OAuth2RequestReplayTypePKCEChallenge, "client", "challenge", time.Unix(160000000, 0))).
			Return(storage.ErrNoRowsAffected),
		s.mock.
			EXPECT().
			SaveOAuth2RequestReplay(s.ctx, model.NewOAuth2RequestReplay(model.OAuth2RequestReplayTypeOpenIDConnectNonce, "client", "nonce", time.Unix(160000000, 0))).
			Return(nil),
		s.mock.
			EXPECT().
			SaveOAuth2RequestReplay(s.ctx, model.NewOAuth2RequestReplay(model.OAuth2RequestReplayTypeOpenIDConnectNonce, "client", "nonce", time.Unix(160000000, 0))).
			Return(storage.ErrNoRowsAffected),
		s.mock.
			EXPECT().
			SaveOAuth2RequestReplay(s.ctx, model.NewOAuth2RequestReplay(model.OAuth2RequestReplayTypeOpenIDConnectNonce, "client", "nonce", time.Unix(160000000, 0))).
			Return(fmt.Errorf("failed to save")),
	)

	s.NoError(s.store.SetPKCEChallengeUsed(s.ctx, "client", "challenge", time.Unix(160000000, 0)))
	s.ErrorIs(s.store.SetPKCEChallengeUsed(s.ctx, "client", "challenge", time.Unix(160000000, 0)), oauthelia2.ErrPKCEChallengeReused)
	s.NoError(s.store.SetOpenIDConnectNonceUsed(s.ctx, "client", "nonce", time.Unix(160000000, 0)))
	s.ErrorIs(s.store.SetOpenIDConnectNonceUsed(s.ctx, "client", "nonce", time.Unix(160000000, 0)), oauthelia2.ErrOpenIDConnectNonceReused)
	s.EqualError(s.store.SetOpenIDConnectNonceUsed(s.ctx, "client", "nonce", time.Unix(160000000, 0)), "failed to save")
}

func (s *StoreSuite) TestCreateDeviceCodeSession() {
	challenge := model.MustNullUUID(model.NewRandomNullUUID())
	session := &oidc.Session{ChallengeID: challenge}

	s.T().Run("ShouldSucceed", func(t *testing.T) {
		s.mock.EXPECT().SaveOAuth2DeviceCodeSession(s.ctx, gomock.Any()).Return(nil)

		err := s.store.CreateDeviceCodeSession(s.ctx, abc, &oauthelia2.DeviceAuthorizeRequest{
			Request: oauthelia2.Request{
				ID:      abc,
				Client:  &oidc.RegisteredClient{ID: "example"},
				Session: session,
			},
		})

		assert.NoError(t, err)
	})

	s.T().Run("ShouldErrOnStorageFailure", func(t *testing.T) {
		s.mock.EXPECT().SaveOAuth2DeviceCodeSession(s.ctx, gomock.Any()).Return(fmt.Errorf("storage error"))

		err := s.store.CreateDeviceCodeSession(s.ctx, abc, &oauthelia2.DeviceAuthorizeRequest{
			Request: oauthelia2.Request{
				ID:      abc,
				Client:  &oidc.RegisteredClient{ID: "example"},
				Session: session,
			},
		})

		assert.EqualError(t, err, "storage error")
	})

	s.T().Run("ShouldErrOnNilSession", func(t *testing.T) {
		err := s.store.CreateDeviceCodeSession(s.ctx, abc, &oauthelia2.DeviceAuthorizeRequest{
			Request: oauthelia2.Request{
				ID:      abc,
				Client:  &oidc.RegisteredClient{ID: "example"},
				Session: nil,
			},
		})

		assert.Error(t, err)
	})
}

func (s *StoreSuite) TestUpdateDeviceCodeSession() {
	challenge := model.MustNullUUID(model.NewRandomNullUUID())
	session := &oidc.Session{ChallengeID: challenge}

	s.T().Run("ShouldSucceed", func(t *testing.T) {
		s.mock.EXPECT().UpdateOAuth2DeviceCodeSessionData(s.ctx, gomock.Any()).Return(nil)

		err := s.store.UpdateDeviceCodeSession(s.ctx, abc, &oauthelia2.DeviceAuthorizeRequest{
			Request: oauthelia2.Request{
				ID:      abc,
				Client:  &oidc.RegisteredClient{ID: "example"},
				Session: session,
			},
		})

		assert.NoError(t, err)
	})

	s.T().Run("ShouldErrOnStorageFailure", func(t *testing.T) {
		s.mock.EXPECT().UpdateOAuth2DeviceCodeSessionData(s.ctx, gomock.Any()).Return(fmt.Errorf("update error"))

		err := s.store.UpdateDeviceCodeSession(s.ctx, abc, &oauthelia2.DeviceAuthorizeRequest{
			Request: oauthelia2.Request{
				ID:      abc,
				Client:  &oidc.RegisteredClient{ID: "example"},
				Session: session,
			},
		})

		assert.EqualError(t, err, "update error")
	})
}

func (s *StoreSuite) TestGetDeviceCodeSession() {
	challenge := model.MustNullUUID(model.NewRandomNullUUID())
	session := &oidc.Session{ChallengeID: challenge}
	sessionData, _ := json.Marshal(session)

	s.T().Run("ShouldSucceedWithActiveSession", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSession(s.ctx, abc).Return(&model.OAuth2DeviceCodeSession{
			Active:    true,
			Signature: abc,
			RequestID: abc,
			ClientID:  "hs256",
			Session:   sessionData,
		}, nil)

		request, err := s.store.GetDeviceCodeSession(s.ctx, abc, session)

		assert.NoError(t, err)
		assert.NotNil(t, request)
	})

	s.T().Run("ShouldErrWhenInactive", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSession(s.ctx, abc).Return(&model.OAuth2DeviceCodeSession{
			Active:    false,
			Signature: abc,
		}, nil)

		request, err := s.store.GetDeviceCodeSession(s.ctx, abc, session)

		assert.Error(t, err)
		assert.Nil(t, request)
	})

	s.T().Run("ShouldErrOnStorageFailure", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSession(s.ctx, abc).Return(nil, fmt.Errorf("load error"))

		request, err := s.store.GetDeviceCodeSession(s.ctx, abc, session)

		assert.EqualError(t, err, "load error")
		assert.Nil(t, request)
	})

	s.T().Run("ShouldCreateDefaultSessionWhenNil", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSession(s.ctx, abc).Return(&model.OAuth2DeviceCodeSession{
			Active:    true,
			Signature: abc,
			RequestID: abc,
			ClientID:  "hs256",
			Session:   sessionData,
		}, nil)

		request, err := s.store.GetDeviceCodeSession(s.ctx, abc, nil)

		assert.NoError(t, err)
		assert.NotNil(t, request)
	})
}

func (s *StoreSuite) TestInvalidateDeviceCodeSession() {
	s.T().Run("ShouldSucceed", func(t *testing.T) {
		s.mock.EXPECT().DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeDeviceAuthorizeCode, abc).Return(nil)

		err := s.store.InvalidateDeviceCodeSession(s.ctx, abc)

		assert.NoError(t, err)
	})

	s.T().Run("ShouldErrOnStorageFailure", func(t *testing.T) {
		s.mock.EXPECT().DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeDeviceAuthorizeCode, abc).Return(fmt.Errorf("deactivate error"))

		err := s.store.InvalidateDeviceCodeSession(s.ctx, abc)

		assert.EqualError(t, err, "deactivate error")
	})
}

func (s *StoreSuite) TestGetDeviceCodeSessionByUserCode() {
	challenge := model.MustNullUUID(model.NewRandomNullUUID())
	session := &oidc.Session{ChallengeID: challenge}
	sessionData, _ := json.Marshal(session)

	s.T().Run("ShouldSucceedWithActiveSession", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSessionByUserCode(s.ctx, "user-code-123").Return(&model.OAuth2DeviceCodeSession{
			Active:    true,
			Signature: abc,
			RequestID: abc,
			ClientID:  "hs256",
			Session:   sessionData,
		}, nil)

		request, err := s.store.GetDeviceCodeSessionByUserCode(s.ctx, "user-code-123", session)

		assert.NoError(t, err)
		assert.NotNil(t, request)
	})

	s.T().Run("ShouldErrWhenInactive", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSessionByUserCode(s.ctx, "user-code-123").Return(&model.OAuth2DeviceCodeSession{
			Active:    false,
			Signature: abc,
		}, nil)

		request, err := s.store.GetDeviceCodeSessionByUserCode(s.ctx, "user-code-123", session)

		assert.Error(t, err)
		assert.Nil(t, request)
	})

	s.T().Run("ShouldErrOnStorageFailure", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSessionByUserCode(s.ctx, "user-code-123").Return(nil, fmt.Errorf("load error"))

		request, err := s.store.GetDeviceCodeSessionByUserCode(s.ctx, "user-code-123", session)

		assert.EqualError(t, err, "load error")
		assert.Nil(t, request)
	})

	s.T().Run("ShouldCreateDefaultSessionWhenNil", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSessionByUserCode(s.ctx, "user-code-123").Return(&model.OAuth2DeviceCodeSession{
			Active:    true,
			Signature: abc,
			RequestID: abc,
			ClientID:  "hs256",
			Session:   sessionData,
		}, nil)

		request, err := s.store.GetDeviceCodeSessionByUserCode(s.ctx, "user-code-123", nil)

		assert.NoError(t, err)
		assert.NotNil(t, request)
	})
}

func (s *StoreSuite) TestSerializationFailureMapping() {
	errBusy := fmt.Errorf("error accessing storage: %w", sqlite3.Error{Code: sqlite3.ErrBusy})

	testCases := []struct {
		name  string
		setup func()
		do    func() (err error)
	}{
		{
			name: "ShouldMapCommit",
			setup: func() {
				s.mock.EXPECT().Commit(s.ctx).Return(errBusy)
			},
			do: func() (err error) {
				return s.store.Commit(s.ctx)
			},
		},
		{
			name: "ShouldMapGetRefreshTokenSession",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, abc).Return(nil, errBusy)
			},
			do: func() (err error) {
				_, err = s.store.GetRefreshTokenSession(s.ctx, abc, nil)

				return err
			},
		},
		{
			name: "ShouldMapRotateRefreshToken",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(nil, errBusy)
			},
			do: func() (err error) {
				return s.store.RotateRefreshToken(s.ctx, abc, abc)
			},
		},
		{
			name: "ShouldMapRotateRefreshTokenWithGracePeriod",
			setup: func() {
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionState(s.ctx, abc).Return(nil, errBusy)
			},
			do: func() (err error) {
				return s.store.RotateRefreshTokenWithGracePeriod(s.ctx, abc, abc, time.Now().Add(time.Minute))
			},
		},
		{
			name: "ShouldMapRevokeRefreshToken",
			setup: func() {
				s.mock.EXPECT().DeactivateOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeRefreshToken, abc).Return(errBusy)
			},
			do: func() (err error) {
				return s.store.RevokeRefreshToken(s.ctx, abc)
			},
		},
		{
			name: "ShouldMapDeleteAccessTokenSession",
			setup: func() {
				s.mock.EXPECT().RevokeOAuth2Session(s.ctx, storage.OAuth2SessionTypeAccessToken, abc).Return(errBusy)
			},
			do: func() (err error) {
				return s.store.DeleteAccessTokenSession(s.ctx, abc)
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			err := tc.do()

			assert.EqualError(t, err, "The request could not be completed due to concurrent access: error accessing storage: database is locked")
			assert.ErrorIs(t, err, oauthelia2.ErrSerializationFailure)
		})
	}
}

func (s *StoreSuite) TestSerializationFailureMappingShouldNotAffectOtherErrors() {
	s.mock.EXPECT().DeactivateOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeRefreshToken, abc).Return(fmt.Errorf("deactivate error"))

	err := s.store.RevokeRefreshToken(s.ctx, abc)

	assert.EqualError(s.T(), err, "deactivate error")
	assert.NotErrorIs(s.T(), err, oauthelia2.ErrSerializationFailure)
}

func (s *StoreSuite) TestCheckAndSetDPoPProofUsed() {
	exp := time.Unix(1700000000, 0).UTC()

	htu := "https://auth.example.com/api/oidc/token"

	proof := model.NewOAuth2DPoPProof("jti", "jkt", "POST", htu, exp)

	s.T().Run("ShouldReturnUnusedForFreshProof", func(t *testing.T) {
		s.mock.EXPECT().CheckAndSetOAuth2DPoPProofUsed(s.ctx, proof, gomock.Any()).Return(false, nil)

		used, err := s.store.CheckAndSetDPoPProofUsed(s.ctx, "jti", "jkt", "nonce", "POST", htu, exp)

		assert.NoError(t, err)
		assert.False(t, used)
	})

	s.T().Run("ShouldReturnUsedForReplayedProof", func(t *testing.T) {
		s.mock.EXPECT().CheckAndSetOAuth2DPoPProofUsed(s.ctx, proof, gomock.Any()).Return(true, nil)

		used, err := s.store.CheckAndSetDPoPProofUsed(s.ctx, "jti", "jkt", "nonce", "POST", htu, exp)

		assert.NoError(t, err)
		assert.True(t, used)
	})

	s.T().Run("ShouldErrOnStorageFailure", func(t *testing.T) {
		s.mock.EXPECT().CheckAndSetOAuth2DPoPProofUsed(s.ctx, proof, gomock.Any()).Return(false, fmt.Errorf("check error"))

		used, err := s.store.CheckAndSetDPoPProofUsed(s.ctx, "jti", "jkt", "nonce", "POST", htu, exp)

		assert.EqualError(t, err, "check error")
		assert.False(t, used)
	})

	s.T().Run("ShouldKeyOnThumbprint", func(t *testing.T) {
		s.mock.EXPECT().CheckAndSetOAuth2DPoPProofUsed(s.ctx, model.NewOAuth2DPoPProof("jti", "other", "POST", htu, exp), gomock.Any()).Return(false, nil)

		used, err := s.store.CheckAndSetDPoPProofUsed(s.ctx, "jti", "other", "nonce", "POST", htu, exp)

		assert.NoError(t, err)
		assert.False(t, used)
	})

	s.T().Run("ShouldNotKeyOnNonce", func(t *testing.T) {
		s.mock.EXPECT().CheckAndSetOAuth2DPoPProofUsed(s.ctx, proof, gomock.Any()).Return(true, nil)

		used, err := s.store.CheckAndSetDPoPProofUsed(s.ctx, "jti", "jkt", "other", "POST", htu, exp)

		assert.NoError(t, err)
		assert.True(t, used)
	})
}

func (s *StoreSuite) TestCreateDPoPNonce() {
	exp := time.Unix(1700000000, 0).UTC()

	nonce := model.OAuth2DPoPNonce{Signature: model.NewOAuth2DPoPNonceSignature("nonce"), ExpiresAt: exp}

	s.T().Run("ShouldSaveNonce", func(t *testing.T) {
		s.mock.EXPECT().SaveOAuth2DPoPNonce(s.ctx, nonce).Return(nil)

		assert.NoError(t, s.store.CreateDPoPNonce(s.ctx, "nonce", exp))
	})

	s.T().Run("ShouldErrOnStorageFailure", func(t *testing.T) {
		s.mock.EXPECT().SaveOAuth2DPoPNonce(s.ctx, nonce).Return(fmt.Errorf("save error"))

		assert.EqualError(t, s.store.CreateDPoPNonce(s.ctx, "nonce", exp), "save error")
	})
}

func (s *StoreSuite) TestIsDPoPNonceValid() {
	signature := model.NewOAuth2DPoPNonceSignature("nonce")

	s.T().Run("ShouldBeValidWhenUnexpired", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DPoPNonce(s.ctx, signature).Return(&model.OAuth2DPoPNonce{Signature: signature, ExpiresAt: time.Now().Add(time.Minute)}, nil)

		valid, err := s.store.IsDPoPNonceValid(s.ctx, "nonce")

		assert.NoError(t, err)
		assert.True(t, valid)
	})

	s.T().Run("ShouldNotBeValidWhenExpired", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DPoPNonce(s.ctx, signature).Return(&model.OAuth2DPoPNonce{Signature: signature, ExpiresAt: time.Now().Add(-time.Minute)}, nil)

		valid, err := s.store.IsDPoPNonceValid(s.ctx, "nonce")

		assert.NoError(t, err)
		assert.False(t, valid)
	})

	s.T().Run("ShouldNotBeValidWhenUnknown", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DPoPNonce(s.ctx, signature).Return(nil, fmt.Errorf("error selecting oauth2 dpop nonce with signature '%s': %w", signature, sql.ErrNoRows))

		valid, err := s.store.IsDPoPNonceValid(s.ctx, "nonce")

		assert.NoError(t, err)
		assert.False(t, valid)
	})

	s.T().Run("ShouldErrOnStorageFailure", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DPoPNonce(s.ctx, signature).Return(nil, fmt.Errorf("load error"))

		valid, err := s.store.IsDPoPNonceValid(s.ctx, "nonce")

		assert.EqualError(t, err, "load error")
		assert.False(t, valid)
	})
}

func TestStoreTokenExchangeCustomJWTUnsupported(t *testing.T) {
	store := &oidc.Store{}

	err := store.SetTokenExchangeCustomJWT(context.TODO(), "https://issuer.example.com", "jti", time.Now().Add(time.Hour))

	require.Error(t, err)
	assert.ErrorIs(t, err, oauthelia2.ErrInvalidRequest)
	assert.Contains(t, oauthelia2.ErrorToRFC6749Error(err).HintField, oidc.TokenTypeJWT)
}

func TestStoreGetSubjectForTokenExchangeUnsupported(t *testing.T) {
	store := &oidc.Store{}

	subject, err := store.GetSubjectForTokenExchange(context.TODO(), nil, map[string]any{})

	assert.Empty(t, subject)
	require.Error(t, err)
	assert.ErrorIs(t, err, oauthelia2.ErrInvalidRequest)
	assert.Contains(t, oauthelia2.ErrorToRFC6749Error(err).HintField, oidc.TokenTypeJWT)
}
