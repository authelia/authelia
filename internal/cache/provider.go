package cache

import (
	"context"
	"time"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

// The Provider is implemented assuming all required encryption and washing is done prior to the values making it to
// the functions.
type Provider interface {
	model.StartupCheck

	// SessionGet should return a session with a matching id (which is a representation of the cookie value hashed using
	// HMAC-SHA256 and encoded into hexadecimal) and issuer (which is a domain).
	SessionGet(ctx context.Context, issuer, id string) (data []byte, err error)

	// SessionGetByPublicID should return a session with a matching public id.
	SessionGetByPublicID(ctx context.Context, issuer, pid string) (data []byte, err error)

	// SessionGetIDsByUsername should return all session ids for a given username and issuer (which is a domain).
	SessionGetIDsByUsername(ctx context.Context, issuer, username string) (ids []string, err error)

	// SessionSave should save a session to the cache, and ensure the id, public id, and username can all be used to
	// find a session. This is traditionally used for new session. If te data for a session has just updated, use
	// SessionSaveData instead.
	SessionSave(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error)

	// SessionSaveData updates the session data in the cache, refreshing the expiry of the session and every lookup
	// which refers to it.
	SessionSaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error)

	// SessionDelete should delete a session from the cache, as well as the related lookup information.
	SessionDelete(ctx context.Context, issuer, id, pid, username string) (err error)

	// SessionChangeID is used to change the cookie value of a session and update the id.
	SessionChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration) (err error)

	// SessionGarbageCollection cleans up old session.
	SessionGarbageCollection(ctx context.Context) (err error)

	// SessionGarbageCollectionFrequency returns the frequency the garbage collection should be run at. A zero value
	// indicates the provider expires sessions itself and no garbage collection is required.
	SessionGarbageCollectionFrequency(ctx context.Context) (frequency time.Duration)
}

func NewSessionRepository(provider Provider) session.Repository {
	return SessionRepository{provider: provider}
}

type SessionRepository struct {
	provider Provider
}

func (s SessionRepository) Get(ctx context.Context, issuer string, id string) (data []byte, err error) {
	return s.provider.SessionGet(ctx, issuer, id)
}

func (s SessionRepository) GetByPublicID(ctx context.Context, issuer string, pid string) (data []byte, err error) {
	return s.provider.SessionGetByPublicID(ctx, issuer, pid)
}

func (s SessionRepository) GetIDsByUsername(ctx context.Context, issuer string, username string) (ids []string, err error) {
	return s.provider.SessionGetIDsByUsername(ctx, issuer, username)
}

func (s SessionRepository) Save(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, data []byte) (err error) {
	return s.provider.SessionSave(ctx, issuer, id, pid, username, expiration, data)
}

func (s SessionRepository) SaveData(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, data []byte) (err error) {
	return s.provider.SessionSaveData(ctx, issuer, id, pid, username, expiration, data)
}

func (s SessionRepository) Delete(ctx context.Context, issuer string, id string, pid string, username string) (err error) {
	return s.provider.SessionDelete(ctx, issuer, id, pid, username)
}

func (s SessionRepository) ChangeID(ctx context.Context, issuer string, oldID string, id string, pid string, username string, expiration time.Duration) (err error) {
	return s.provider.SessionChangeID(ctx, issuer, oldID, id, pid, username, expiration)
}

func (s SessionRepository) GarbageCollection(ctx context.Context) (err error) {
	return s.provider.SessionGarbageCollection(ctx)
}

func (s SessionRepository) GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return s.provider.SessionGarbageCollectionFrequency(ctx)
}

var (
	_ session.Repository = (*SessionRepository)(nil)
)
