package session

import (
	"encoding/json"
	"time"

	"github.com/fasthttp/session/v2"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Session a session provider.
type Session struct {
	Config schema.SessionCookie

	sessionHolder *session.Session
}

// NewDefaultUserSession returns a new default UserSession for this session provider.
func (p *Session) NewDefaultUserSession() (userSession UserSession) {
	userSession = NewDefaultUserSession()

	userSession.CookieDomain = p.Config.Domain

	return userSession
}

// GetSession return the user session from a request.
func (p *Session) GetSession(ctx *fasthttp.RequestCtx) (userSession UserSession, err error) {
	var store *session.Store

	if store, err = p.sessionHolder.Get(ctx); err != nil {
		return p.NewDefaultUserSession(), err
	}

	userSessionJSON, ok := store.Get(userSessionStorerKey).([]byte)

	// If userSession is not yet defined we create the new session with default values
	// and save it in the store.
	if !ok {
		userSession = p.NewDefaultUserSession()

		store.Set(userSessionStorerKey, userSession)

		return userSession, nil
	}

	if err = json.Unmarshal(userSessionJSON, &userSession); err != nil {
		return p.NewDefaultUserSession(), err
	}

	return userSession, nil
}

// SaveSession save the user session.
func (p *Session) SaveSession(ctx *fasthttp.RequestCtx, userSession UserSession) (err error) {
	var (
		store           *session.Store
		userSessionJSON []byte
	)

	if store, err = p.sessionHolder.Get(ctx); err != nil {
		return err
	}

	if userSessionJSON, err = json.Marshal(userSession); err != nil {
		return err
	}

	store.Set(userSessionStorerKey, userSessionJSON)

	if err = p.sessionHolder.Save(ctx, store); err != nil {
		return err
	}

	return nil
}

// RegenerateSession regenerate a session ID.
func (p *Session) RegenerateSession(ctx *fasthttp.RequestCtx) error {
	return p.sessionHolder.Regenerate(ctx)
}

// DestroySession destroy a session ID and delete the cookie.
func (p *Session) DestroySession(ctx *fasthttp.RequestCtx) error {
	return p.sessionHolder.Destroy(ctx)
}

// UpdateExpiration update the expiration of the cookie and session.
func (p *Session) UpdateExpiration(ctx *fasthttp.RequestCtx, expiration time.Duration) (err error) {
	var store *session.Store

	if store, err = p.sessionHolder.Get(ctx); err != nil {
		return err
	}

	err = store.SetExpiration(expiration)
	if err != nil {
		return err
	}

	return p.sessionHolder.Save(ctx, store)
}

// GetExpiration get the expiration of the current session.
func (p *Session) GetExpiration(ctx *fasthttp.RequestCtx) (time.Duration, error) {
	store, err := p.sessionHolder.Get(ctx)
	if err != nil {
		return time.Duration(0), err
	}

	return store.GetExpiration(), nil
}

func NewEncapsulatedSession(base *Session, ctx *fasthttp.RequestCtx) *EncapsulatedSession {
	return &EncapsulatedSession{base: base, ctx: ctx}
}

type EncapsulatedSession struct {
	base *Session

	ctx *fasthttp.RequestCtx
}

// NewDefaultUserSession returns a new default UserSession for this session provider.
func (p *EncapsulatedSession) NewDefaultUserSession() (userSession UserSession) {
	return p.base.NewDefaultUserSession()
}

// GetSession return the user session from a request.
func (p *EncapsulatedSession) GetSession() (userSession UserSession, err error) {
	return p.base.GetSession(p.ctx)
}

// SaveSession save the user session.
func (p *EncapsulatedSession) SaveSession(userSession UserSession) (err error) {
	return p.base.SaveSession(p.ctx, userSession)
}

// RegenerateSession regenerate a session ID.
func (p *EncapsulatedSession) RegenerateSession() (err error) {
	return p.base.RegenerateSession(p.ctx)
}

// DestroySession destroy a session ID and delete the cookie.
func (p *EncapsulatedSession) DestroySession() (err error) {
	return p.base.DestroySession(p.ctx)
}

// UpdateSessionExpiration update the expiration of the cookie and session.
func (p *EncapsulatedSession) UpdateSessionExpiration(expiration time.Duration) (err error) {
	return p.base.UpdateExpiration(p.ctx, expiration)
}

// GetSessionExpiration get the expiration of the current session.
func (p *EncapsulatedSession) GetSessionExpiration() (expiration time.Duration, err error) {
	return p.base.GetExpiration(p.ctx)
}

func (p *EncapsulatedSession) GetSessionConfig() (config schema.SessionCookie) {
	return p.base.Config
}

type Manager interface {
	NewDefaultUserSession() (userSession UserSession)
	GetSession() (userSession UserSession, err error)
	SaveSession(userSession UserSession) (err error)
	RegenerateSession() (err error)
	DestroySession() (err error)
	UpdateSessionExpiration(expiration time.Duration) (err error)
	GetSessionExpiration() (expiration time.Duration, err error)
	GetSessionConfig() (config schema.SessionCookie)
}
