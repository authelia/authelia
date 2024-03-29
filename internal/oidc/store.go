package oidc

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

// NewStore returns a Store when provided with a schema.OpenIDConnect and storage.Provider.
func NewStore(config *schema.IdentityProvidersOpenIDConnect, provider storage.Provider) (store *Store) {
	store = &Store{
		ClientStore: NewMemoryClientStore(config),
		provider:    provider,
	}

	return store
}

func NewMemoryClientStore(config *schema.IdentityProvidersOpenIDConnect) (store *MemoryClientStore) {
	logger := logging.Logger()

	store = &MemoryClientStore{
		clients: map[string]Client{},
	}

	for _, client := range config.Clients {
		policy := authorization.NewLevel(client.AuthorizationPolicy)
		logger.Debugf("Registering client %s with policy %s (%v)", client.ID, client.AuthorizationPolicy, policy)

		store.clients[client.ID] = NewClient(client, config)
	}

	return store
}

// GetRegisteredClient returns a Client matching the provided id.
func (s *MemoryClientStore) GetRegisteredClient(_ context.Context, id string) (client Client, err error) {
	client, ok := s.clients[id]
	if !ok {
		return nil, oauthelia2.ErrInvalidClient.WithDebugf("Client with id '%s' does not appear to be a registered client.", id)
	}

	return client, nil
}

// GenerateOpaqueUserID either retrieves or creates an opaque user id from a sectorID and username.
func (s *Store) GenerateOpaqueUserID(ctx context.Context, sectorID, username string) (opaqueID *model.UserOpaqueIdentifier, err error) {
	if opaqueID, err = s.provider.LoadUserOpaqueIdentifierBySignature(ctx, "openid", sectorID, username); err != nil {
		return nil, err
	} else if opaqueID == nil {
		if opaqueID, err = model.NewUserOpaqueIdentifier("openid", sectorID, username); err != nil {
			return nil, err
		}

		if err = s.provider.SaveUserOpaqueIdentifier(ctx, *opaqueID); err != nil {
			return nil, err
		}
	}

	return opaqueID, nil
}

// GetSubject returns a subject UUID for a username. If it exists, it returns the existing one, otherwise it creates and saves it.
func (s *Store) GetSubject(ctx context.Context, sectorID, username string) (subject uuid.UUID, err error) {
	var opaqueID *model.UserOpaqueIdentifier

	if opaqueID, err = s.GenerateOpaqueUserID(ctx, sectorID, username); err != nil {
		return uuid.UUID{}, err
	}

	return opaqueID.Identifier, nil
}

// IsValidClientID returns true if the provided id exists in the OpenIDConnectProvider.Clients map.
func (s *Store) IsValidClientID(ctx context.Context, id string) (valid bool) {
	_, err := s.GetRegisteredClient(ctx, id)

	return err == nil
}

// BeginTX starts a transaction.
// This implements a portion of fosite storage.Transactional interface.
func (s *Store) BeginTX(ctx context.Context) (c context.Context, err error) {
	return s.provider.BeginTX(ctx)
}

// Commit completes a transaction.
// This implements a portion of fosite storage.Transactional interface.
func (s *Store) Commit(ctx context.Context) (err error) {
	return s.provider.Commit(ctx)
}

// Rollback rolls a transaction back.
// This implements a portion of fosite storage.Transactional interface.
func (s *Store) Rollback(ctx context.Context) (err error) {
	return s.provider.Rollback(ctx)
}

// GetClient loads the client by its ID or returns an error if the client does not exist or another error occurred.
// This implements a portion of oauthelia2.ClientManager.
func (s *Store) GetClient(ctx context.Context, id string) (client oauthelia2.Client, err error) {
	return s.GetRegisteredClient(ctx, id)
}

// ClientAssertionJWTValid returns an error if the JTI is known or the DB check failed and nil if the JTI is not known.
// This implements a portion of oauthelia2.ClientManager.
func (s *Store) ClientAssertionJWTValid(ctx context.Context, jti string) (err error) {
	signature := fmt.Sprintf("%x", sha256.Sum256([]byte(jti)))

	blacklistedJTI, err := s.provider.LoadOAuth2BlacklistedJTI(ctx, signature)

	switch {
	case errors.Is(sql.ErrNoRows, err):
		return nil
	case err != nil:
		return err
	case blacklistedJTI.ExpiresAt.After(time.Now()):
		return oauthelia2.ErrJTIKnown
	default:
		return nil
	}
}

// SetClientAssertionJWT marks a JTI as known for the given expiry time. Before inserting the new JTI, it will clean
// up any existing JTIs that have expired as those tokens can not be replayed due to the expiry.
// This implements a portion of oauthelia2.ClientManager.
func (s *Store) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) (err error) {
	blacklistedJTI := model.NewOAuth2BlacklistedJTI(jti, exp)

	return s.provider.SaveOAuth2BlacklistedJTI(ctx, blacklistedJTI)
}

// CreateAuthorizeCodeSession stores the authorization request for a given authorization code.
// This implements a portion of oauth2.AuthorizeCodeStorage.
func (s *Store) CreateAuthorizeCodeSession(ctx context.Context, code string, request oauthelia2.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypeAuthorizeCode, code, request)
}

// InvalidateAuthorizeCodeSession is called when an authorize code is being used. The state of the authorization
// code should be set to invalid and consecutive requests to GetAuthorizeCodeSession should return the
// ErrInvalidatedAuthorizeCode error.
// This implements a portion of oauth2.AuthorizeCodeStorage.
func (s *Store) InvalidateAuthorizeCodeSession(ctx context.Context, code string) (err error) {
	return s.provider.DeactivateOAuth2Session(ctx, storage.OAuth2SessionTypeAuthorizeCode, code)
}

// GetAuthorizeCodeSession hydrates the session based on the given code and returns the authorization request.
// If the authorization code has been invalidated with `InvalidateAuthorizeCodeSession`, this
// method should return the ErrInvalidatedAuthorizeCode error.
// Make sure to also return the oauthelia2.Requester value when returning the oauthelia2.ErrInvalidatedAuthorizeCode error!
// This implements a portion of oauth2.AuthorizeCodeStorage.
func (s *Store) GetAuthorizeCodeSession(ctx context.Context, code string, session oauthelia2.Session) (request oauthelia2.Requester, err error) {
	return s.loadRequesterBySignature(ctx, storage.OAuth2SessionTypeAuthorizeCode, code, session)
}

// CreateAccessTokenSession stores the authorization request for a given access token.
// This implements a portion of oauth2.AccessTokenStorage.
func (s *Store) CreateAccessTokenSession(ctx context.Context, signature string, request oauthelia2.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypeAccessToken, signature, request)
}

// DeleteAccessTokenSession marks an access token session as deleted.
// This implements a portion of oauth2.AccessTokenStorage.
func (s *Store) DeleteAccessTokenSession(ctx context.Context, signature string) (err error) {
	return s.revokeSessionBySignature(ctx, storage.OAuth2SessionTypeAccessToken, signature)
}

// RevokeAccessToken revokes an access token as specified in: https://datatracker.ietf.org/doc/html/rfc7009#section-2.1
// If the token passed to the request is an access token, the server MAY revoke the respective refresh token as well.
// This implements a portion of oauth2.TokenRevocationStorage.
func (s *Store) RevokeAccessToken(ctx context.Context, requestID string) (err error) {
	return s.revokeSessionByRequestID(ctx, storage.OAuth2SessionTypeAccessToken, requestID)
}

// GetAccessTokenSession gets the authorization request for a given access token.
// This implements a portion of oauth2.AccessTokenStorage.
func (s *Store) GetAccessTokenSession(ctx context.Context, signature string, session oauthelia2.Session) (request oauthelia2.Requester, err error) {
	return s.loadRequesterBySignature(ctx, storage.OAuth2SessionTypeAccessToken, signature, session)
}

// CreateRefreshTokenSession stores the authorization request for a given refresh token.
// This implements a portion of oauth2.RefreshTokenStorage.
func (s *Store) CreateRefreshTokenSession(ctx context.Context, signature string, request oauthelia2.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypeRefreshToken, signature, request)
}

// DeleteRefreshTokenSession marks the authorization request for a given refresh token as deleted.
// This implements a portion of oauth2.RefreshTokenStorage.
func (s *Store) DeleteRefreshTokenSession(ctx context.Context, signature string) (err error) {
	return s.revokeSessionBySignature(ctx, storage.OAuth2SessionTypeRefreshToken, signature)
}

// RevokeRefreshToken revokes a refresh token as specified in: https://datatracker.ietf.org/doc/html/rfc7009#section-2.1
// If the particular token is a refresh token and the authorization server supports the revocation of access tokens,
// then the authorization server SHOULD also invalidate all access tokens based on the same authorization grant (see Implementation Note).
// This implements a portion of oauth2.TokenRevocationStorage.
func (s *Store) RevokeRefreshToken(ctx context.Context, requestID string) (err error) {
	return s.provider.DeactivateOAuth2SessionByRequestID(ctx, storage.OAuth2SessionTypeRefreshToken, requestID)
}

// RevokeRefreshTokenMaybeGracePeriod revokes an access token as specified in: https://datatracker.ietf.org/doc/html/rfc7009#section-2.1
// If the token passed to the request is an access token, the server MAY revoke the respective refresh token as well.
// This implements a portion of oauth2.TokenRevocationStorage.
func (s *Store) RevokeRefreshTokenMaybeGracePeriod(ctx context.Context, requestID string, signature string) (err error) {
	return s.RevokeRefreshToken(ctx, requestID)
}

// GetRefreshTokenSession gets the authorization request for a given refresh token.
// This implements a portion of oauth2.RefreshTokenStorage.
func (s *Store) GetRefreshTokenSession(ctx context.Context, signature string, session oauthelia2.Session) (request oauthelia2.Requester, err error) {
	return s.loadRequesterBySignature(ctx, storage.OAuth2SessionTypeRefreshToken, signature, session)
}

// CreatePKCERequestSession stores the authorization request for a given PKCE request.
// This implements a portion of pkce.PKCERequestStorage.
func (s *Store) CreatePKCERequestSession(ctx context.Context, signature string, request oauthelia2.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypePKCEChallenge, signature, request)
}

// DeletePKCERequestSession marks the authorization request for a given PKCE request as deleted.
// This implements a portion of pkce.PKCERequestStorage.
func (s *Store) DeletePKCERequestSession(ctx context.Context, signature string) (err error) {
	return s.revokeSessionBySignature(ctx, storage.OAuth2SessionTypePKCEChallenge, signature)
}

// GetPKCERequestSession gets the authorization request for a given PKCE request.
// This implements a portion of pkce.PKCERequestStorage.
func (s *Store) GetPKCERequestSession(ctx context.Context, signature string, session oauthelia2.Session) (requester oauthelia2.Requester, err error) {
	return s.loadRequesterBySignature(ctx, storage.OAuth2SessionTypePKCEChallenge, signature, session)
}

// CreateOpenIDConnectSession creates an OpenID Connect 1.0 connect session for a given authorize code.
// This is relevant for explicit OpenID Connect 1.0 flow.
// This implements a portion of openid.OpenIDConnectRequestStorage.
func (s *Store) CreateOpenIDConnectSession(ctx context.Context, authorizeCode string, request oauthelia2.Requester) (err error) {
	return s.saveSession(ctx, storage.OAuth2SessionTypeOpenIDConnect, authorizeCode, request)
}

// DeleteOpenIDConnectSession just implements the method required by fosite even though it's unused.
// This implements a portion of openid.OpenIDConnectRequestStorage.
func (s *Store) DeleteOpenIDConnectSession(ctx context.Context, authorizeCode string) (err error) {
	return s.revokeSessionBySignature(ctx, storage.OAuth2SessionTypeOpenIDConnect, authorizeCode)
}

// GetOpenIDConnectSession returns error:
// - nil if a session was found,
// - ErrNoSessionFound if no session was found
// - or an arbitrary error if an error occurred.
// This implements a portion of openid.OpenIDConnectRequestStorage.
func (s *Store) GetOpenIDConnectSession(ctx context.Context, authorizeCode string, request oauthelia2.Requester) (r oauthelia2.Requester, err error) {
	return s.loadRequesterBySignature(ctx, storage.OAuth2SessionTypeOpenIDConnect, authorizeCode, request.GetSession())
}

// CreatePARSession stores the pushed authorization request context. The requestURI is used to derive the key.
// This implements a portion of oauthelia2.PARStorage.
func (s *Store) CreatePARSession(ctx context.Context, requestURI string, request oauthelia2.AuthorizeRequester) (err error) {
	var par *model.OAuth2PARContext

	if par, err = model.NewOAuth2PARContext(requestURI, request); err != nil {
		return err
	}

	return s.provider.SaveOAuth2PARContext(ctx, *par)
}

// GetPARSession gets the push authorization request context. The caller is expected to merge the AuthorizeRequest.
// This implements a portion of oauthelia2.PARStorage.
func (s *Store) GetPARSession(ctx context.Context, requestURI string) (request oauthelia2.AuthorizeRequester, err error) {
	var par *model.OAuth2PARContext

	if par, err = s.provider.LoadOAuth2PARContext(ctx, requestURI); err != nil {
		return nil, err
	}

	return par.ToAuthorizeRequest(ctx, NewSession(), s)
}

// DeletePARSession deletes the context.
// This implements a portion of oauthelia2.PARStorage.
func (s *Store) DeletePARSession(ctx context.Context, requestURI string) (err error) {
	return s.provider.RevokeOAuth2PARContext(ctx, requestURI)
}

// IsJWTUsed implements an interface required for RFC7523.
func (s *Store) IsJWTUsed(ctx context.Context, jti string) (used bool, err error) {
	if err = s.ClientAssertionJWTValid(ctx, jti); err != nil {
		return true, err
	}

	return false, nil
}

// MarkJWTUsedForTime implements an interface required for rfc7523.RFC7523KeyStorage.
func (s *Store) MarkJWTUsedForTime(ctx context.Context, jti string, exp time.Time) (err error) {
	return s.SetClientAssertionJWT(ctx, jti, exp)
}

func (s *Store) loadRequesterBySignature(ctx context.Context, sessionType storage.OAuth2SessionType, signature string, session oauthelia2.Session) (r oauthelia2.Requester, err error) {
	var (
		sessionModel *model.OAuth2Session
	)

	sessionModel, err = s.provider.LoadOAuth2Session(ctx, sessionType, signature)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, oauthelia2.ErrNotFound
		default:
			return nil, err
		}
	}

	if r, err = sessionModel.ToRequest(ctx, session, s); err != nil {
		return nil, err
	}

	if !sessionModel.Active {
		switch sessionType {
		case storage.OAuth2SessionTypeAuthorizeCode:
			return r, oauthelia2.ErrInvalidatedAuthorizeCode
		default:
			return r, oauthelia2.ErrInactiveToken
		}
	}

	return r, nil
}

func (s *Store) saveSession(ctx context.Context, sessionType storage.OAuth2SessionType, signature string, r oauthelia2.Requester) (err error) {
	var session *model.OAuth2Session

	if session, err = model.NewOAuth2SessionFromRequest(signature, r); err != nil {
		return err
	}

	return s.provider.SaveOAuth2Session(ctx, sessionType, *session)
}

func (s *Store) revokeSessionBySignature(ctx context.Context, sessionType storage.OAuth2SessionType, signature string) (err error) {
	return s.provider.RevokeOAuth2Session(ctx, sessionType, signature)
}

func (s *Store) revokeSessionByRequestID(ctx context.Context, sessionType storage.OAuth2SessionType, requestID string) (err error) {
	if err = s.provider.RevokeOAuth2SessionByRequestID(ctx, sessionType, requestID); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return oauthelia2.ErrNotFound
		default:
			return err
		}
	}

	return nil
}
