package oidc

import (
	"context"
	"errors"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/storage"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
)

// NewOpenIDConnectStore returns a new OpenIDConnectStore using the provided schema.OpenIDConnectConfiguration.
func NewOpenIDConnectStore(configuration *schema.OpenIDConnectConfiguration) (store *OpenIDConnectStore) {
	clients := make(map[string]fosite.Client)

	for _, clientConf := range configuration.Clients {
		policy := authorization.PolicyToLevel(clientConf.Policy)
		logging.Logger().Debugf("Registering client %s with policy %s (%v)", clientConf.ID, clientConf.Policy, policy)

		client := &InternalClient{
			ID:            clientConf.ID,
			Description:   clientConf.Description,
			Policy:        authorization.PolicyToLevel(clientConf.Policy),
			Secret:        []byte(clientConf.Secret),
			RedirectURIs:  clientConf.RedirectURIs,
			GrantTypes:    clientConf.GrantTypes,
			ResponseTypes: clientConf.ResponseTypes,
			Scopes:        clientConf.Scopes,
		}

		clients[client.ID] = client
	}

	return &OpenIDConnectStore{
		memory: &storage.MemoryStore{
			IDSessions:             make(map[string]fosite.Requester),
			Users:                  map[string]storage.MemoryUserRelation{},
			AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
			AccessTokens:           map[string]fosite.Requester{},
			RefreshTokens:          map[string]storage.StoreRefreshToken{},
			PKCES:                  map[string]fosite.Requester{},
			AccessTokenRequestIDs:  map[string]string{},
			RefreshTokenRequestIDs: map[string]string{},
		},
	}
}

// OpenIDConnectStore is Authelia's internal representation of the fosite.Storage interface.
//
//	Currently it is mostly just implementing a decorator pattern other then GetInternalClient.
//	The long term plan is to have these methods interact with the Authelia storage and
//	session providers where applicable.
type OpenIDConnectStore struct {
	clients map[string]*InternalClient
	memory  *storage.MemoryStore
}

// GetClientPolicy retrieves the policy from the client with the matching provided id.
func (s OpenIDConnectStore) GetClientPolicy(id string) (level authorization.Level) {
	client, err := s.GetInternalClient(id)
	if err != nil {
		return authorization.TwoFactor
	}

	return client.Policy
}

// GetInternalClient returns a fosite.Client asserted as an InternalClient matching the provided id.
func (s OpenIDConnectStore) GetInternalClient(id string) (client *InternalClient, err error) {
	client, ok := s.clients[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return client, nil
}

// IsValidClientID returns true if the provided id exists in the OpenIDConnectProvider.Clients map.
func (s OpenIDConnectStore) IsValidClientID(id string) (valid bool) {
	_, err := s.GetInternalClient(id)

	return err == nil
}

// CreateOpenIDConnectSession decorates fosite's storage.MemoryStore CreateOpenIDConnectSession method.
func (s *OpenIDConnectStore) CreateOpenIDConnectSession(ctx context.Context, authorizeCode string, requester fosite.Requester) error {
	return s.memory.CreateOpenIDConnectSession(ctx, authorizeCode, requester)
}

// GetOpenIDConnectSession decorates fosite's storage.MemoryStore GetOpenIDConnectSession method.
func (s *OpenIDConnectStore) GetOpenIDConnectSession(ctx context.Context, authorizeCode string, requester fosite.Requester) (fosite.Requester, error) {
	return s.memory.GetOpenIDConnectSession(ctx, authorizeCode, requester)
}

// DeleteOpenIDConnectSession decorates fosite's storage.MemoryStore DeleteOpenIDConnectSession method.
func (s *OpenIDConnectStore) DeleteOpenIDConnectSession(ctx context.Context, authorizeCode string) error {
	return s.memory.DeleteOpenIDConnectSession(ctx, authorizeCode)
}

// GetClient decorates fosite's storage.MemoryStore GetClient method.
func (s *OpenIDConnectStore) GetClient(_ context.Context, id string) (fosite.Client, error) {
	return s.GetInternalClient(id)
}

// ClientAssertionJWTValid decorates fosite's storage.MemoryStore ClientAssertionJWTValid method.
func (s *OpenIDConnectStore) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	return s.memory.ClientAssertionJWTValid(ctx, jti)
}

// SetClientAssertionJWT decorates fosite's storage.MemoryStore SetClientAssertionJWT method.
func (s *OpenIDConnectStore) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	return s.memory.SetClientAssertionJWT(ctx, jti, exp)
}

// CreateAuthorizeCodeSession decorates fosite's storage.MemoryStore CreateAuthorizeCodeSession method.
func (s *OpenIDConnectStore) CreateAuthorizeCodeSession(ctx context.Context, code string, req fosite.Requester) error {
	return s.memory.CreateAuthorizeCodeSession(ctx, code, req)
}

// GetAuthorizeCodeSession decorates fosite's storage.MemoryStore GetAuthorizeCodeSession method.
func (s *OpenIDConnectStore) GetAuthorizeCodeSession(ctx context.Context, code string, session fosite.Session) (fosite.Requester, error) {
	return s.memory.GetAuthorizeCodeSession(ctx, code, session)
}

// InvalidateAuthorizeCodeSession decorates fosite's storage.MemoryStore InvalidateAuthorizeCodeSession method.
func (s *OpenIDConnectStore) InvalidateAuthorizeCodeSession(ctx context.Context, code string) error {
	return s.memory.InvalidateAuthorizeCodeSession(ctx, code)
}

// CreatePKCERequestSession decorates fosite's storage.MemoryStore CreatePKCERequestSession method.
func (s *OpenIDConnectStore) CreatePKCERequestSession(ctx context.Context, code string, req fosite.Requester) error {
	return s.memory.CreatePKCERequestSession(ctx, code, req)
}

// GetPKCERequestSession decorates fosite's storage.MemoryStore GetPKCERequestSession method.
func (s *OpenIDConnectStore) GetPKCERequestSession(ctx context.Context, code string, session fosite.Session) (fosite.Requester, error) {
	return s.memory.GetPKCERequestSession(ctx, code, session)
}

// DeletePKCERequestSession decorates fosite's storage.MemoryStore DeletePKCERequestSession method.
func (s *OpenIDConnectStore) DeletePKCERequestSession(ctx context.Context, code string) error {
	return s.memory.DeletePKCERequestSession(ctx, code)
}

// CreateAccessTokenSession decorates fosite's storage.MemoryStore CreateAccessTokenSession method.
func (s *OpenIDConnectStore) CreateAccessTokenSession(ctx context.Context, signature string, req fosite.Requester) error {
	return s.memory.CreateAccessTokenSession(ctx, signature, req)
}

// GetAccessTokenSession decorates fosite's storage.MemoryStore GetAccessTokenSession method.
func (s *OpenIDConnectStore) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return s.memory.GetAccessTokenSession(ctx, signature, session)
}

// DeleteAccessTokenSession decorates fosite's storage.MemoryStore DeleteAccessTokenSession method.
func (s *OpenIDConnectStore) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	return s.memory.DeleteAccessTokenSession(ctx, signature)
}

// CreateRefreshTokenSession decorates fosite's storage.MemoryStore CreateRefreshTokenSession method.
func (s *OpenIDConnectStore) CreateRefreshTokenSession(ctx context.Context, signature string, req fosite.Requester) error {
	return s.memory.CreateRefreshTokenSession(ctx, signature, req)
}

// GetRefreshTokenSession decorates fosite's storage.MemoryStore GetRefreshTokenSession method.
func (s *OpenIDConnectStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return s.memory.GetRefreshTokenSession(ctx, signature, session)
}

// DeleteRefreshTokenSession decorates fosite's storage.MemoryStore DeleteRefreshTokenSession method.
func (s *OpenIDConnectStore) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	return s.memory.DeleteRefreshTokenSession(ctx, signature)
}

// Authenticate decorates fosite's storage.MemoryStore Authenticate method.
func (s *OpenIDConnectStore) Authenticate(ctx context.Context, name string, secret string) error {
	return s.memory.Authenticate(ctx, name, secret)
}

// RevokeRefreshToken decorates fosite's storage.MemoryStore RevokeRefreshToken method.
func (s *OpenIDConnectStore) RevokeRefreshToken(ctx context.Context, requestID string) error {
	return s.memory.RevokeRefreshToken(ctx, requestID)
}

// RevokeAccessToken decorates fosite's storage.MemoryStore RevokeAccessToken method.
func (s *OpenIDConnectStore) RevokeAccessToken(ctx context.Context, requestID string) error {
	return s.memory.RevokeAccessToken(ctx, requestID)
}

// GetPublicKey decorates fosite's storage.MemoryStore GetPublicKey method.
func (s *OpenIDConnectStore) GetPublicKey(ctx context.Context, issuer string, subject string, keyID string) (*jose.JSONWebKey, error) {
	return s.memory.GetPublicKey(ctx, issuer, subject, keyID)
}

// GetPublicKeys decorates fosite's storage.MemoryStore GetPublicKeys method.
func (s *OpenIDConnectStore) GetPublicKeys(ctx context.Context, issuer string, subject string) (*jose.JSONWebKeySet, error) {
	return s.memory.GetPublicKeys(ctx, issuer, subject)
}

// GetPublicKeyScopes decorates fosite's storage.MemoryStore GetPublicKeyScopes method.
func (s *OpenIDConnectStore) GetPublicKeyScopes(ctx context.Context, issuer string, subject string, keyID string) ([]string, error) {
	return s.memory.GetPublicKeyScopes(ctx, issuer, subject, keyID)
}

// IsJWTUsed decorates fosite's storage.MemoryStore IsJWTUsed method.
func (s *OpenIDConnectStore) IsJWTUsed(ctx context.Context, jti string) (bool, error) {
	return s.memory.IsJWTUsed(ctx, jti)
}

// MarkJWTUsedForTime decorates fosite's storage.MemoryStore MarkJWTUsedForTime method.
func (s *OpenIDConnectStore) MarkJWTUsedForTime(ctx context.Context, jti string, exp time.Time) error {
	return s.memory.SetClientAssertionJWT(ctx, jti, exp)
}
