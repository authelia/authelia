// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

	"github.com/authelia/authelia/v4/internal/authorization"
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
	}, nil)

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
	}, nil)

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
	}, nil)

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
	}, nil)

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
	}, provider)

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
	}, provider)

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

func TestCreateDeviceCodeSessionShouldRefuseDuplicateUserCode(t *testing.T) {
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
						Scopes:              []string{oidc.ScopeOpenID},
						Secret:              tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}, provider)

	testCases := []struct {
		name      string
		signature string
		userCode  string
		err       error
	}{
		{name: "ShouldCreateFirstSession", signature: "sig-device-dup-1", userCode: "sig-user-dup", err: nil},
		{name: "ShouldRefuseDuplicateUserCode", signature: "sig-device-dup-2", userCode: "sig-user-dup", err: oauthelia2.ErrDuplicateUserCode},
		{name: "ShouldCreateUniqueUserCode", signature: "sig-device-dup-3", userCode: "sig-user-unique", err: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requester := &oauthelia2.DeviceAuthorizeRequest{
				Request: oauthelia2.Request{
					ID:             tc.signature,
					Client:         &oidc.RegisteredClient{ID: myclient},
					RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID},
					Session:        &oidc.Session{},
				},
				DeviceCodeSignature: tc.signature,
				UserCodeSignature:   tc.userCode,
			}

			err := s.CreateDeviceCodeSession(ctx, tc.signature, requester)

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.err)
			}
		})
	}
}

func TestSetClientAssertionJWTShouldRefuseKnownJTIForClient(t *testing.T) {
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
			},
		},
	}, provider)

	testCases := []struct {
		name     string
		clientID string
		jti      string
		exp      time.Time
		err      error
	}{
		{name: "ShouldSetFirstUse", clientID: "client-a", jti: "jti-1", exp: time.Now().Add(time.Hour), err: nil},
		{name: "ShouldRefuseSecondUse", clientID: "client-a", jti: "jti-1", exp: time.Now().Add(time.Hour), err: oauthelia2.ErrJTIKnown},
		{name: "ShouldSetSameJTIForAnotherClient", clientID: "client-b", jti: "jti-1", exp: time.Now().Add(time.Hour), err: nil},
		{name: "ShouldSetAnotherJTI", clientID: "client-a", jti: "jti-2", exp: time.Now().Add(time.Hour), err: nil},
		{name: "ShouldSetExpired", clientID: "client-a", jti: "jti-3", exp: time.Now().Add(-time.Hour), err: nil},
		{name: "ShouldSetOverExpired", clientID: "client-a", jti: "jti-3", exp: time.Now().Add(time.Hour), err: nil},
		{name: "ShouldRefuseOverUnexpired", clientID: "client-a", jti: "jti-3", exp: time.Now().Add(time.Hour), err: oauthelia2.ErrJTIKnown},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.SetClientAssertionJWT(ctx, tc.clientID, tc.jti, tc.exp)

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.err)
			}
		})
	}

	assert.ErrorIs(t, s.ClientAssertionJWTValid(ctx, "client-a", "jti-1"), oauthelia2.ErrJTIKnown)
	assert.ErrorIs(t, s.ClientAssertionJWTValid(ctx, "client-b", "jti-1"), oauthelia2.ErrJTIKnown)
	assert.NoError(t, s.ClientAssertionJWTValid(ctx, "client-b", "jti-2"))
	assert.NoError(t, s.ClientAssertionJWTValid(ctx, "client-c", "jti-1"))
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
	}, s.mock)
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
			LoadOAuth2BlacklistedJTI(s.ctx, "3ff686f228cc5f1f480c78ccbfe6193c28d3bbc258a46f5858b1fecff7025111").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "3ff686f228cc5f1f480c78ccbfe6193c28d3bbc258a46f5858b1fecff7025111",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "7dab5206b9812435edf29885482e066817a006498e63449a580ad9f797cdcb8a").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "7dab5206b9812435edf29885482e066817a006498e63449a580ad9f797cdcb8a",
				ExpiresAt: time.Now().Add(-time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "afcbd1e85207005b9b26671765dad1f810c2154674a40e026fbaf8e5667f4fe2").
			Return(nil, fmt.Errorf("failed to load")),
	)

	s.EqualError(s.store.ClientAssertionJWTValid(s.ctx, "example", "066ee771-e156-4886-b99f-ee09b0d3edf4"), "jti_known")
	s.NoError(s.store.ClientAssertionJWTValid(s.ctx, "example", "5dad3ff7-e4f2-41b6-98a3-b73d872076ce"))
	s.EqualError(s.store.ClientAssertionJWTValid(s.ctx, "example", "65471ccb-d650-4006-a95f-cb4f4e3d7202"), "failed to load")
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
			Return(storage.ErrNoRowsAffected),
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
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "1").
			Return(nil),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionAccessSignature(s.ctx, "1").
			Return("at_paired_1", nil),
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
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "2").
			Return(nil),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionAccessSignature(s.ctx, "2").
			Return("", nil),
		s.mock.
			EXPECT().
			RevokeOAuth2SessionByRequestID(s.ctx, storage.OAuth2SessionTypeAccessToken, "65471ccb-d650-4006-a95f-cb4f4e3d7201").
			Return(fmt.Errorf("not found")),
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "3").
			Return(nil),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionAccessSignature(s.ctx, "3").
			Return("", fmt.Errorf("no refresh token session")),
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "4").
			Return(storage.ErrNoRowsAffected),
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "5").
			Return(fmt.Errorf("deactivate error")),
		s.mock.
			EXPECT().
			DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, "6").
			Return(nil),
		s.mock.
			EXPECT().
			LoadOAuth2RefreshTokenSessionAccessSignature(s.ctx, "6").
			Return("at_paired_6", nil),
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
	s.ErrorIs(s.store.RotateRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7204", "4"), oauthelia2.ErrInactiveToken)
	s.EqualError(s.store.RotateRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7205", "5"), "deactivate error")
	s.NoError(s.store.RotateRefreshToken(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7206", "6"))

	s.NoError(s.store.DeletePKCERequestSession(s.ctx, "pkce1"))
	s.EqualError(s.store.DeletePKCERequestSession(s.ctx, "pkce2"), "not found")

	s.NoError(s.store.DeleteOpenIDConnectSession(s.ctx, "ac_1"))
	s.EqualError(s.store.DeleteOpenIDConnectSession(s.ctx, "ac_2"), "not found")

	s.NoError(s.store.DeletePARSession(s.ctx, "urn:par1"))
	s.EqualError(s.store.DeletePARSession(s.ctx, "urn:par2"), "not found")
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

func (s *StoreSuite) TestIsRFC7523JWTUsed() {
	gomock.InOrder(
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "3ff686f228cc5f1f480c78ccbfe6193c28d3bbc258a46f5858b1fecff7025111").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "3ff686f228cc5f1f480c78ccbfe6193c28d3bbc258a46f5858b1fecff7025111",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "7dab5206b9812435edf29885482e066817a006498e63449a580ad9f797cdcb8a").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "7dab5206b9812435edf29885482e066817a006498e63449a580ad9f797cdcb8a",
				ExpiresAt: time.Now().Add(-time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "afcbd1e85207005b9b26671765dad1f810c2154674a40e026fbaf8e5667f4fe2").
			Return(nil, fmt.Errorf("failed to load")),
	)

	used, err := s.store.IsRFC7523JWTUsed(s.ctx, "example", "066ee771-e156-4886-b99f-ee09b0d3edf4")
	s.True(used)
	s.EqualError(err, "jti_known")

	used, err = s.store.IsRFC7523JWTUsed(s.ctx, "example", "5dad3ff7-e4f2-41b6-98a3-b73d872076ce")
	s.False(used)
	s.NoError(err)

	used, err = s.store.IsRFC7523JWTUsed(s.ctx, "example", "65471ccb-d650-4006-a95f-cb4f4e3d7202")
	s.True(used)
	s.EqualError(err, "failed to load")
}

func (s *StoreSuite) TestMarkRFC7523JWTUsedForTime() {
	gomock.InOrder(
		s.mock.EXPECT().
			SaveOAuth2BlacklistedJTI(s.ctx, model.OAuth2BlacklistedJTI{Signature: "afcbd1e85207005b9b26671765dad1f810c2154674a40e026fbaf8e5667f4fe2", ExpiresAt: time.Unix(160000000, 0)}).
			Return(nil),
		s.mock.EXPECT().SaveOAuth2BlacklistedJTI(s.ctx, model.OAuth2BlacklistedJTI{Signature: "084a787fff055a1c2b023780064045fb342fa390b38c2c07dc784cb679b98752", ExpiresAt: time.Unix(160000000, 0)}).
			Return(fmt.Errorf("already marked")),
	)

	s.NoError(s.store.MarkRFC7523JWTUsedForTime(s.ctx, "example", "65471ccb-d650-4006-a95f-cb4f4e3d7202", time.Unix(160000000, 0)))
	s.EqualError(s.store.MarkRFC7523JWTUsedForTime(s.ctx, "example", "65471ccb-d650-4006-a95f-cb4f4e3d7201", time.Unix(160000000, 0)), "already marked")
}

func (s *StoreSuite) TestUpdateRefreshTokenSession() {
	challenge := model.MustNullUUID(model.NewRandomNullUUID())
	session := &oidc.Session{
		ChallengeID: challenge,
	}
	sessionData, _ := json.Marshal(session)

	expected := model.OAuth2Session{ChallengeID: challenge, RequestID: abc, ClientID: "example", Signature: "rt_update", Active: true, Session: sessionData, RequestedScopes: model.StringSlicePipeDelimited{}, GrantedScopes: model.StringSlicePipeDelimited{}}

	gomock.InOrder(
		s.mock.EXPECT().UpdateOAuth2RefreshTokenSession(s.ctx, expected).Return(nil),
		s.mock.EXPECT().UpdateOAuth2RefreshTokenSession(s.ctx, expected).Return(fmt.Errorf("error updating: %w", storage.ErrNoRowsAffected)),
		s.mock.EXPECT().UpdateOAuth2RefreshTokenSession(s.ctx, expected).Return(fmt.Errorf("update error")),
	)

	request := &oauthelia2.Request{
		ID:      abc,
		Client:  &oidc.RegisteredClient{ID: "example"},
		Session: session,
	}

	s.NoError(s.store.UpdateRefreshTokenSession(s.ctx, "rt_update", request))
	s.ErrorIs(s.store.UpdateRefreshTokenSession(s.ctx, "rt_update", request), oauthelia2.ErrNotFound)
	s.EqualError(s.store.UpdateRefreshTokenSession(s.ctx, "rt_update", request), "update error")
	s.EqualError(s.store.UpdateRefreshTokenSession(s.ctx, "rt_update", nil), "failed to create new *model.OAuth2Session: the oauthelia2.Requester was nil")
}

func (s *StoreSuite) TestDeletePARSessionShouldReturnNotFoundWhenAlreadyDeleted() {
	s.mock.EXPECT().
		RevokeOAuth2PushedAuthorizationSession(s.ctx, "urn:par-redeemed").
		Return(fmt.Errorf("error revoking: %w", storage.ErrNoRowsAffected))

	s.ErrorIs(s.store.DeletePARSession(s.ctx, "urn:par-redeemed"), oauthelia2.ErrNotFound)
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

	s.T().Run("ShouldReturnRequestWithErrWhenInactive", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSession(s.ctx, abc).Return(&model.OAuth2DeviceCodeSession{
			Active:    false,
			Signature: abc,
			RequestID: abc,
			ClientID:  "hs256",
			Session:   sessionData,
		}, nil)

		request, err := s.store.GetDeviceCodeSession(s.ctx, abc, session)

		assert.ErrorIs(t, err, oauthelia2.ErrInvalidatedDeviceCode)
		require.NotNil(t, request)
		assert.Equal(t, abc, request.GetID())
		assert.Equal(t, abc, request.GetDeviceCodeSignature())
	})

	s.T().Run("ShouldErrNotFoundWhenMissing", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSession(s.ctx, abc).Return(nil, fmt.Errorf("error selecting oauth2 device code session: %w", sql.ErrNoRows))

		request, err := s.store.GetDeviceCodeSession(s.ctx, abc, session)

		assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
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

	s.T().Run("ShouldErrInvalidatedWhenAlreadyInactive", func(t *testing.T) {
		s.mock.EXPECT().DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeDeviceAuthorizeCode, abc).Return(fmt.Errorf("error deactivating: %w", storage.ErrNoRowsAffected))

		err := s.store.InvalidateDeviceCodeSession(s.ctx, abc)

		assert.ErrorIs(t, err, oauthelia2.ErrInvalidatedDeviceCode)
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

	s.T().Run("ShouldReturnRequestWithErrWhenInactive", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSessionByUserCode(s.ctx, "user-code-123").Return(&model.OAuth2DeviceCodeSession{
			Active:    false,
			Signature: abc,
			RequestID: abc,
			ClientID:  "hs256",
			Session:   sessionData,
		}, nil)

		request, err := s.store.GetDeviceCodeSessionByUserCode(s.ctx, "user-code-123", session)

		assert.ErrorIs(t, err, oauthelia2.ErrInvalidatedDeviceCode)
		require.NotNil(t, request)
		assert.Equal(t, abc, request.GetID())
		assert.Equal(t, abc, request.GetDeviceCodeSignature())
	})

	s.T().Run("ShouldErrNotFoundWhenMissing", func(t *testing.T) {
		s.mock.EXPECT().LoadOAuth2DeviceCodeSessionByUserCode(s.ctx, "user-code-123").Return(nil, fmt.Errorf("error selecting oauth2 device code session: %w", sql.ErrNoRows))

		request, err := s.store.GetDeviceCodeSessionByUserCode(s.ctx, "user-code-123", session)

		assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
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
				s.mock.EXPECT().DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, abc).Return(errBusy)
			},
			do: func() (err error) {
				return s.store.RotateRefreshToken(s.ctx, abc, abc)
			},
		},
		{
			name: "ShouldMapRotateRefreshTokenAccessSignature",
			setup: func() {
				s.mock.EXPECT().DeactivateOAuth2Session(s.ctx, storage.OAuth2SessionTypeRefreshToken, abc).Return(nil)
				s.mock.EXPECT().LoadOAuth2RefreshTokenSessionAccessSignature(s.ctx, abc).Return("", errBusy)
			},
			do: func() (err error) {
				return s.store.RotateRefreshToken(s.ctx, abc, abc)
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
