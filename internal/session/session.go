package session

import (
	"encoding/json"
	"time"

	fasthttpsession "github.com/fasthttp/session/v2"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Session a session provider.
type Session struct {
	Config schema.SessionCookieConfiguration

	sessionHolder *fasthttpsession.Session
}

// GetSession return the user session from a request.
func (p *Session) GetSession(ctx *fasthttp.RequestCtx) (UserSession, error) {
	store, err := p.sessionHolder.Get(ctx)

	if err != nil {
		return NewDefaultUserSession(), err
	}

	userSessionJSON, ok := store.Get(userSessionStorerKey).([]byte)

	// If userSession is not yet defined we create the new session with default values
	// and save it in the store.
	if !ok {
		userSession := NewDefaultUserSession()

		store.Set(userSessionStorerKey, userSession)

		return userSession, nil
	}

	var userSession UserSession
	err = json.Unmarshal(userSessionJSON, &userSession)

	if err != nil {
		return NewDefaultUserSession(), err
	}

	return userSession, nil
}

// SaveSession save the user session.
func (p *Session) SaveSession(ctx *fasthttp.RequestCtx, userSession UserSession) error {
	store, err := p.sessionHolder.Get(ctx)

	if err != nil {
		return err
	}

	userSessionJSON, err := json.Marshal(userSession)

	if err != nil {
		return err
	}

	store.Set(userSessionStorerKey, userSessionJSON)

	err = p.sessionHolder.Save(ctx, store)

	if err != nil {
		return err
	}

	return nil
}

// RegenerateSession regenerate a session ID.
func (p *Session) RegenerateSession(ctx *fasthttp.RequestCtx) error {
	err := p.sessionHolder.Regenerate(ctx)

	return err
}

// DestroySession destroy a session ID and delete the cookie.
func (p *Session) DestroySession(ctx *fasthttp.RequestCtx) error {
	return p.sessionHolder.Destroy(ctx)
}

// UpdateExpiration update the expiration of the cookie and session.
func (p *Session) UpdateExpiration(ctx *fasthttp.RequestCtx, expiration time.Duration) error {
	store, err := p.sessionHolder.Get(ctx)

	if err != nil {
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
