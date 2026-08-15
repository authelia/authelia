package session

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

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

type DefaultStrategy struct {
	config schema.SessionCookie

	issuer   string
	domain   string
	samesite http.SameSite

	clock clock.Provider

	codec      Codec
	repository Repository
}

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

func (p *DefaultStrategy) NewDefault() (userSession UserSession) {
	return p.New("")
}

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

	var (
		id   string
		data []byte
	)

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

	if data, err = p.codec.Seal(p.config.Domain, *session); err != nil {
		return fmt.Errorf("error occurred encoding session: %w", err)
	}

	expiration := p.getExpiration(*session)

	cookie := p.newCookie(id, p.getExpires(expiration))

	sid := p.codec.Sign([]byte(id))

	if err = p.repository.Save(ctx, p.issuer, sid, session.PublicID, session.Username, expiration, data); err != nil {
		return fmt.Errorf("error occurred saving session to registry: %w", err)
	}

	ctx.SetCookie(cookie)

	return nil
}

func (p *DefaultStrategy) Regenerate(ctx Context) (err error) {
	var (
		oldSID, id string
		session    *UserSession
	)

	if oldSID, session, err = p.get(ctx); err != nil {
		return err
	}

	if id, err = p.codec.GenerateSessionID(); err != nil {
		return fmt.Errorf("error occurred generating session ID: %w", err)
	}

	sid := p.codec.Sign([]byte(id))
	expiration := p.getExpiration(*session)

	// An anonymous request has no persisted session to rename, so issuing the new identifier is sufficient.
	if len(oldSID) != 0 {
		if err = p.repository.ChangeID(ctx, p.issuer, oldSID, sid, session.PublicID, session.Username, expiration); err != nil {
			return fmt.Errorf("error occurred changing session ID: %w", err)
		}
	}

	ctx.SetCookie(p.newCookie(id, p.getExpires(expiration)))

	return nil
}

func (p *DefaultStrategy) Destroy(ctx Context) (err error) {
	defer ctx.ClearCookie(p.newDeletionCookie())

	// The get error is deliberately discarded. A session which can't be decoded, such as one sealed with a rotated
	// secret, or which records a different cookie domain, still has a backend record which must be removed or it
	// outlives the cookie. Once the record is gone the session is destroyed regardless of whether it was readable, and
	// a genuine backend failure still surfaces via the delete below.
	id, userSession, _ := p.get(ctx)

	// A request without a session cookie has no backend record to remove.
	if len(id) == 0 {
		return nil
	}

	var pid, username string

	if userSession != nil {
		pid = userSession.PublicID
		username = userSession.Username
	}

	if err = p.repository.Delete(ctx, p.issuer, id, pid, username); err != nil {
		return fmt.Errorf("error occurred deleting session from backend: %w", err)
	}

	return nil
}

func (p *DefaultStrategy) get(ctx Context) (id string, session *UserSession, err error) {
	cookie := p.getCookieID(ctx)

	userSession := p.NewDefault()

	// A request without a session cookie is anonymous, so a new default session is returned rather than an error.
	if len(cookie) == 0 {
		return "", &userSession, nil
	}

	id = p.codec.Sign([]byte(cookie))

	var data []byte

	if data, err = p.repository.Get(ctx, p.issuer, id); err != nil {
		return id, nil, fmt.Errorf("error occurred getting session from backend: %w", err)
	}

	// The backend returns no data and no error when the session is unknown or has expired, which is treated the same
	// as an anonymous request rather than an error.
	if len(data) == 0 {
		return id, &userSession, nil
	}

	session = &UserSession{}

	if err = p.codec.Open(p.config.Domain, session, data); err != nil {
		return id, nil, fmt.Errorf("error occurred decoding session: %w", err)
	}

	if session.CookieDomain != p.config.Domain {
		return id, nil, fmt.Errorf("error occurred getting session: domain does not match cookie domain")
	}

	return id, session, nil
}

func (p *DefaultStrategy) getCookieID(ctx Context) (id string) {
	return ctx.GetCookie(p.config.Name)
}

func (p *DefaultStrategy) newCookie(id string, expires time.Time) (cookie *http.Cookie) {
	//nolint:gosec // The SameSite attribute is determined by the validated configuration which restricts it to 'none', 'lax', or 'strict'.
	return &http.Cookie{
		Name:     p.config.Name,
		Value:    id,
		Path:     "/",
		Domain:   p.domain,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: p.samesite,
	}
}

// newDeletionCookie returns an expired form of the session cookie which instructs the user agent to discard it. It
// deliberately reuses newCookie so the name, domain, and path always match the cookie which was set, as user agents key
// cookies on all three and would otherwise retain the original alongside this one.
func (p *DefaultStrategy) newDeletionCookie() (cookie *http.Cookie) {
	return p.newCookie("", p.clock.Now().Add(-cookieDeletionOffset))
}

// getExpiration returns the session expiration, taking the remember me preference of the session into account.
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
