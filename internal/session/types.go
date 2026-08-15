package session

import (
	"context"
	"net/http"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Identity of the user who is being verified.
type Identity struct {
	Username    string
	Email       string
	DisplayName string
}

type Context interface {
	context.Context

	GetCookie(name string) string
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

type Provider interface {
	GetStrategy(domain string) (strategy Strategy, err error)
}

// The Strategy is the main interface consumers of this API should be using. This allows inspecting the session config,
// creating new sessions, getting current sessions, saving sessions, and destroying sessions.
type Strategy interface {
	GetConfig() (config schema.SessionCookie)
	New(username string) (userSession UserSession)
	NewDefault() (userSession UserSession)
	Get(ctx Context) (session *UserSession, err error)
	Save(ctx Context, userSession *UserSession) (err error)
	Regenerate(ctx Context) (err error)
	Destroy(ctx Context) (err error)
}

// The Repository is the backend storage implementation which holds the sessions for the Strategy.
type Repository interface {
	Get(ctx context.Context, issuer string, id string) (data []byte, err error)
	GetByPublicID(ctx context.Context, issuer string, pid string) (data []byte, err error)
	GetIDsByUsername(ctx context.Context, issuer string, username string) (ids []string, err error)
	Save(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, data []byte) (err error)
	SaveData(ctx context.Context, issuer string, id string, pid string, username string, expiration time.Duration, data []byte) (err error)
	Delete(ctx context.Context, issuer string, id string, pid string, username string) (err error)
	ChangeID(ctx context.Context, issuer string, oldID string, id string, pid string, username string, expiration time.Duration) (err error)
	GarbageCollection(ctx context.Context) (err error)
	GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration)
}

// The Codec handles obfuscation and privacy functionality such as generating private and public session identifiers,
// sealing session data and opening it, signing and verifying values, etc.
type Codec interface {
	GeneratePublicID() (id string, err error)
	GenerateSessionID() (id string, err error)
	Verify(data []byte, signature string) bool
	Sign(data []byte) string
	Seal(domain string, session UserSession) (data []byte, err error)
	Open(domain string, session *UserSession, src []byte) (err error)
}
