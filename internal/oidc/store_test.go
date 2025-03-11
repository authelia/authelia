package oidc_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

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
						Name:                myclientdesc,
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
		Name:                myclientdesc,
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
	assert.Equal(t, myclientdesc, client.GetName())
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
		Name:                myclientdesc,
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
						Name:                myclientdesc,
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
			LoadOAuth2BlacklistedJTI(s.ctx, "3a240379e8286a7a8ff5e99d68567e0e5e34e80168b8feffa89d3d33dea95b63").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "3a240379e8286a7a8ff5e99d68567e0e5e34e80168b8feffa89d3d33dea95b63",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "e7f67ad76c80d57d34b19598462817932aec21d2806a08a786a8d4b9dd476068").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "e7f67ad76c80d57d34b19598462817932aec21d2806a08a786a8d4b9dd476068",
				ExpiresAt: time.Now().Add(-time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "f29ef0d85303a09411b76001c579980f1b1b7fc9deb1fa647875a724f4f231c6").
			Return(nil, fmt.Errorf("failed to load")),
	)

	s.EqualError(s.store.ClientAssertionJWTValid(s.ctx, "066ee771-e156-4886-b99f-ee09b0d3edf4"), "jti_known")
	s.NoError(s.store.ClientAssertionJWTValid(s.ctx, "5dad3ff7-e4f2-41b6-98a3-b73d872076ce"))
	s.EqualError(s.store.ClientAssertionJWTValid(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7202"), "failed to load")
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
			SaveOAuth2PARContext(s.ctx, model.OAuth2PARContext{Signature: abc, RequestID: abc, ClientID: "example", Session: sessionData}).
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

	s.NoError(s.store.CreateRefreshTokenSession(s.ctx, abc, &oauthelia2.Request{
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
			RevokeOAuth2PARContext(s.ctx, "urn:par1").
			Return(nil),
		s.mock.
			EXPECT().
			RevokeOAuth2PARContext(s.ctx, "urn:par2").
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

	s.NoError(s.store.RevokeRefreshTokenMaybeGracePeriod(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7200", "1"))
	s.EqualError(s.store.RevokeRefreshTokenMaybeGracePeriod(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7201", "2"), "not found")
	s.EqualError(s.store.RevokeRefreshTokenMaybeGracePeriod(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7202", "3"), "sql: no rows in result set")

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
			LoadOAuth2PARContext(s.ctx, "urn:par").
			Return(&model.OAuth2PARContext{Signature: abc, RequestID: abc, ClientID: "hs256", Session: sessionData}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2PARContext(s.ctx, "urn:par").
			Return(nil, sql.ErrNoRows),
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
	s.EqualError(err, "sql: no rows in result set")
}

func (s *StoreSuite) TestIsJWTUsed() {
	gomock.InOrder(
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "3a240379e8286a7a8ff5e99d68567e0e5e34e80168b8feffa89d3d33dea95b63").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "3a240379e8286a7a8ff5e99d68567e0e5e34e80168b8feffa89d3d33dea95b63",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "e7f67ad76c80d57d34b19598462817932aec21d2806a08a786a8d4b9dd476068").
			Return(&model.OAuth2BlacklistedJTI{
				ID:        1,
				Signature: "e7f67ad76c80d57d34b19598462817932aec21d2806a08a786a8d4b9dd476068",
				ExpiresAt: time.Now().Add(-time.Hour),
			}, nil),
		s.mock.
			EXPECT().
			LoadOAuth2BlacklistedJTI(s.ctx, "f29ef0d85303a09411b76001c579980f1b1b7fc9deb1fa647875a724f4f231c6").
			Return(nil, fmt.Errorf("failed to load")),
	)

	used, err := s.store.IsJWTUsed(s.ctx, "066ee771-e156-4886-b99f-ee09b0d3edf4")
	s.True(used)
	s.EqualError(err, "jti_known")

	used, err = s.store.IsJWTUsed(s.ctx, "5dad3ff7-e4f2-41b6-98a3-b73d872076ce")
	s.False(used)
	s.NoError(err)

	used, err = s.store.IsJWTUsed(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7202")
	s.True(used)
	s.EqualError(err, "failed to load")
}

func (s *StoreSuite) TestMarkJWTUsedForTime() {
	gomock.InOrder(
		s.mock.EXPECT().
			SaveOAuth2BlacklistedJTI(s.ctx, model.OAuth2BlacklistedJTI{Signature: "f29ef0d85303a09411b76001c579980f1b1b7fc9deb1fa647875a724f4f231c6", ExpiresAt: time.Unix(160000000, 0)}).
			Return(nil),
		s.mock.EXPECT().SaveOAuth2BlacklistedJTI(s.ctx, model.OAuth2BlacklistedJTI{Signature: "0dab0de97ed4e05da82763497448daf4f6b555c99218100e3ef5a81f36232940", ExpiresAt: time.Unix(160000000, 0)}).
			Return(fmt.Errorf("already marked")),
	)

	s.NoError(s.store.MarkJWTUsedForTime(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7202", time.Unix(160000000, 0)))
	s.EqualError(s.store.MarkJWTUsedForTime(s.ctx, "65471ccb-d650-4006-a95f-cb4f4e3d7201", time.Unix(160000000, 0)), "already marked")
}
