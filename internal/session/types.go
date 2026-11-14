// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

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

	// IsAuthzEndpointRequest returns true when the request is being handled by an authorization endpoint, which are
	// the only endpoints the authorization cookie is honored on and the only endpoints the session cookie is not.
	IsAuthzEndpointRequest() bool
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

// LoggingContext is an optional interface a Context may implement to receive the failures a Strategy recovers from
// rather than returning. A Context which doesn't implement it has them logged by the global logger.
type LoggingContext interface {
	Context

	// GetLogger returns the logger for the request.
	GetLogger() (logger *logrus.Entry)
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
	// Repository fails, or when the session can't be opened or is bound to another cookie domain. An authenticated
	// session which the authz cookie of the request isn't linked to is issued a new authz cookie. The request is recorded
	// as activity on the session, which retains it for the inactivity timeout from now when one applies.
	Get(ctx Context) (session *UserSession, err error)

	// Save persists the session and delivers the session cookie to the user agent, generating the identifiers the
	// session needs when it has none. The session is a pointer as the generated public identifier is written back to it,
	// which the caller must retain to avoid orphaning it on a later save.
	Save(ctx Context, userSession *UserSession) (err error)

	// Regenerate issues a new session identifier for the session of the request, moving the persisted session to it and
	// preserving the session data. It mitigates session fixation and is expected at every authentication level change.
	Regenerate(ctx Context) (err error)

	// Destroy removes the persisted session of the request and instructs the user agent to discard the session cookie,
	// the authz cookie, and the CSRF cookie. A request which carries no session cookie has the session its authz cookie
	// is linked to removed instead.
	Destroy(ctx Context) (err error)

	// CSRFToken returns the CSRF token of the established session of the request. It returns ErrCSRFTokenNoSession when
	// the request has no established session, and ErrCSRFTokenInvalid when the session has no CSRF secret.
	CSRFToken(ctx Context) (token string, err error)

	// VerifyCSRFToken returns nil if the token is the CSRF token of the established session of the request. It returns
	// ErrCSRFTokenNoSession when the request has no established session, and ErrCSRFTokenInvalid when the token doesn't
	// match.
	VerifyCSRFToken(ctx Context, token string) (err error)

	// RegenerateCSRFToken replaces the CSRF secret of the established session of the request without changing the
	// session identifier, and delivers the new token to the user agent.
	RegenerateCSRFToken(ctx Context) (err error)

	// SetCSRFCookie delivers the CSRF token of the established session of the request to the user agent, generating a
	// CSRF secret for an established session which has none.
	SetCSRFCookie(ctx Context) (err error)

	// Authz returns the restricted view of this Strategy which the authorization endpoints use. The view resolves the
	// session from the authz cookie alone and can only update the activity of the session or destroy it.
	Authz() (strategy AuthzStrategy)
}

// The AuthzStrategy is the restricted view of a Strategy for the authorization endpoints. It resolves the session from
// the authz cookie, never from the session cookie, and has no means to create a session, to change who a session
// belongs to, or to change how it was authenticated.
type AuthzStrategy interface {
	// GetConfig returns the session cookie configuration of the Strategy this view belongs to.
	GetConfig() (config schema.SessionCookie)

	// NewDefault returns an anonymous session bound to the cookie domain of this view, which can't be persisted.
	NewDefault() (userSession UserSession)

	// Get returns the session the authz cookie of the request is linked to, which is an anonymous session when the
	// request carries no authz cookie or it doesn't resolve to a session which records it, and records the request as
	// activity on the session. An error is returned only when the Repository fails.
	Get(ctx Context) (session *UserSession, err error)

	// SaveActivity persists the refresh time of the given session onto the stored session, discarding every other
	// difference. It does nothing when the request has no stored session.
	SaveActivity(ctx Context, userSession *UserSession) (err error)

	// Destroy removes the persisted session of the request and instructs the user agent to discard both cookies.
	Destroy(ctx Context) (err error)
}

// AuthzManager is the interface that wraps the methods of an AuthzStrategy bound to a specific request.
type AuthzManager interface {
	// NewDefaultUserSession returns an anonymous session bound to the cookie domain of the underlying AuthzStrategy.
	NewDefaultUserSession() (userSession UserSession)

	// GetSession returns the session the authz cookie of the bound request is linked to.
	GetSession() (userSession UserSession, err error)

	// SaveSessionActivity persists the refresh time of the session of the bound request.
	SaveSessionActivity(userSession *UserSession) (err error)

	// DestroySession removes the persisted session of the bound request and instructs the user agent to discard both
	// cookies.
	DestroySession() (err error)

	// GetSessionConfig returns the session cookie configuration of the underlying AuthzStrategy.
	GetSessionConfig() (config schema.SessionCookie)
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

	// GetSessionTTL returns how long the record had left until it expired when it was retrieved. A zero value indicates
	// the record doesn't expire or the Repository doesn't report it.
	GetSessionTTL() (ttl time.Duration)
}

// NewRecord returns a Record for the given signature and sealed session data which doesn't report when it expires.
func NewRecord(signature string, data []byte) (record Record) {
	return &defaultRecord{signature: signature, data: data}
}

// NewRecordWithTTL returns a Record for the given signature and sealed session data which had the given time left until
// it expired when it was retrieved.
func NewRecordWithTTL(signature string, data []byte, ttl time.Duration) (record Record) {
	return &defaultRecord{signature: signature, data: data, ttl: ttl}
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
	// associated with it. It returns ErrSessionSuperseded without writing anything when the session was moved to another
	// identifier or destroyed.
	Save(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, data []byte) (err error)

	// SaveData updates only the sealed session data and expiration of the existing session record for the identifier,
	// and isn't guaranteed to create the record or update the public identifier or username associated with it. When
	// previous isn't nil the record is only updated if its sealed data is still exactly previous, and otherwise nothing
	// is written and ErrSessionSuperseded is returned, which includes when the record was destroyed or moved. This allows
	// a caller which read the record to write it back without discarding a write made to it in the meantime.
	SaveData(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, previous, data []byte) (err error)

	// Touch extends the expiration of the existing session record for the identifier, and every lookup which refers to
	// it, to the given expiration from now. It never shortens the expiration, never changes the sealed data, and does
	// nothing when there is no such record or it was destroyed or moved.
	Touch(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration) (err error)

	// Delete destroys the session record for the identifier and removes the public identifier and username lookups which
	// refer to it. The destroyed session is retained until it would have expired so that a later Save of it is refused.
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

// The Backend is implemented by the cache and storage providers which hold sessions, and is adapted to the Repository
// interface by NewRepository.
type Backend interface {
	model.StartupCheck

	// SessionGet should return a session with a matching id (which is a representation of the cookie value hashed using
	// HMAC-SHA256 and encoded into hexadecimal) and issuer (which is a domain).
	SessionGet(ctx context.Context, issuer, id string) (record Record, err error)

	// SessionGetByPublicID should return a session with a matching public id, which records the id it is stored against
	// as the caller has no way to derive it.
	SessionGetByPublicID(ctx context.Context, issuer, pid string) (record Record, err error)

	// SessionGetIDsByUsername should return all session ids for a given username and issuer (which is a domain).
	SessionGetIDsByUsername(ctx context.Context, issuer, username string) (ids []string, err error)

	// SessionSave should save a session to the cache, and ensure the id, public id, and username can all be used to
	// find a session. This is traditionally used for new sessions. If the data for a session has just been updated, use
	// SessionSaveData instead.
	SessionSave(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error)

	// SessionSaveData updates the session data in the cache, refreshing the expiry of the session and every lookup
	// which refers to it. When previous isn't nil the session is only updated if its data is still exactly previous, and
	// otherwise ErrSessionSuperseded is returned.
	SessionSaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, previous, data []byte) (err error)

	// SessionTouch should extend the expiry of an existing session and every lookup which refers to it without changing
	// its data. It should never shorten the expiry, and should do nothing when the session doesn't exist or was destroyed
	// or moved.
	SessionTouch(ctx context.Context, issuer, id, pid, username string, expiration time.Duration) (err error)

	// SessionDelete should delete a session from the cache, as well as the related lookup information.
	SessionDelete(ctx context.Context, issuer, id, pid, username string) (err error)

	// SessionChangeID is used to change the cookie value of a session and update the id. The data is written as part of
	// the move as the caller reseals it against the new id.
	SessionChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error)

	// SessionGarbageCollection cleans up the sessions which have expired.
	SessionGarbageCollection(ctx context.Context) (err error)

	// SessionGarbageCollectionFrequency returns the frequency the garbage collection should be run at. A zero value
	// indicates the provider expires sessions itself and no garbage collection is required.
	SessionGarbageCollectionFrequency(ctx context.Context) (frequency time.Duration)
}

// NewRepository returns a Repository backed by the given Backend.
func NewRepository(backend Backend) Repository {
	return BackendRepository{Backend: backend}
}

// BackendRepository adapts a Backend to the Repository interface.
type BackendRepository struct {
	Backend Backend
}

// StartupCheck implements the Repository interface.
func (r BackendRepository) StartupCheck() (err error) {
	return r.Backend.StartupCheck()
}

// Get implements the Repository interface.
func (r BackendRepository) Get(ctx context.Context, issuer, id string) (record Record, err error) {
	return r.Backend.SessionGet(ctx, issuer, id)
}

// GetByPublicID implements the Repository interface.
func (r BackendRepository) GetByPublicID(ctx context.Context, issuer, pid string) (record Record, err error) {
	return r.Backend.SessionGetByPublicID(ctx, issuer, pid)
}

// GetIDsByUsername implements the Repository interface.
func (r BackendRepository) GetIDsByUsername(ctx context.Context, issuer, username string) (ids []string, err error) {
	return r.Backend.SessionGetIDsByUsername(ctx, issuer, username)
}

// Save implements the Repository interface.
func (r BackendRepository) Save(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	return r.Backend.SessionSave(ctx, issuer, id, pid, username, expiration, data)
}

// SaveData implements the Repository interface.
func (r BackendRepository) SaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, previous, data []byte) (err error) {
	return r.Backend.SessionSaveData(ctx, issuer, id, pid, username, expiration, previous, data)
}

// Touch implements the Repository interface.
func (r BackendRepository) Touch(ctx context.Context, issuer, id, pid, username string, expiration time.Duration) (err error) {
	return r.Backend.SessionTouch(ctx, issuer, id, pid, username, expiration)
}

// Delete implements the Repository interface.
func (r BackendRepository) Delete(ctx context.Context, issuer, id, pid, username string) (err error) {
	return r.Backend.SessionDelete(ctx, issuer, id, pid, username)
}

// ChangeID implements the Repository interface.
func (r BackendRepository) ChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	return r.Backend.SessionChangeID(ctx, issuer, oldID, id, pid, username, expiration, data)
}

// GarbageCollection implements the Repository interface.
func (r BackendRepository) GarbageCollection(ctx context.Context) (err error) {
	return r.Backend.SessionGarbageCollection(ctx)
}

// GarbageCollectionFrequency implements the Repository interface.
func (r BackendRepository) GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return r.Backend.SessionGarbageCollectionFrequency(ctx)
}

var (
	_ Repository = BackendRepository{}
)

// The Codec handles obfuscation and privacy functionality such as generating private and public session identifiers,
// sealing session data and opening it, signing and verifying values, etc.
type Codec interface {
	// GeneratePublicID returns a new random public identifier, which is the value shared with consumers that reference
	// a session without being able to derive the identifier it's stored against.
	GeneratePublicID() (id string, err error)

	// GenerateSessionID returns a new random session identifier, which is the raw value encoded into the session cookie
	// and is never stored by the Repository itself.
	GenerateSessionID() (id []byte, err error)

	// GenerateCSRFSecret returns a new random CSRF secret, which is stored in the session and signed to derive the CSRF
	// token delivered to the user agent.
	GenerateCSRFSecret() (secret []byte, err error)

	// Verify returns true when the signature is the signature of the given data, comparing them in constant time. A
	// signature which isn't valid hexadecimal is never verified.
	Verify(data []byte, signature string) bool

	// Sign returns the hexadecimal encoded HMAC signature of the given data using the session key. The session cookie
	// value is signed this way to derive the identifier a session is stored against, so the Repository never holds the
	// cookie value itself.
	Sign(data []byte) string

	// SignAuthz returns the hexadecimal encoded HMAC signature of the given data using the session authz key, which the
	// authz cookie value is derived from and signed with to derive the identifier its index is stored against.
	SignAuthz(data []byte) string

	// VerifyCSRF returns true when the signature is the CSRF signature of the given data, comparing them in constant
	// time. It uses the session CSRF key, so a signature produced for any other purpose is never verified.
	VerifyCSRF(data []byte, signature string) bool

	// SignCSRF returns the hexadecimal encoded HMAC signature of the given data using the session CSRF key, which is
	// how the CSRF token is derived from the CSRF secret of a session.
	SignCSRF(data []byte) string

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
	ttl       time.Duration
}

func (r *defaultRecord) GetSessionSignature() (signature string) {
	return r.signature
}

func (r *defaultRecord) GetSessionData() (data []byte) {
	return r.data
}

func (r *defaultRecord) GetSessionTTL() (ttl time.Duration) {
	return r.ttl
}
