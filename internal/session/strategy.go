// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// NewStrategy returns a new Strategy for the given session cookie configuration. The issuer is the signature of the
// Authelia URL, falling back to the cookie domain, which scopes every record this strategy stores. The session cookie
// is host only and always strict, and the configured same site value is applied to the authz cookie which is scoped to
// the cookie domain.
func NewStrategy(config schema.SessionCookie, clock clock.Provider, codec Codec, storage Repository) (provider Strategy) {
	sameSite := newSameSite(config.SameSite)

	value := config.Domain

	if config.AutheliaURL != nil {
		value = config.AutheliaURL.String()
	}

	scope := strings.ToLower(value)

	p := &DefaultStrategy{
		config:        config,
		issuer:        codec.Sign([]byte(scope)),
		name:          SessionCookieName(config.Name),
		authzName:     AuthzCookieName(config.Name),
		bindingName:   UserAgentBindingCookieName(config.Name),
		authzIssuer:   codec.SignAuthz([]byte(issuerPrefixAuthz + scope)),
		authzDomain:   newDomain(config.Domain, sameSite),
		authzSameSite: sameSite,
		clock:         clock,
		codec:         codec,
		repository:    storage,
	}

	p.authz = &DefaultAuthzStrategy{base: p}

	return p
}

// DefaultStrategy is the default Strategy which stores the session identifier in a cookie and the session data in a
// Repository, scoped to a single configured session cookie domain.
type DefaultStrategy struct {
	config schema.SessionCookie

	issuer string
	name   string

	authzName     string
	authzIssuer   string
	authzDomain   string
	authzSameSite http.SameSite

	bindingName string

	authz *DefaultAuthzStrategy

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

// Get returns the session of the request.
//
//   - A request to an authorization endpoint always receives an anonymous session, as the session cookie is not
//     honored there.
//   - A request with no session cookie, or a malformed one, receives an anonymous session.
//   - A session cookie which resolves to no record receives an anonymous session. This includes a session which
//     expired or was destroyed, and a cookie which a regeneration replaced.
//   - An error which wraps ErrRepositoryGet is returned when the Repository fails.
//   - An error is returned when the record can't be opened, is stored against another identifier, or is bound to
//     another cookie domain.
//   - The request is recorded as activity on the session when an inactivity timeout applies.
//   - An authenticated session which the authz cookie of the request isn't linked to is issued a new authz cookie. No
//     cookie is issued when the session has expired or the link can't be stored.
//   - An anonymous session has the authz cookie of the request discarded. The cookie is kept when it's linked to
//     another session or to a session which was moved to a new identifier, as another request may have already
//     replaced the cookies this request was made with.
func (p *DefaultStrategy) Get(ctx Context) (session *UserSession, err error) {
	if ctx.IsAuthzEndpointRequest() {
		userSession := p.NewDefault()

		return &userSession, nil
	}

	_, session, err = p.get(ctx, true)

	return session, err
}

// Save persists the session and issues its cookies.
//
//   - It returns an error which wraps ErrNotPermittedOnAuthzEndpointRequest on a request to an authorization endpoint.
//   - It returns an error when the session is nil or is bound to another cookie domain.
//   - The session is a pointer as the cookie domain, public identifier, CSRF secret, and expiry are written back to
//     it. The caller must retain the public identifier to avoid orphaning it on a later save.
//   - An anonymous session which holds no state isn't persisted. The record and cookies of the session of the
//     request are discarded instead, as the absence of a session is equivalent to it.
//   - A request with no session cookie has a new session identifier generated.
//   - The expiry of the session is extended from now.
//   - The session is persisted with the CSRF secret this request last established for it when there is one, as a
//     session retrieved before it was regenerated holds the secret the regeneration replaced. A session with no
//     secret takes the stored one, or has one generated.
//   - The session cookie and the CSRF cookie are issued.
//   - An authenticated session is issued the authz cookie. The save fails when the link for it can't be stored.
//   - An anonymous session is never issued the authz cookie, and any authz cookie of the request is discarded when
//     the request has a session cookie.
//   - A save of a session which was destroyed or regenerated after it was retrieved is dropped, and an error which
//     wraps ErrSessionSuperseded is returned, so a stale request can't restore a session.
//   - A new anonymous session whose session cookie refers to a destroyed or regenerated session is persisted against
//     a new identifier instead, so its state isn't lost. The authz cookie of the request is left as it is, as another
//     request may have already replaced it with the authz cookie of the regenerated session.
func (p *DefaultStrategy) Save(ctx Context, session *UserSession) (err error) {
	if session == nil {
		return errors.New("error occurred saving session: it is nil")
	}

	if ctx.IsAuthzEndpointRequest() {
		return fmt.Errorf("error occurred saving session: %w", ErrNotPermittedOnAuthzEndpointRequest)
	}

	switch {
	case session.CookieDomain == "":
		session.CookieDomain = p.config.Domain
	case session.CookieDomain != p.config.Domain:
		return errors.New("error occurred saving session: domain does not match cookie domain")
	}

	if !session.RequiresPersistence() {
		return p.discard(ctx, session)
	}

	// A session which has neither a user nor a public identifier was not loaded from a record, so being superseded means
	// the cookie refers to a session which was destroyed, or was moved before this request retrieved it.
	abandoned := session.Username == "" && session.PublicID == ""

	id := p.getCookieID(ctx, p.name)

	existing := len(id) != 0

	if !existing {
		if id, err = p.codec.GenerateSessionID(); err != nil {
			return fmt.Errorf("error occurred generating session ID: %w", err)
		}
	}

	if session.PublicID == "" {
		if session.PublicID, err = p.codec.GeneratePublicID(); err != nil {
			return fmt.Errorf("error occurred generating session public ID: %w", err)
		}
	}

	p.adoptCSRFSecret(ctx, session)

	if err = p.ensureCSRFSecret(ctx, id, existing, session); err != nil {
		return err
	}

	if err = p.persist(ctx, id, existing, session); !errors.Is(err, ErrSessionSuperseded) || !abandoned {
		return err
	}

	if id, err = p.codec.GenerateSessionID(); err != nil {
		return fmt.Errorf("error occurred generating session ID: %w", err)
	}

	// The new identifier was never issued, so the authz cookie of the request is left as it is. Another request may have
	// already replaced it with the authz cookie of the session this one was moved to.
	return p.persist(ctx, id, false, session)
}

func (p *DefaultStrategy) persist(ctx Context, id []byte, existing bool, session *UserSession) (err error) {
	var data []byte

	sid := p.codec.Sign(id)

	expiration := p.getExpiration(*session)
	session.Expires = p.getExpiresAt(expiration)

	if data, err = p.codec.Seal(p.config.Domain, sid, *session); err != nil {
		return fmt.Errorf("error occurred encoding session: %w", err)
	}

	expires := p.getCookieExpires(expiration)

	var (
		authz []byte
		aid   string
	)

	if !session.IsAnonymous() {
		authz, aid = p.getAuthzValue(sid)

		if err = p.saveAuthzIndex(ctx, aid, sid, expiration); err != nil {
			return fmt.Errorf("error occurred saving session authz index: %w", err)
		}
	}

	if err = p.repository.Save(ctx, p.issuer, sid, session.PublicID, session.Username, p.getRetention(*session, expiration), data); err != nil {
		if len(aid) != 0 && (!existing || errors.Is(err, ErrSessionSuperseded)) {
			_ = p.deleteAuthzIndex(ctx, aid)
		}

		return fmt.Errorf("error occurred saving session to registry: %w", err)
	}

	p.setCached(ctx, p.cacheKey(), session)
	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	ctx.SetCookie(p.newCookie(id, expires))
	ctx.SetCookie(p.newCSRFCookie(session.CSRF, expires))

	switch {
	case len(authz) != 0:
		ctx.SetCookie(p.newAuthzCookie(authz, expires))
	case existing:
		p.clearAuthzCookie(ctx)
	}

	return nil
}

// Regenerate issues a new session identifier, and with it a new authz cookie, for the current session, preserving the
// session data and public identifier, which mitigates session fixation across an authentication level change. A session
// which can't be read is deleted and replaced with an anonymous session as none of its data can be reused, whereas a
// failure to retrieve it from the Repository is returned as the session may be valid. An established session which was
// destroyed or regenerated by another request after it was read is never recreated, and ErrSessionSuperseded is
// returned instead. The previous session cookie never resolves to the regenerated session. It returns
// ErrNotPermittedOnAuthzEndpointRequest on a request to an authorization endpoint.
func (p *DefaultStrategy) Regenerate(ctx Context) (err error) {
	if ctx.IsAuthzEndpointRequest() {
		return fmt.Errorf("error occurred regenerating session: %w", ErrNotPermittedOnAuthzEndpointRequest)
	}

	var (
		oldSID    string
		id, authz []byte
		session   *UserSession
	)

	if oldSID, session, err = p.getRegenerate(ctx); err != nil {
		return err
	}

	if id, err = p.codec.GenerateSessionID(); err != nil {
		return fmt.Errorf("error occurred generating session ID: %w", err)
	}

	sid := p.codec.Sign(id)
	expiration := p.getExpiration(*session)
	expires := p.getCookieExpires(expiration)

	if len(oldSID) != 0 {
		if authz, err = p.move(ctx, oldSID, sid, expiration, session); err != nil {
			return err
		}
	}

	ctx.SetCookie(p.newCookie(id, expires))

	if len(session.CSRF) != 0 {
		ctx.SetCookie(p.newCSRFCookie(session.CSRF, expires))
	}

	switch {
	case len(authz) != 0:
		ctx.SetCookie(p.newAuthzCookie(authz, expires))
	case len(oldSID) != 0 && session.IsAnonymous():
		p.clearAuthzCookie(ctx)
	}

	return nil
}

func (p *DefaultStrategy) move(ctx Context, oldSID, sid string, expiration time.Duration, session *UserSession) (authz []byte, err error) {
	var data []byte

	session.Expires = p.getExpiresAt(expiration)

	if session.PublicID != "" {
		if err = p.generateCSRFSecret(session); err != nil {
			return nil, err
		}
	}

	if data, err = p.codec.Seal(p.config.Domain, sid, *session); err != nil {
		return nil, fmt.Errorf("error occurred encoding session: %w", err)
	}

	switch err = p.repository.ChangeID(ctx, p.issuer, oldSID, sid, session.PublicID, session.Username, p.getRetention(*session, expiration), data); {
	case err == nil:
		if !session.IsAnonymous() {
			p.retireAuthzIndex(ctx, oldSID)
		}
	case session.PublicID != "" || !errors.Is(err, ErrSessionSuperseded):
		p.setCached(ctx, p.cacheKey(), nil)
		p.setCached(ctx, p.cacheKeyAuthz(), nil)

		return nil, fmt.Errorf("error occurred changing session ID: %w", err)
	}

	p.setCached(ctx, p.cacheKey(), session)
	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	if session.IsAnonymous() {
		return nil, nil
	}

	authz, aid := p.getAuthzValue(sid)

	if p.saveAuthzIndex(ctx, aid, sid, expiration) != nil {
		return nil, nil
	}

	return authz, nil
}

func (p *DefaultStrategy) getRegenerate(ctx Context) (id string, session *UserSession, err error) {
	if id, session, err = p.get(ctx, false); err == nil || errors.Is(err, ErrRepositoryGet) {
		return id, session, err
	}

	if err = p.repository.Delete(ctx, p.issuer, id, "", ""); err != nil {
		return "", nil, fmt.Errorf("error occurred deleting unreadable session from backend: %w", err)
	}

	_, aid := p.getAuthzValue(id)

	_ = p.deleteAuthzIndex(ctx, aid)

	p.setCached(ctx, p.cacheKey(), nil)
	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	p.clearAuthzCookie(ctx)

	anonymous := p.NewDefault()

	return "", &anonymous, nil
}

// Destroy removes the backend record for the current session and instructs the user agent to discard the session
// cookie, the authz cookie, and the CSRF cookie. A request which has no session cookie has the session its authz cookie is linked to
// removed instead. It returns ErrNotPermittedOnAuthzEndpointRequest on a request to an authorization endpoint.
func (p *DefaultStrategy) Destroy(ctx Context) (err error) {
	if ctx.IsAuthzEndpointRequest() {
		return fmt.Errorf("error occurred destroying session: %w", ErrNotPermittedOnAuthzEndpointRequest)
	}

	id, userSession, _ := p.get(ctx, false)

	if len(id) == 0 {
		if err = p.authz.remove(ctx); err != nil {
			err = fmt.Errorf("error occurred destroying session: %w", err)
		}
	} else {
		err = p.destroyByID(ctx, id, userSession)
	}

	if err != nil {
		return err
	}

	p.setCached(ctx, p.cacheKey(), nil)
	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	p.clearCookies(ctx)

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

func (p *DefaultStrategy) discard(ctx Context, session *UserSession) (err error) {
	id := p.getCookieID(ctx, p.name)

	if len(id) == 0 {
		return nil
	}

	if err = p.destroyByID(ctx, p.codec.Sign(id), session); err != nil {
		return err
	}

	p.setCached(ctx, p.cacheKey(), nil)
	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	p.clearCookies(ctx)

	return nil
}

func (p *DefaultStrategy) get(ctx Context, heal bool) (id string, session *UserSession, err error) {
	cookie := p.getCookieID(ctx, p.name)

	userSession := p.NewDefault()

	if len(cookie) == 0 {
		return "", &userSession, nil
	}

	id = p.codec.Sign(cookie)

	if cached, ok := p.getCached(ctx, p.cacheKey()); ok {
		if heal {
			p.heal(ctx, id, cached)
		}

		return id, cached, nil
	}

	var record Record

	if record, err = p.repository.Get(ctx, p.issuer, id); err != nil {
		return id, nil, fmt.Errorf("%w: %w", ErrRepositoryGet, err)
	}

	if record == nil || len(record.GetSessionData()) == 0 {
		logContext(ctx).Debug("Session cookie doesn't resolve to a session, which has either expired, exceeded the inactivity timeout, or been destroyed")

		if heal {
			p.clearStaleAuthzCookie(ctx, id)
		}

		p.setCached(ctx, p.cacheKey(), &userSession)

		return id, &userSession, nil
	}

	if record.GetSessionSignature() != id {
		return id, nil, errors.New("error occurred getting session: signature does not match the session identifier")
	}

	session = &UserSession{}

	if err = p.codec.Open(p.config.Domain, record, session); err != nil {
		return id, nil, fmt.Errorf("error occurred decoding session: %w", err)
	}

	if session.CookieDomain != p.config.Domain {
		return id, nil, errors.New("error occurred getting session: domain does not match cookie domain")
	}

	if heal {
		p.heal(ctx, id, session)
	}

	p.touch(ctx, id, record, session)

	p.setCached(ctx, p.cacheKey(), session)

	return id, session, nil
}

func (p *DefaultStrategy) heal(ctx Context, sid string, session *UserSession) {
	if session.IsAnonymous() {
		p.clearStaleAuthzCookie(ctx, sid)

		return
	}

	p.healAuthz(ctx, sid, session)
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

func logContext(ctx Context) (logger *logrus.Entry) {
	if lctx, ok := ctx.(LoggingContext); ok {
		return lctx.GetLogger()
	}

	return logrus.NewEntry(logging.Logger())
}
