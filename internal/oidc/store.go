package oidc

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func NewOpenIDConnectStore(config *schema.OpenIDConnectConfiguration, provider storage.Provider) (store *OpenIDConnectStore) {
	logger := logging.Logger()

	store = &OpenIDConnectStore{
		provider: provider,
		clients:  map[string]*Client{},
	}

	for _, client := range config.Clients {
		policy := authorization.PolicyToLevel(client.Policy)
		logger.Debugf("Registering client %s with policy %s (%v)", client.ID, client.Policy, policy)

		store.clients[client.ID] = NewClient(client)
	}

	return store
}

// GetClientPolicy retrieves the policy from the client with the matching provided id.
func (s OpenIDConnectStore) GetClientPolicy(id string) (level authorization.Level) {
	client, err := s.GetFullClient(id)
	if err != nil {
		return authorization.TwoFactor
	}

	return client.Policy
}

// GetFullClient returns a fosite.Client asserted as an Client matching the provided id.
func (s OpenIDConnectStore) GetFullClient(id string) (client *Client, err error) {
	client, ok := s.clients[id]
	if !ok {
		return nil, fosite.ErrNotFound
	}

	return client, nil
}

// IsValidClientID returns true if the provided id exists in the OpenIDConnectProvider.Clients map.
func (s OpenIDConnectStore) IsValidClientID(id string) (valid bool) {
	_, err := s.GetFullClient(id)

	return err == nil
}

func (s *OpenIDConnectStore) GetClient(_ context.Context, id string) (client fosite.Client, err error) {
	return s.GetFullClient(id)
}

func (s *OpenIDConnectStore) BeginTX(ctx context.Context) (c context.Context, err error) {
	return s.provider.BeginTX(ctx)
}

func (s *OpenIDConnectStore) Commit(ctx context.Context) (err error) {
	return s.provider.Commit(ctx)
}

func (s *OpenIDConnectStore) Rollback(ctx context.Context) (err error) {
	return s.provider.Rollback(ctx)
}

func (s *OpenIDConnectStore) ClientAssertionJWTValid(ctx context.Context, jti string) (err error) {
	signature := fmt.Sprintf("%x", sha256.Sum256([]byte(jti)))

	blacklistedJTI, err := s.provider.LoadOAuth2BlacklistedJTI(ctx, signature)

	switch {
	case errors.Is(sql.ErrNoRows, err):
		return nil
	case err != nil:
		return err
	case blacklistedJTI.ExpiresAt.After(time.Now()):
		return fosite.ErrJTIKnown
	default:
		return nil
	}
}

func (s *OpenIDConnectStore) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) (err error) {
	blacklistedJTI := model.NewOAuth2BlacklistedJTI(jti, exp)

	return s.provider.SaveOAuth2BlacklistedJTI(ctx, blacklistedJTI)
}

func (s *OpenIDConnectStore) CreateAuthorizeCodeSession(ctx context.Context, code string, request fosite.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypeAuthorizeCode, code, request)
}

func (s *OpenIDConnectStore) InvalidateAuthorizeCodeSession(ctx context.Context, code string) (err error) {
	return s.revokeSessionBySignature(ctx, storage.OAuth2SessionTypeAuthorizeCode, code)
}

func (s *OpenIDConnectStore) GetAuthorizeCodeSession(ctx context.Context, code string, session fosite.Session) (request fosite.Requester, err error) {
	return s.loadSessionBySignature(ctx, storage.OAuth2SessionTypeAuthorizeCode, code, session)
}

func (s *OpenIDConnectStore) CreateAccessTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypeAccessToken, signature, request)
}

func (s *OpenIDConnectStore) DeleteAccessTokenSession(ctx context.Context, signature string) (err error) {
	return s.revokeSessionBySignature(ctx, storage.OAuth2SessionTypeAccessToken, signature)
}

func (s *OpenIDConnectStore) RevokeAccessToken(ctx context.Context, requestID string) (err error) {
	return s.revokeSessionByRequestID(ctx, storage.OAuth2SessionTypeAccessToken, requestID)
}

func (s *OpenIDConnectStore) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	return s.loadSessionBySignature(ctx, storage.OAuth2SessionTypeAccessToken, signature, session)
}

func (s *OpenIDConnectStore) CreateRefreshTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypeRefreshToken, signature, request)
}

func (s *OpenIDConnectStore) DeleteRefreshTokenSession(ctx context.Context, signature string) (err error) {
	return s.revokeSessionBySignature(ctx, storage.OAuth2SessionTypeRefreshToken, signature)
}

func (s *OpenIDConnectStore) RevokeRefreshToken(ctx context.Context, requestID string) (err error) {
	return s.revokeSessionByRequestID(ctx, storage.OAuth2SessionTypeRefreshToken, requestID)
}

func (s *OpenIDConnectStore) RevokeRefreshTokenMaybeGracePeriod(ctx context.Context, requestID string, signature string) (err error) {
	return s.RevokeRefreshToken(ctx, requestID)
}

func (s *OpenIDConnectStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {
	return s.loadSessionBySignature(ctx, storage.OAuth2SessionTypeRefreshToken, signature, session)
}

func (s *OpenIDConnectStore) CreatePKCERequestSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypePKCEChallenge, signature, request)
}

func (s *OpenIDConnectStore) DeletePKCERequestSession(ctx context.Context, signature string) (err error) {
	return s.revokeSessionBySignature(ctx, storage.OAuth2SessionTypeAccessToken, signature)
}

func (s *OpenIDConnectStore) GetPKCERequestSession(ctx context.Context, signature string, session fosite.Session) (requester fosite.Requester, err error) {
	return s.loadSessionBySignature(ctx, storage.OAuth2SessionTypePKCEChallenge, signature, session)
}

func (s *OpenIDConnectStore) CreateOpenIDConnectSession(ctx context.Context, authorizeCode string, request fosite.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypeOpenIDConnect, authorizeCode, request)
}

func (s *OpenIDConnectStore) DeleteOpenIDConnectSession(ctx context.Context, authorizeCode string) (err error) {
	return s.revokeSessionBySignature(ctx, storage.OAuth2SessionTypeAccessToken, authorizeCode)
}

func (s *OpenIDConnectStore) GetOpenIDConnectSession(ctx context.Context, authorizeCode string, request fosite.Requester) (r fosite.Requester, err error) {
	return s.loadSessionBySignature(ctx, storage.OAuth2SessionTypeOpenIDConnect, authorizeCode, request.GetSession())
}

func (s *OpenIDConnectStore) loadSessionBySignature(ctx context.Context, sessionType storage.OAuth2SessionType, signature string, session fosite.Session) (r fosite.Requester, err error) {
	var (
		sessionModel *model.OAuth2Session
	)

	if sessionModel, err = s.provider.LoadOAuth2Session(ctx, sessionType, signature); err != nil {
		return nil, err
	}

	return sessionModel.ToRequest(ctx, session, s)
}

func (s *OpenIDConnectStore) saveSession(ctx context.Context, sessionType storage.OAuth2SessionType, signature string, r fosite.Requester) (err error) {
	var session *model.OAuth2Session

	if session, err = model.NewOAuth2SessionFromRequest(signature, r); err != nil {
		return err
	}

	return s.provider.SaveOAuth2Session(ctx, sessionType, session)
}

func (s *OpenIDConnectStore) revokeSessionBySignature(ctx context.Context, sessionType storage.OAuth2SessionType, signature string) (err error) {
	return s.provider.RevokeOAuth2Session(ctx, sessionType, signature)
}

func (s *OpenIDConnectStore) revokeSessionByRequestID(ctx context.Context, sessionType storage.OAuth2SessionType, requestID string) (err error) {
	return s.provider.RevokeOAuth2SessionByRequestID(ctx, sessionType, requestID)
}

/*
// NewOpenIDConnectStore returns a new OpenIDConnectStore using the provided schema.OpenIDConnectConfiguration.
func NewOpenIDConnectStore(configuration *schema.OpenIDConnectConfiguration) (store *OpenIDConnectStore) {
	logger := logging.Logger()

	store = &OpenIDConnectStore{
		memory: &storage.MemoryStore{
			IDSessions:             map[string]fosite.Requester{},
			Users:                  map[string]storage.MemoryUserRelation{},
			AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
			AccessTokens:           map[string]fosite.Requester{},
			RefreshTokens:          map[string]storage.StoreRefreshToken{},
			PKCES:                  map[string]fosite.Requester{},
			AccessTokenRequestIDs:  map[string]string{},
			RefreshTokenRequestIDs: map[string]string{},
		},
	}

	store.clients = make(map[string]*Client)

	for _, client := range configuration.Clients {
		policy := authorization.PolicyToLevel(client.Policy)
		logger.Debugf("Registering client %s with policy %s (%v)", client.ID, client.Policy, policy)

		store.clients[client.ID] = NewClient(client)
	}

	return store
}

// GetClientPolicy retrieves the policy from the client with the matching provided id.
func (s OpenIDConnectStore) GetClientPolicy(id string) (level authorization.Level) {
	client, err := s.GetFullClient(id)
	if err != nil {
		return authorization.TwoFactor
	}

	return client.Policy
}

// GetFullClient returns a fosite.Client asserted as an Client matching the provided id.
func (s OpenIDConnectStore) GetFullClient(id string) (client *Client, err error) {
	client, ok := s.clients[id]
	if !ok {
		return nil, fosite.ErrNotFound
	}

	return client, nil
}

// IsValidClientID returns true if the provided id exists in the OpenIDConnectProvider.Clients map.
func (s OpenIDConnectStore) IsValidClientID(id string) (valid bool) {
	_, err := s.GetFullClient(id)

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
	return s.GetFullClient(id)
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

// RevokeAccessToken decorates fosite's storage.MemoryStore RevokeAccessToken method.
func (s *OpenIDConnectStore) RevokeAccessToken(ctx context.Context, requestID string) error {
	return s.memory.RevokeAccessToken(ctx, requestID)
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

// RevokeRefreshToken decorates fosite's storage.MemoryStore RevokeRefreshToken method.
func (s *OpenIDConnectStore) RevokeRefreshToken(ctx context.Context, requestID string) error {
	return s.memory.RevokeRefreshToken(ctx, requestID)
}

// RevokeRefreshTokenMaybeGracePeriod decorates fosite's storage.MemoryStore RevokeRefreshTokenMaybeGracePeriod method.
func (s OpenIDConnectStore) RevokeRefreshTokenMaybeGracePeriod(ctx context.Context, requestID string, signature string) error {
	return s.memory.RevokeRefreshTokenMaybeGracePeriod(ctx, requestID, signature)
}.


*/

/*
ResourceOwnerPasswordCredentialsGrantStorage
// Authenticate decorates fosite's storage.MemoryStore Authenticate method.
func (s *OpenIDConnectStore) Authenticate(ctx context.Context, name string, secret string) error {
	return s.memory.Authenticate(ctx, name, secret)
}.
*/

/*
rfc7523.RFC7523KeyStorage

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
	return s.memory.MarkJWTUsedForTime(ctx, jti, exp)
}.

*/
