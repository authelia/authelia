package cache

import (
	"context"
	"time"

	"github.com/authelia/authelia/v4/internal/model"
)

// The Provider is implemented assuming all required encryption and washing is done prior to the values making it to
// the functions.
type Provider interface {
	model.StartupCheck

	// SessionGet should return a session with a matching id (which is a representation of the cookie value hashed using
	// HMAC-SHA256 and encoded into hexadecimal) and issuer (which is a domain).
	SessionGet(ctx context.Context, id, issuer string) (data []byte, err error)

	// SessionGetByPublicID should return a session with a matching public id (which is a UUIDv7 string).
	SessionGetByPublicID(ctx context.Context, pubid, issuer string) (data []byte, err error)

	// SessionGetIDsByUsername should return all session ids for a given username and issuer (which is a domain).
	SessionGetIDsByUsername(ctx context.Context, username, issuer string) (ids []string, err error)

	// SessionSave should save a session to the cache, and ensure the id, public id, and username can all be used to
	// find a session. This is traditionally used for new session. If te data for a session has just updated, use
	// SessionSaveData instead.
	SessionSave(ctx context.Context, id, pubid, username, issuer string, expiration time.Duration, data []byte) (err error)

	// SessionSetUsername is used when SaveSession has been called without a username to link the session to a
	// specific user.
	SessionSetUsername(ctx context.Context, id, username, issuer string) (err error)

	// SessionSaveData updates the session data in the cache.
	SessionSaveData(ctx context.Context, id, issuer string, expiration time.Duration, data []byte) (err error)

	// SessionDelete should delete a session from the cache, as well as the related lookup information.
	SessionDelete(ctx context.Context, id, pubid, username, issuer string) (err error)

	// SessionChangeID is used to change the cookie value of a session and update the id.
	SessionChangeID(ctx context.Context, oldID, id, pubid, username, issuer string, expiration time.Duration) (err error)

	// SessionGarbageCollection cleans up old session.
	SessionGarbageCollection(ctx context.Context) (err error)

	// SessionGarbageCollectionRequired returns true if the garbage collection should be run.
	SessionGarbageCollectionRequired(ctx context.Context) (required bool)

	//GetCachedValue(ctx context.Context, category, lookup string) (data []byte, err error)
	//SetCacheValue(ctx context.Context, category, lookup string, expiration time.Duration, data []byte) (err error)
}
