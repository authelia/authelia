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
	SessionGet(ctx context.Context, issuer, id string) (record session.Record, err error)

	// SessionGetByPublicID should return a session with a matching public id, which records the id it is stored against
	// as the caller has no way to derive it.
	SessionGetByPublicID(ctx context.Context, issuer, pid string) (record session.Record, err error)

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

	// SessionChangeID is used to change the cookie value of a session and update the id. The data is written as part of
	// the move as the caller reseals it against the new id.
	SessionChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error)

	// SessionGarbageCollection cleans up old session.
	SessionGarbageCollection(ctx context.Context) (err error)

	// SessionGarbageCollectionFrequency returns the frequency the garbage collection should be run at. A zero value
	// indicates the provider expires sessions itself and no garbage collection is required.
	SessionGarbageCollectionFrequency(ctx context.Context) (frequency time.Duration)
}

// NewSessionRepository returns a session.Repository backed by the given Provider.
func NewSessionRepository(provider Provider) session.Repository {
	return SessionRepository{provider: provider}
}

// SessionRepository adapts a Provider to the session.Repository interface.
type SessionRepository struct {
	provider Provider
}

// Get implements the session.Repository interface.
func (s SessionRepository) Get(ctx context.Context, issuer string, id string) (record session.Record, err error) {
	return s.provider.SessionGet(ctx, issuer, id)
}

// GetByPublicID implements the session.Repository interface.
func (s SessionRepository) GetByPublicID(ctx context.Context, issuer string, pid string) (record session.Record, err error) {
	return s.provider.SessionGetByPublicID(ctx, issuer, pid)
}

// GetIDsByUsername implements the session.Repository interface.
func (s SessionRepository) GetIDsByUsername(ctx context.Context, issuer string, username string) (ids []string, err error) {
	return s.provider.SessionGetIDsByUsername(ctx, issuer, username)
}

// Save implements the session.Repository interface.
func (s SessionRepository) Save(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, data []byte) (err error) {
	return s.provider.SessionSave(ctx, issuer, id, pid, username, expiration, data)
}

// SaveData implements the session.Repository interface.
func (s SessionRepository) SaveData(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, data []byte) (err error) {
	return s.provider.SessionSaveData(ctx, issuer, id, pid, username, expiration, data)
}

// Delete implements the session.Repository interface.
func (s SessionRepository) Delete(ctx context.Context, issuer string, id string, pid string, username string) (err error) {
	return s.provider.SessionDelete(ctx, issuer, id, pid, username)
}

// ChangeID implements the session.Repository interface.
func (s SessionRepository) ChangeID(ctx context.Context, issuer string, oldID string, id string, pid string, username string, expiration time.Duration, data []byte) (err error) {
	return s.provider.SessionChangeID(ctx, issuer, oldID, id, pid, username, expiration, data)
}

// GarbageCollection implements the session.Repository interface.
func (s SessionRepository) GarbageCollection(ctx context.Context) (err error) {
	return s.provider.SessionGarbageCollection(ctx)
}

// GarbageCollectionFrequency implements the session.Repository interface.
func (s SessionRepository) GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return s.provider.SessionGarbageCollectionFrequency(ctx)
}

var (
	_ session.Repository = (*SessionRepository)(nil)
)
