// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewStrategy returns a new Strategy for the given session cookie configuration. The issuer is the signature of the
// Authelia URL, falling back to the cookie domain, which scopes every record this strategy stores. The session cookie
// is always strict, and the configured same site value is applied to the authz cookie.
func NewStrategy(config schema.SessionCookie, clock clock.Provider, codec Codec, storage Repository) (provider Strategy) {
	sameSite := newSameSite(config.SameSite)

	var scope string

	if config.AutheliaURL == nil {
		scope = strings.ToLower(config.Domain)
	} else {
		scope = strings.ToLower(config.AutheliaURL.String())
	}

	return &DefaultStrategy{
		config:        config,
		issuer:        codec.Sign([]byte(scope)),
		domain:        newDomain(config.Domain, http.SameSiteStrictMode),
		authzName:     config.Name + cookieNameSuffixAuthz,
		authzIssuer:   codec.SignAuthz([]byte(issuerPrefixAuthz + scope)),
		authzDomain:   newDomain(config.Domain, sameSite),
		authzSameSite: sameSite,
		clock:         clock,
		codec:         codec,
		repository:    storage,
	}
}

// DefaultStrategy is the default Strategy which stores the session identifier in a cookie and the session data in a
// Repository, scoped to a single configured session cookie domain.
type DefaultStrategy struct {
	config schema.SessionCookie

	issuer string
	domain string

	authzName     string
	authzIssuer   string
	authzDomain   string
	authzSameSite http.SameSite

	clock clock.Provider

	codec      Codec
	repository Repository
}

// GetConfig returns the session cookie configuration of this strategy.
func (p *DefaultStrategy) GetConfig() (config schema.SessionCookie) {
	return p.config
}

// New returns a session for the given username bound to this strategies cookie domain. Giving a session a username at
// construction is the only supported way to do so, see NewUserSession.
func (p *DefaultStrategy) New(username string) (userSession UserSession) {
	userSession = NewUserSession(username)

	userSession.CookieDomain = p.config.Domain

	return userSession
}

// NewDefault returns an anonymous session bound to this strategies cookie domain.
func (p *DefaultStrategy) NewDefault() (userSession UserSession) {
	return p.New("")
}

// Get returns the session for the given request, returning an anonymous session when the request has no session
// cookie or the cookie does not resolve to a readable record. An authenticated session which the authz cookie of the
// request isn't linked to is issued a new authz cookie. A request to an authorization endpoint always receives an
// anonymous session as the session cookie is not honored there.
func (p *DefaultStrategy) Get(ctx Context) (session *UserSession, err error) {
	if ctx.IsAuthzEndpointRequest() {
		userSession := p.NewDefault()

		return &userSession, nil
	}

	_, session, err = p.get(ctx, true)

	return session, err
}

// Save persists the session. The userSession is a pointer as a session which has no public identifier or authz
// identifier has one generated for it, and the caller must observe them to avoid orphaning them on a subsequent save.
// It returns ErrNotPermittedOnAuthzEndpointRequest on a request to an authorization endpoint.
func (p *DefaultStrategy) Save(ctx Context, session *UserSession) (err error) {
	if session == nil {
		return fmt.Errorf("error occurred saving session: it is nil")
	}

	if ctx.IsAuthzEndpointRequest() {
		return fmt.Errorf("error occurred saving session: %w", ErrNotPermittedOnAuthzEndpointRequest)
	}

	switch {
	case session.CookieDomain == "":
		session.CookieDomain = p.config.Domain
	case session.CookieDomain != p.config.Domain:
		return fmt.Errorf("error occurred saving session: domain does not match cookie domain")
	}

	var id, data []byte

	if id = p.getCookieID(ctx, p.config.Name); len(id) == 0 {
		if id, err = p.codec.GenerateSessionID(); err != nil {
			return fmt.Errorf("error occurred generating session ID: %w", err)
		}
	}

	if session.PublicID == "" {
		if session.PublicID, err = p.codec.GeneratePublicID(); err != nil {
			return fmt.Errorf("error occurred generating session public ID: %w", err)
		}
	}

	sid := p.codec.Sign(id)
	authz, aid := p.getAuthzValue(sid)

	expiration := p.getExpiration(*session)
	session.Expires = p.getExpiresAt(expiration)

	if data, err = p.codec.Seal(p.config.Domain, sid, *session); err != nil {
		return fmt.Errorf("error occurred encoding session: %w", err)
	}

	expires := p.getCookieExpires(expiration)

	if err = p.repository.Save(ctx, p.issuer, sid, session.PublicID, session.Username, expiration, data); err != nil {
		if errors.Is(err, ErrSessionSuperseded) {
			return nil
		}

		return fmt.Errorf("error occurred saving session to registry: %w", err)
	}

	p.setCached(ctx, p.cacheKey(), session)
	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	ctx.SetCookie(p.newCookie(id, expires))

	if session.IsAnonymous() {
		return nil
	}

	if err = p.saveAuthzIndex(ctx, aid, sid, expiration); err == nil {
		ctx.SetCookie(p.newAuthzCookie(authz, expires))
	}

	return nil
}

// Regenerate issues a new session identifier and a new authz identifier for the current session, preserving the
// session data and public identifier, which mitigates session fixation across an authentication level change. It
// returns ErrNotPermittedOnAuthzEndpointRequest on a request to an authorization endpoint.
func (p *DefaultStrategy) Regenerate(ctx Context) (err error) {
	if ctx.IsAuthzEndpointRequest() {
		return fmt.Errorf("error occurred regenerating session: %w", ErrNotPermittedOnAuthzEndpointRequest)
	}

	var (
		oldSID    string
		id, authz []byte
		session   *UserSession
	)

	if oldSID, session, err = p.get(ctx, false); err != nil {
		return err
	}

	if id, err = p.codec.GenerateSessionID(); err != nil {
		return fmt.Errorf("error occurred generating session ID: %w", err)
	}

	sid := p.codec.Sign(id)
	expiration := p.getExpiration(*session)
	expires := p.getCookieExpires(expiration)

	if len(oldSID) != 0 {
		var data []byte

		session.Expires = p.getExpiresAt(expiration)

		if data, err = p.codec.Seal(p.config.Domain, sid, *session); err != nil {
			return fmt.Errorf("error occurred encoding session: %w", err)
		}

		if err = p.repository.ChangeID(ctx, p.issuer, oldSID, sid, session.PublicID, session.Username, expiration, data); err != nil {
			return fmt.Errorf("error occurred changing session ID: %w", err)
		}

		p.setCached(ctx, p.cacheKey(), session)
		p.setCached(ctx, p.cacheKeyAuthz(), nil)

		var aid string

		// The session has moved at this point so failing to store the authz index must not prevent the session cookie
		// being issued, and only withholds the authz cookie, which the next request the session is read on issues again.
		if authz, aid = p.getAuthzValue(sid); session.IsAnonymous() || p.saveAuthzIndex(ctx, aid, sid, expiration) != nil {
			authz = nil
		}

		_, previous := p.getAuthzValue(oldSID)

		_ = p.deleteAuthzIndex(ctx, previous)
	}

	ctx.SetCookie(p.newCookie(id, expires))

	if len(authz) != 0 {
		ctx.SetCookie(p.newAuthzCookie(authz, expires))
	}

	return nil
}

// Destroy removes the backend record for the current session and instructs the user agent to discard both the session
// cookie and the authz cookie. A request which has no session cookie has the session its authz cookie is linked to
// removed instead. It returns ErrNotPermittedOnAuthzEndpointRequest on a request to an authorization endpoint.
func (p *DefaultStrategy) Destroy(ctx Context) (err error) {
	if ctx.IsAuthzEndpointRequest() {
		return fmt.Errorf("error occurred destroying session: %w", ErrNotPermittedOnAuthzEndpointRequest)
	}

	id, userSession, _ := p.get(ctx, false)

	if len(id) == 0 {
		err = p.destroyByAuthz(ctx)
	} else {
		err = p.destroyByID(ctx, id, userSession)
	}

	if err != nil {
		return err
	}

	p.setCached(ctx, p.cacheKey(), nil)
	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	ctx.ClearCookie(p.newDeletionCookie())
	ctx.ClearCookie(p.newAuthzDeletionCookie())

	return nil
}

func (p *DefaultStrategy) destroyByID(ctx Context, id string, userSession *UserSession) (err error) {
	var pid, username string

	if userSession != nil {
		pid = userSession.PublicID
		username = userSession.Username
	}

	if err = p.repository.Delete(ctx, p.issuer, id, pid, username); err != nil {
		return fmt.Errorf("error occurred deleting session from backend: %w", err)
	}

	_, aid := p.getAuthzValue(id)

	_ = p.deleteAuthzIndex(ctx, aid)

	return nil
}

func (p *DefaultStrategy) destroyByAuthz(ctx Context) (err error) {
	sid, aid, userSession, _ := (&DefaultAuthzStrategy{base: p}).resolve(ctx)

	if userSession != nil {
		if err = p.repository.Delete(ctx, p.issuer, sid, userSession.PublicID, userSession.Username); err != nil {
			return fmt.Errorf("error occurred deleting session from backend: %w", err)
		}
	}

	if aid != "" {
		_ = p.deleteAuthzIndex(ctx, aid)
	}

	return nil
}

func (p *DefaultStrategy) get(ctx Context, heal bool) (id string, session *UserSession, err error) {
	cookie := p.getCookieID(ctx, p.config.Name)

	userSession := p.NewDefault()

	if len(cookie) == 0 {
		return "", &userSession, nil
	}

	id = p.codec.Sign(cookie)

	if cached, ok := p.getCached(ctx, p.cacheKey()); ok {
		return id, cached, nil
	}

	var record Record

	if record, err = p.repository.Get(ctx, p.issuer, id); err != nil {
		return id, nil, fmt.Errorf("%w: %w", ErrRepositoryGet, err)
	}

	if record == nil || len(record.GetSessionData()) == 0 {
		p.setCached(ctx, p.cacheKey(), &userSession)

		return id, &userSession, nil
	}

	if record.GetSessionSignature() != id {
		return id, nil, fmt.Errorf("error occurred getting session: signature does not match the session identifier")
	}

	session = &UserSession{}

	if err = p.codec.Open(p.config.Domain, record, session); err != nil {
		return id, nil, fmt.Errorf("error occurred decoding session: %w", err)
	}

	if session.CookieDomain != p.config.Domain {
		return id, nil, fmt.Errorf("error occurred getting session: domain does not match cookie domain")
	}

	if heal && !session.IsAnonymous() {
		p.healAuthz(ctx, id, session)
	}

	p.setCached(ctx, p.cacheKey(), session)

	return id, session, nil
}

func (p *DefaultStrategy) healAuthz(ctx Context, sid string, session *UserSession) {
	value, aid := p.getAuthzValue(sid)

	if current := p.getCookieID(ctx, p.authzName); len(current) != 0 && subtle.ConstantTimeCompare(current, value) == 1 {
		return
	}

	remaining, ok := p.getRemaining(*session)
	if !ok {
		return
	}

	if err := p.saveAuthzIndex(ctx, aid, sid, remaining); err != nil {
		return
	}

	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	ctx.SetCookie(p.newAuthzCookie(value, p.getCookieExpires(p.getExpiration(*session))))
}

func (p *DefaultStrategy) cacheKey() string {
	return p.config.Domain
}

func (p *DefaultStrategy) cacheKeyAuthz() string {
	return cacheKeyPrefixAuthz + p.config.Domain
}

func (p *DefaultStrategy) getCached(ctx Context, key string) (session *UserSession, ok bool) {
	caching, ok := ctx.(CachingContext)
	if !ok {
		return nil, false
	}

	cached, ok := caching.CachedSession(key)
	if !ok || cached == nil {
		return nil, false
	}

	value := cached.deepCopy()

	return &value, true
}

func (p *DefaultStrategy) setCached(ctx Context, key string, session *UserSession) {
	caching, ok := ctx.(CachingContext)
	if !ok {
		return
	}

	if session == nil {
		caching.CacheSession(key, nil)

		return
	}

	value := session.deepCopy()

	caching.CacheSession(key, &value)
}

func (p *DefaultStrategy) getCookieID(ctx Context, name string) (id []byte) {
	value := ctx.GetCookie(name)

	if len(value) == 0 {
		return nil
	}

	var err error

	if id, err = base64.RawURLEncoding.DecodeString(value); err != nil || len(id) != sessionIDLength {
		return nil
	}

	return id
}

func (p *DefaultStrategy) getAuthzValue(sid string) (value []byte, aid string) {
	value, _ = hex.DecodeString(p.codec.SignAuthz([]byte(authzValuePrefix + p.issuer + ":" + sid)))

	return value, p.codec.SignAuthz(value)
}

func (p *DefaultStrategy) saveAuthzIndex(ctx Context, aid, sid string, expiration time.Duration) (err error) {
	return p.repository.Save(ctx, p.authzIssuer, aid, getAuthzIndexPublicID(aid), "", expiration, []byte(sid))
}

func (p *DefaultStrategy) deleteAuthzIndex(ctx Context, aid string) (err error) {
	return p.repository.Delete(ctx, p.authzIssuer, aid, getAuthzIndexPublicID(aid), "")
}

func (p *DefaultStrategy) newCookie(id []byte, expires time.Time) (cookie *http.Cookie) {
	return &http.Cookie{
		Name:     p.config.Name,
		Value:    encodeCookieID(id),
		Path:     "/",
		Domain:   p.domain,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func (p *DefaultStrategy) newAuthzCookie(id []byte, expires time.Time) (cookie *http.Cookie) {
	//nolint:gosec // The SameSite attribute is determined by the validated configuration which restricts it to 'none', 'lax', or 'strict'.
	return &http.Cookie{
		Name:     p.authzName,
		Value:    encodeCookieID(id),
		Path:     "/",
		Domain:   p.authzDomain,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: p.authzSameSite,
	}
}

func (p *DefaultStrategy) newAuthzDeletionCookie() (cookie *http.Cookie) {
	return p.newAuthzCookie(nil, p.clock.Now().Add(-cookieDeletionOffset))
}

func (p *DefaultStrategy) newDeletionCookie() (cookie *http.Cookie) {
	return p.newCookie(nil, p.clock.Now().Add(-cookieDeletionOffset))
}

func (p *DefaultStrategy) getExpiration(userSession UserSession) (expiration time.Duration) {
	if userSession.KeepMeLoggedIn && !p.config.DisableRememberMe {
		return p.config.RememberMe
	}

	return p.config.Expiration
}

func (p *DefaultStrategy) getExpiresAt(expiration time.Duration) (expires int64) {
	if expiration == 0 {
		return 0
	}

	return p.clock.Now().Add(expiration).Unix()
}

func (p *DefaultStrategy) getRemaining(session UserSession) (remaining time.Duration, ok bool) {
	if session.Expires == 0 {
		return p.getExpiration(session), true
	}

	remaining = time.Unix(session.Expires, 0).Sub(p.clock.Now())

	return remaining, remaining > 0
}

func (p *DefaultStrategy) getCookieExpires(expiration time.Duration) (exp time.Time) {
	if expiration == 0 {
		return expireUnlimited
	}

	lifetime := max(expiration, sessionCookieLifetimeMinimum)

	if !p.config.DisableRememberMe {
		lifetime = max(lifetime, p.config.RememberMe)
	}

	return p.clock.Now().Add(lifetime)
}

func equalSignature(a, b string) bool {
	return len(a) != 0 && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func getAuthzIndexPublicID(aid string) string {
	if len(aid) <= authzIndexPublicIDLength {
		return aid
	}

	return aid[:authzIndexPublicIDLength]
}

func encodeCookieID(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}

func newDomain(value string, samSite http.SameSite) string {
	switch {
	case strings.HasPrefix(value, "."), samSite == http.SameSiteStrictMode:
		return value
	default:
		return "." + value
	}
}

func newSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
