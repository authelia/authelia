// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"net/http"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// Identity of the user who is being verified.
type Identity struct {
	Username    string
	Email       string
	DisplayName string
}

// Context is the request context a Strategy operates against, providing access to the cookies of the request and
// response.
type Context interface {
	context.Context

	// GetCookie returns the value of the request cookie with the given name, or an empty string when the request carries
	// no such cookie.
	GetCookie(name string) string

	// SetCookie sets the given cookie on the response. The cookie is also applied to the request so that reads within
	// the remainder of this request observe it rather than the value the user agent sent.
	SetCookie(cookie *http.Cookie)

	// ClearCookie expires the given cookie with the user agent. The cookie must already be expired, and must carry the
	// same name, domain, and path as the cookie being cleared, as user agents key cookies on all three and would
	// otherwise retain the original.
	ClearCookie(cookie *http.Cookie)
}

// CachingContext is an optional interface a Context may implement to retain the session it loaded for the duration of
// the request. A Strategy which is given one loads the session from the Repository once rather than once per consumer,
// as a single request is commonly read by a middleware and then again by the handler behind it. It is keyed by cookie
// domain because a request may be handled on behalf of a target which belongs to a different domain than the request
// itself. A Context which doesn't implement it simply loads the session every time.
type CachingContext interface {
	Context

	// CachedSession returns the session retained for the given cookie domain, if there is one.
	CachedSession(domain string) (session *UserSession, ok bool)

	// CacheSession retains the session for the given cookie domain. A nil session discards what is retained.
	CacheSession(domain string, session *UserSession)
}

// Provider resolves the Strategy responsible for a given session cookie domain.
type Provider interface {
	// GetStrategy returns the Strategy for the given cookie domain, returning an error when no session cookie is
	// configured for it.
	GetStrategy(domain string) (strategy Strategy, err error)
}

// The Strategy is the main interface consumers of this API should be using. This allows inspecting the session config,
// creating new sessions, getting current sessions, saving sessions, and destroying sessions.
type Strategy interface {
	// GetConfig returns the session cookie configuration this Strategy was constructed with.
	GetConfig() (config schema.SessionCookie)

	// New returns a session for the given username bound to the cookie domain of this Strategy. Naming the user at
	// construction is the only supported way to give a session a username, as changing it on an existing session would
	// hand the session of one user to another.
	New(username string) (userSession UserSession)

	// NewDefault returns an anonymous session bound to the cookie domain of this Strategy.
	NewDefault() (userSession UserSession)

	// Get returns the session of the request, which is an anonymous session when the request carries no session cookie
	// or the cookie doesn't resolve to a session which is still held by the Repository. An error is returned when the
	// Repository fails, or when the session can't be opened or is bound to another cookie domain.
	Get(ctx Context) (session *UserSession, err error)

	// Save persists the session and delivers the session cookie to the user agent, generating the identifiers the
	// session needs when it has none. The session is a pointer as the generated public identifier is written back to it,
	// which the caller must retain to avoid orphaning it on a later save.
	Save(ctx Context, userSession *UserSession) (err error)

	// Regenerate issues a new session identifier for the session of the request, moving the persisted session to it and
	// preserving the session data. It mitigates session fixation and is expected at every authentication level change.
	Regenerate(ctx Context) (err error)

	// Destroy removes the persisted session of the request and instructs the user agent to discard the session cookie.
	Destroy(ctx Context) (err error)
}

// The Record is a session as a Repository holds it. It carries the signature the session is stored against alongside
// the sealed session itself, as the Codec requires the former to open the latter. A Repository returns a nil Record for
// a session which doesn't exist or which has expired.
type Record interface {
	// GetSessionSignature returns the identifier the session is stored against, which is bound into the additional
	// authenticated data the session is sealed with.
	GetSessionSignature() (signature string)

	// GetSessionData returns the sealed session.
	GetSessionData() (data []byte)
}

// NewRecord returns a Record for the given signature and sealed session data.
func NewRecord(signature string, data []byte) (record Record) {
	return &defaultRecord{signature: signature, data: data}
}

// The Repository is the backend storage implementation which holds the sessions for the Strategy.
type Repository interface {
	// Get returns the session record stored against the identifier for the issuer, returning a nil record without an
	// error when there is no such session or it has expired.
	Get(ctx context.Context, issuer string, id string) (record Record, err error)

	// GetByPublicID returns the session record stored against the public identifier for the issuer. The record carries
	// the identifier it is stored against, which the caller requires to open it and can't derive from the public
	// identifier.
	GetByPublicID(ctx context.Context, issuer string, pid string) (record Record, err error)

	// GetIDsByUsername returns the identifiers of every session of the given username for the issuer, which allows the
	// sessions of a single user to be acted on without scanning every session.
	GetIDsByUsername(ctx context.Context, issuer string, username string) (ids []string, err error)

	// Save creates or replaces the session record for the identifier, including the public identifier and username
	// associated with it.
	Save(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, data []byte) (err error)

	// SaveData updates only the sealed session data and expiration of the existing session record for the identifier,
	// and isn't guaranteed to create the record or update the public identifier or username associated with it.
	SaveData(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, data []byte) (err error)

	// Delete removes the session record for the identifier along with the public identifier and username lookups which
	// refer to it.
	Delete(ctx context.Context, issuer string, id string, pid string, username string) (err error)

	// ChangeID moves the session record stored against oldID to id, writing the sealed data as part of the move as the
	// caller reseals it against the new identifier, and repointing the lookups which refer to it.
	ChangeID(ctx context.Context, issuer string, oldID string, id string, pid string, username string, expiration time.Duration, data []byte) (err error)

	// GarbageCollection removes the sessions which have expired, for a backend which doesn't expire them itself.
	GarbageCollection(ctx context.Context) (err error)

	// GarbageCollectionFrequency returns how often GarbageCollection should be run. A zero value indicates the backend
	// expires sessions itself and no collection is required.
	GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration)

	model.StartupCheck
}

// The Codec handles obfuscation and privacy functionality such as generating private and public session identifiers,
// sealing session data and opening it, signing and verifying values, etc.
type Codec interface {
	// GeneratePublicID returns a new random public identifier, which is the value shared with consumers that reference
	// a session without being able to derive the identifier it's stored against.
	GeneratePublicID() (id string, err error)

	// GenerateSessionID returns a new random session identifier, which is the raw value encoded into the session cookie
	// and is never stored by the Repository itself.
	GenerateSessionID() (id []byte, err error)

	// Verify returns true when the signature is the signature of the given data, comparing them in constant time. A
	// signature which isn't valid hexadecimal is never verified.
	Verify(data []byte, signature string) bool

	// Sign returns the hexadecimal encoded HMAC signature of the given data. The session cookie value is signed this way
	// to derive the identifier a session is stored against, so the Repository never holds the cookie value itself.
	Sign(data []byte) string

	// Seal marshals and encrypts the session, binding the result to the cookie domain and the identifier it's stored
	// against so that it can't be opened after being moved to another domain or another record.
	Seal(domain, id string, session UserSession) (data []byte, err error)

	// Open decrypts and unmarshals the data of the record into the session, requiring the cookie domain and the
	// identifier the record reports to match those it was sealed with. A nil record, or one with no data, leaves the
	// session untouched and returns no error.
	Open(domain string, record Record, session *UserSession) (err error)
}

type defaultRecord struct {
	signature string
	data      []byte
}

func (r *defaultRecord) GetSessionSignature() (signature string) {
	return r.signature
}

func (r *defaultRecord) GetSessionData() (data []byte) {
	return r.data
}
