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
	"net"
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

	return &DefaultStrategy{
		config:        config,
		issuer:        codec.Sign([]byte(scope)),
		name:          SessionCookieName(config.Name),
		authzName:     AuthzCookieName(config.Name),
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
	name   string

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
// request isn't linked to is issued a new authz cookie, and the user agent is instructed to discard the authz cookie
// when the session cookie resolves to an anonymous session. A request to an authorization endpoint always receives an
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
// An anonymous session which holds no state isn't persisted, and the records and cookies of any previously persisted
// session for the request are discarded instead, as the absence of a session is equivalent to it. An anonymous session
// is never issued the authz cookie, and the user agent is instructed to discard any authz cookie it holds for the
// session. A save which was superseded is dropped so a stale request can't restore a session, except for a new
// anonymous session whose cookie refers to a destroyed session, which is given a new identifier so its state isn't
// lost. The session is persisted with the CSRF secret this request last established for it rather than the one it
// holds, as a session retrieved before it was regenerated holds the secret the regeneration replaced. It returns
// ErrNotPermittedOnAuthzEndpointRequest on a request to an authorization endpoint.
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
	// the cookie refers to a destroyed session rather than the save having lost a race with a login or regeneration.
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

	if err = p.persist(ctx, id, existing, session); !errors.Is(err, ErrSessionSuperseded) {
		return err
	}

	if !abandoned {
		return nil
	}

	if id, err = p.codec.GenerateSessionID(); err != nil {
		return fmt.Errorf("error occurred generating session ID: %w", err)
	}

	return p.persist(ctx, id, existing, session)
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

	var authz []byte

	if !session.IsAnonymous() {
		var aid string

		authz, aid = p.getAuthzValue(sid)

		if err = p.saveAuthzIndex(ctx, aid, sid, expiration); err != nil {
			return fmt.Errorf("error occurred saving session authz index: %w", err)
		}
	}

	if err = p.repository.Save(ctx, p.issuer, sid, session.PublicID, session.Username, p.getRetention(*session, expiration), data); err != nil {
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

// Regenerate issues a new session identifier and a new authz identifier for the current session, preserving the
// session data and public identifier, which mitigates session fixation across an authentication level change. A session
// which can't be read is deleted and replaced with an anonymous session as none of its data can be reused, whereas a
// failure to retrieve it from the Repository is returned as the session may be valid. It returns
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
		var data []byte

		session.Expires = p.getExpiresAt(expiration)

		// The CSRF secret is regenerated alongside the session identifier of an established session, so a token issued
		// before an authentication level change isn't accepted after it.
		if session.PublicID != "" {
			if err = p.generateCSRFSecret(session); err != nil {
				return err
			}
		}

		if data, err = p.codec.Seal(p.config.Domain, sid, *session); err != nil {
			return fmt.Errorf("error occurred encoding session: %w", err)
		}

		if err = p.repository.ChangeID(ctx, p.issuer, oldSID, sid, session.PublicID, session.Username, p.getRetention(*session, expiration), data); err != nil {
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
	ctx.ClearCookie(p.newCSRFDeletionCookie())

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
	sid, aid, userSession, err := (&DefaultAuthzStrategy{base: p}).resolve(ctx)
	if err != nil {
		return fmt.Errorf("error occurred destroying session: %w", err)
	}

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

// discard removes the records and cookies of the request for a session which doesn't require persistence. The cookies
// are cleared rather than retained as the Repository refuses later saves against a deleted record, which would otherwise
// prevent the request from persisting a session again until the deleted record expired.
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

	ctx.ClearCookie(p.newDeletionCookie())
	ctx.ClearCookie(p.newAuthzDeletionCookie())
	ctx.ClearCookie(p.newCSRFDeletionCookie())

	return nil
}

// CSRFToken returns the CSRF token of the established session of the request. The token is the CSRF secret stored in
// the session signed alongside the cookie domain, so the secret itself is never delivered to the user agent. It returns
// ErrCSRFTokenNoSession when the request has no established session, and ErrCSRFTokenInvalid when the established
// session has no CSRF secret.
func (p *DefaultStrategy) CSRFToken(ctx Context) (token string, err error) {
	var session *UserSession

	if session, err = p.getCSRFSession(ctx); err != nil {
		return "", err
	}

	if len(session.CSRF) == 0 {
		return "", ErrCSRFTokenInvalid
	}

	return p.csrfToken(session.CSRF), nil
}

// VerifyCSRFToken returns nil if the token is the CSRF token of the established session of the request, comparing them
// in constant time. It returns ErrCSRFTokenNoSession when the request has no established session, ErrCSRFTokenInvalid
// when the token doesn't match, and any other error when the session couldn't be retrieved.
func (p *DefaultStrategy) VerifyCSRFToken(ctx Context, token string) (err error) {
	var session *UserSession

	if session, err = p.getCSRFSession(ctx); err != nil {
		return err
	}

	if len(session.CSRF) == 0 || len(token) == 0 || !p.codec.VerifyCSRF(p.csrfTokenMessage(session.CSRF), token) {
		return ErrCSRFTokenInvalid
	}

	return nil
}

// RegenerateCSRFToken replaces the CSRF secret of the established session of the request without changing the session
// identifier, persisting the session and delivering the new token to the user agent. Tokens issued before it are no
// longer accepted. It returns ErrCSRFTokenNoSession when the request has no established session.
func (p *DefaultStrategy) RegenerateCSRFToken(ctx Context) (err error) {
	var session *UserSession

	if session, err = p.getCSRFSession(ctx); err != nil {
		return err
	}

	if err = p.generateCSRFSecret(session); err != nil {
		return err
	}

	p.setCached(ctx, p.cacheKey(), session)

	return p.Save(ctx, session)
}

// SetCSRFCookie delivers the CSRF token of the established session of the request to the user agent, generating a CSRF
// secret for an established session which has none. A request without an established session has no token to deliver,
// which isn't an error.
func (p *DefaultStrategy) SetCSRFCookie(ctx Context) (err error) {
	var session *UserSession

	if session, err = p.getCSRFSession(ctx); err != nil {
		if errors.Is(err, ErrCSRFTokenNoSession) {
			return nil
		}

		return err
	}

	if len(session.CSRF) == 0 {
		return p.Save(ctx, session)
	}

	ctx.SetCookie(p.newCSRFCookie(session.CSRF, p.getCookieExpires(p.getExpiration(*session))))

	return nil
}

// getCSRFSession returns the established session of the request. A request to an authorization endpoint never has one
// as the session cookie is not honored there, and neither does a request whose session can't be read, as nothing can
// act on its behalf and the request may be the one which replaces it such as a login.
func (p *DefaultStrategy) getCSRFSession(ctx Context) (session *UserSession, err error) {
	if ctx.IsAuthzEndpointRequest() {
		return nil, ErrCSRFTokenNoSession
	}

	if _, session, err = p.get(ctx, false); err != nil {
		if errors.Is(err, ErrRepositoryGet) {
			return nil, err
		}

		return nil, ErrCSRFTokenNoSession
	}

	if session == nil || session.PublicID == "" {
		return nil, ErrCSRFTokenNoSession
	}

	return session, nil
}

func (p *DefaultStrategy) adoptCSRFSecret(ctx Context, session *UserSession) {
	if cached, ok := p.getCached(ctx, p.cacheKey()); ok && len(cached.CSRF) != 0 && cached.PublicID == session.PublicID {
		session.CSRF = cached.CSRF
	}
}

// ensureCSRFSecret gives the session a CSRF secret when it has none. A session which is already persisted has its secret
// established by establishCSRFSecret, as concurrent requests for it must all deliver the same secret.
func (p *DefaultStrategy) ensureCSRFSecret(ctx Context, id []byte, existing bool, session *UserSession) (err error) {
	if len(session.CSRF) != 0 {
		return nil
	}

	if existing {
		return p.establishCSRFSecret(ctx, p.codec.Sign(id), session)
	}

	return p.generateCSRFSecret(session)
}

// establishCSRFSecret gives the session the CSRF secret of its persisted record, adding one to the record when it has
// none. The secret is added by a compare and swap, so when concurrent requests add one only the first succeeds and the
// others adopt its secret, rather than each delivering a secret to the user agent which only one of them persisted. A
// record which doesn't exist or can't be opened has nothing to agree with, so the session is given a new secret.
func (p *DefaultStrategy) establishCSRFSecret(ctx Context, sid string, session *UserSession) (err error) {
	for range csrfSecretEstablishAttempts {
		var (
			record Record
			data   []byte
		)

		if record, err = p.repository.Get(ctx, p.issuer, sid); err != nil {
			return fmt.Errorf("%w: %w", ErrRepositoryGet, err)
		}

		stored := &UserSession{}

		if record == nil || len(record.GetSessionData()) == 0 || record.GetSessionSignature() != sid || p.codec.Open(p.config.Domain, record, stored) != nil {
			return p.generateCSRFSecret(session)
		}

		if len(stored.CSRF) != 0 {
			session.CSRF = stored.CSRF

			return nil
		}

		remaining, ok := p.getRemaining(*stored)
		if !ok {
			return p.generateCSRFSecret(session)
		}

		if err = p.generateCSRFSecret(stored); err != nil {
			return err
		}

		if data, err = p.codec.Seal(p.config.Domain, sid, *stored); err != nil {
			return fmt.Errorf("error occurred encoding session: %w", err)
		}

		switch err = p.repository.SaveData(ctx, p.issuer, sid, stored.PublicID, stored.Username, p.getRetention(*stored, remaining), record.GetSessionData(), data); {
		case err == nil:
			session.CSRF = stored.CSRF

			return nil
		case !errors.Is(err, ErrSessionSuperseded):
			return fmt.Errorf("error occurred saving session CSRF secret: %w", err)
		}
	}

	return fmt.Errorf("error occurred establishing session CSRF secret: %w", ErrSessionSuperseded)
}

func (p *DefaultStrategy) generateCSRFSecret(session *UserSession) (err error) {
	if session.CSRF, err = p.codec.GenerateCSRFSecret(); err != nil {
		return fmt.Errorf("error occurred generating session CSRF secret: %w", err)
	}

	return nil
}

func (p *DefaultStrategy) csrfToken(secret []byte) (token string) {
	return p.codec.SignCSRF(p.csrfTokenMessage(secret))
}

func (p *DefaultStrategy) csrfTokenMessage(secret []byte) (message []byte) {
	message = make([]byte, 0, len(cookieValuePrefixCSRFToken)+len(p.config.Domain)+1+len(secret))

	message = append(message, cookieValuePrefixCSRFToken...)
	message = append(message, p.config.Domain...)
	message = append(message, ':')

	return append(message, secret...)
}

// SessionCookieName returns the name of the session cookie for the given configured cookie name. The session cookie is
// only honored on the Authelia URL host and is never read by scripts, so it's host only and HTTP only and carries the
// prefix which requires it to be both.
func SessionCookieName(name string) string {
	return CookieNamePrefixHostHTTP + name
}

// AuthzCookieName returns the name of the authz cookie for the given configured cookie name. The authz cookie is scoped
// to the cookie domain, so it carries the prefix which only requires it to be secure.
func AuthzCookieName(name string) string {
	return CookieNamePrefixSecure + name + CookieNameSuffixAuthz
}

// CSRFCookieName returns the name of the cookie which delivers the CSRF token of the session stored in the cookie with
// the given name to the user agent.
func CSRFCookieName(name string) string {
	return CookieNamePrefixHost + name + CookieNameSuffixCSRFToken
}

func (p *DefaultStrategy) newCSRFCookie(secret []byte, expires time.Time) (cookie *http.Cookie) {
	return p.newCSRFCookieValue(p.csrfToken(secret), expires)
}

func (p *DefaultStrategy) newCSRFDeletionCookie() (cookie *http.Cookie) {
	return p.newCSRFCookieValue("", p.clock.Now().Add(-cookieDeletionOffset))
}

func (p *DefaultStrategy) newCSRFCookieValue(value string, expires time.Time) (cookie *http.Cookie) {
	//nolint:gosec // The cookie is deliberately readable by scripts, and carries a token derived from the session rather than the session identifier.
	return &http.Cookie{
		Name:     CSRFCookieName(p.config.Name),
		Value:    value,
		Path:     "/",
		Expires:  expires,
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	}
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
			p.clearAuthzCookie(ctx)
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
		p.clearAuthzCookie(ctx)

		return
	}

	p.healAuthz(ctx, sid, session)
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

func (p *DefaultStrategy) clearAuthzCookie(ctx Context) {
	if ctx.GetCookie(p.authzName) == "" {
		return
	}

	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	ctx.ClearCookie(p.newAuthzDeletionCookie())
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
	value, _ = hex.DecodeString(p.codec.SignAuthz([]byte(cookieValuePrefixAuthz + p.issuer + ":" + sid)))

	return value, p.codec.SignAuthz(value)
}

// saveAuthzIndex stores the record which links the authz cookie to the session. A failure is logged as the callers
// recover from it by withholding the authz cookie, which leaves the session usable on the portal but not on the
// authorization endpoints until a later request issues the authz cookie again.
func (p *DefaultStrategy) saveAuthzIndex(ctx Context, aid, sid string, expiration time.Duration) (err error) {
	if err = p.repository.Save(ctx, p.authzIssuer, aid, getAuthzIndexPublicID(aid), "", expiration, []byte(sid)); err != nil {
		logContext(ctx).WithError(err).Error("Error occurred saving the session authz index, the authz cookie will not be issued")
	}

	return err
}

// deleteAuthzIndex removes the record which links the authz cookie to the session. A failure is logged as the callers
// recover from it, as a record which outlives its session no longer resolves to it.
func (p *DefaultStrategy) deleteAuthzIndex(ctx Context, aid string) (err error) {
	if err = p.repository.Delete(ctx, p.authzIssuer, aid, getAuthzIndexPublicID(aid), ""); err != nil {
		logContext(ctx).WithError(err).Warn("Error occurred deleting the session authz index")
	}

	return err
}

func (p *DefaultStrategy) newCookie(id []byte, expires time.Time) (cookie *http.Cookie) {
	return &http.Cookie{
		Name:     p.name,
		Value:    encodeCookieID(id),
		Path:     "/",
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
	switch {
	case userSession.Username == "" && p.config.AnonymousExpiration > 0:
		return p.config.AnonymousExpiration
	case userSession.KeepMeLoggedIn && !p.config.DisableRememberMe:
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

// getInactivity returns the inactivity timeout which applies to the session. It doesn't apply to an anonymous session or
// a remembered session.
func (p *DefaultStrategy) getInactivity(session UserSession) (inactivity time.Duration, ok bool) {
	if p.config.Inactivity <= 0 || session.IsAnonymous() || (session.KeepMeLoggedIn && !p.config.DisableRememberMe) {
		return 0, false
	}

	return p.config.Inactivity, true
}

// getRetention returns how long the record of the session is retained for given how long remains until it expires,
// which is capped at the inactivity timeout when one applies to the session. The record therefore lapses once the
// session is inactive for that long unless activity extends it, which can't extend it beyond when the session expires.
// A zero remaining value indicates the session doesn't expire.
func (p *DefaultStrategy) getRetention(session UserSession, remaining time.Duration) (retention time.Duration) {
	inactivity, ok := p.getInactivity(session)

	switch {
	case !ok:
		return remaining
	case remaining <= 0:
		return inactivity
	default:
		return min(inactivity, remaining)
	}
}

// getTouchInterval returns how far activity must move the inactivity deadline of a session before its record is
// extended, so a session in use isn't written to on every request.
func (p *DefaultStrategy) getTouchInterval(inactivity time.Duration) (interval time.Duration) {
	if p.config.Expiration > 0 {
		inactivity = min(inactivity, p.config.Expiration)
	}

	return min(max(inactivity/sessionTouchDivisor, time.Second), inactivity/2)
}

// touch records activity on the session by extending its record to the inactivity timeout from now, capped at when the
// session expires. It only does so once the record would be extended by at least the touch interval, and never for a
// record which doesn't report when it expires. A session which reports no expiration while one is configured predates
// the expiration being configured, and isn't extended as doing so would allow activity to retain it indefinitely.
func (p *DefaultStrategy) touch(ctx Context, sid string, record Record, session *UserSession) {
	inactivity, ok := p.getInactivity(*session)
	if !ok || record.GetSessionTTL() <= 0 || (session.Expires == 0 && p.getExpiration(*session) != 0) {
		return
	}

	remaining, ok := p.getRemaining(*session)
	if !ok {
		return
	}

	retention := p.getRetention(*session, remaining)

	if retention-record.GetSessionTTL() < p.getTouchInterval(inactivity) {
		return
	}

	if err := p.repository.Touch(ctx, p.issuer, sid, session.PublicID, session.Username, retention); err != nil {
		logContext(ctx).WithError(err).Error("Error occurred extending the session for activity")
	}
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

func logContext(ctx Context) (logger *logrus.Entry) {
	if lctx, ok := ctx.(LoggingContext); ok {
		return lctx.GetLogger()
	}

	return logrus.NewEntry(logging.Logger())
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
	case strings.HasPrefix(value, "."), samSite == http.SameSiteStrictMode, net.ParseIP(value) != nil:
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
