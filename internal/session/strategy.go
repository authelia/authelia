// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewStrategy returns a new Strategy for the given session cookie configuration. The issuer is the signature of the
// Authelia URL, falling back to the cookie domain, which scopes every record this strategy stores.
func NewStrategy(config schema.SessionCookie, clock clock.Provider, codec Codec, storage Repository) (provider Strategy) {
	sameSite := newSameSite(config.SameSite)

	var issuer string

	if config.AutheliaURL == nil {
		issuer = codec.Sign([]byte(strings.ToLower(config.Domain)))
	} else {
		issuer = codec.Sign([]byte(strings.ToLower(config.AutheliaURL.String())))
	}

	return &DefaultStrategy{
		config:     config,
		issuer:     issuer,
		domain:     newDomain(config.Domain, sameSite),
		samesite:   sameSite,
		clock:      clock,
		codec:      codec,
		repository: storage,
	}
}

// DefaultStrategy is the default Strategy which stores the session identifier in a cookie and the session data in a
// Repository, scoped to a single configured session cookie domain.
type DefaultStrategy struct {
	config schema.SessionCookie

	issuer   string
	domain   string
	samesite http.SameSite

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

	// The session records the configured domain rather than the cookie domain, as the leading dot which may be present
	// on the latter is a cookie attribute concern and consumers compare this against the configured domain.
	userSession.CookieDomain = p.config.Domain

	return userSession
}

// NewDefault returns an anonymous session bound to this strategies cookie domain.
func (p *DefaultStrategy) NewDefault() (userSession UserSession) {
	return p.New("")
}

// Get returns the session for the given request, returning an anonymous session when the request has no session
// cookie or the cookie does not resolve to a readable record.
func (p *DefaultStrategy) Get(ctx Context) (session *UserSession, err error) {
	_, session, err = p.get(ctx)

	return session, err
}

// Save persists the session. The userSession is a pointer as a session which has no public identifier has one generated
// for it, and the caller must observe it to avoid orphaning the generated identifier on a subsequent save.
func (p *DefaultStrategy) Save(ctx Context, session *UserSession) (err error) {
	if session == nil {
		return fmt.Errorf("error occurred saving session: it is nil")
	}

	switch {
	case session.CookieDomain == "":
		session.CookieDomain = p.config.Domain
	case session.CookieDomain != p.config.Domain:
		return fmt.Errorf("error occurred saving session: domain does not match cookie domain")
	}

	var id, data []byte

	if id = p.getCookieID(ctx); len(id) == 0 {
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

	if data, err = p.codec.Seal(p.config.Domain, sid, *session); err != nil {
		return fmt.Errorf("error occurred encoding session: %w", err)
	}

	expiration := p.getExpiration(*session)

	cookie := p.newCookie(id, p.getExpires(expiration))

	if err = p.repository.Save(ctx, p.issuer, sid, session.PublicID, session.Username, expiration, data); err != nil {
		return fmt.Errorf("error occurred saving session to registry: %w", err)
	}

	p.setCached(ctx, session)

	ctx.SetCookie(cookie)

	return nil
}

// Regenerate issues a new session identifier for the current session, preserving the session data and public
// identifier, which mitigates session fixation across an authentication level change.
func (p *DefaultStrategy) Regenerate(ctx Context) (err error) {
	var (
		oldSID  string
		id      []byte
		session *UserSession
	)

	if oldSID, session, err = p.get(ctx); err != nil {
		return err
	}

	if id, err = p.codec.GenerateSessionID(); err != nil {
		return fmt.Errorf("error occurred generating session ID: %w", err)
	}

	sid := p.codec.Sign(id)
	expiration := p.getExpiration(*session)

	// An anonymous request has no persisted session to rename, so issuing the new identifier is sufficient.
	if len(oldSID) != 0 {
		var data []byte

		if data, err = p.codec.Seal(p.config.Domain, sid, *session); err != nil {
			return fmt.Errorf("error occurred encoding session: %w", err)
		}

		if err = p.repository.ChangeID(ctx, p.issuer, oldSID, sid, session.PublicID, session.Username, expiration, data); err != nil {
			return fmt.Errorf("error occurred changing session ID: %w", err)
		}
	}

	ctx.SetCookie(p.newCookie(id, p.getExpires(expiration)))

	return nil
}

// Destroy removes the backend record for the current session and instructs the user agent to discard the cookie.
func (p *DefaultStrategy) Destroy(ctx Context) (err error) {
	id, userSession, _ := p.get(ctx)

	if len(id) != 0 {
		var pid, username string

		if userSession != nil {
			pid = userSession.PublicID
			username = userSession.Username
		}

		if err = p.repository.Delete(ctx, p.issuer, id, pid, username); err != nil {
			return fmt.Errorf("error occurred deleting session from backend: %w", err)
		}
	}

	p.setCached(ctx, nil)

	ctx.ClearCookie(p.newDeletionCookie())

	return nil
}

func (p *DefaultStrategy) get(ctx Context) (id string, session *UserSession, err error) {
	cookie := p.getCookieID(ctx)

	userSession := p.NewDefault()

	if len(cookie) == 0 {
		return "", &userSession, nil
	}

	id = p.codec.Sign(cookie)

	if cached, ok := p.getCached(ctx); ok {
		return id, cached, nil
	}

	var record Record

	if record, err = p.repository.Get(ctx, p.issuer, id); err != nil {
		return id, nil, fmt.Errorf("%w: %w", ErrRepositoryGet, err)
	}

	if record == nil || len(record.GetSessionData()) == 0 {
		p.setCached(ctx, &userSession)

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

	p.setCached(ctx, session)

	return id, session, nil
}

func (p *DefaultStrategy) getCached(ctx Context) (session *UserSession, ok bool) {
	caching, ok := ctx.(CachingContext)
	if !ok {
		return nil, false
	}

	cached, ok := caching.CachedSession(p.config.Domain)
	if !ok || cached == nil {
		return nil, false
	}

	value := cached.deepCopy()

	return &value, true
}

func (p *DefaultStrategy) setCached(ctx Context, session *UserSession) {
	caching, ok := ctx.(CachingContext)
	if !ok {
		return
	}

	if session == nil {
		caching.CacheSession(p.config.Domain, nil)

		return
	}

	value := session.deepCopy()

	caching.CacheSession(p.config.Domain, &value)
}

func (p *DefaultStrategy) getCookieID(ctx Context) (id []byte) {
	value := ctx.GetCookie(p.config.Name)

	if len(value) == 0 {
		return nil
	}

	var err error

	if id, err = base64.RawURLEncoding.DecodeString(value); err != nil {
		return nil
	}

	return id
}

func (p *DefaultStrategy) newCookie(id []byte, expires time.Time) (cookie *http.Cookie) {
	//nolint:gosec // The SameSite attribute is determined by the validated configuration which restricts it to 'none', 'lax', or 'strict'.
	return &http.Cookie{
		Name:     p.config.Name,
		Value:    encodeCookieID(id),
		Path:     "/",
		Domain:   p.domain,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: p.samesite,
	}
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

func (p *DefaultStrategy) getExpires(expiration time.Duration) (exp time.Time) {
	if expiration == 0 {
		return expireUnlimited
	}

	return p.clock.Now().Add(expiration)
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
