package session

import (
	"encoding/json"
	"time"

	"github.com/clems4ever/authelia/configuration/schema"
	fasthttpsession "github.com/fasthttp/session"
	"github.com/valyala/fasthttp"
)

// Provider a session provider.
type Provider struct {
	sessionHolder *fasthttpsession.Session
}

// NewProvider instantiate a session provider given a configuration.
func NewProvider(configuration schema.SessionConfiguration) *Provider {
	providerConfig := NewProviderConfig(configuration)

	provider := new(Provider)
	provider.sessionHolder = fasthttpsession.New(providerConfig.config)
	err := provider.sessionHolder.SetProvider(providerConfig.providerName, providerConfig.providerConfig)
	if err != nil {
		panic(err)
	}
	return provider
}

// GetSession return the user session from a request
func (p *Provider) GetSession(ctx *fasthttp.RequestCtx) (UserSession, error) {
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
func (p *Provider) SaveSession(ctx *fasthttp.RequestCtx, userSession UserSession) error {
	store, err := p.sessionHolder.Get(ctx)

	if err != nil {
		return err
	}

	userSessionJSON, err := json.Marshal(userSession)

	if err != nil {
		return err
	}

	store.Set(userSessionStorerKey, userSessionJSON)
	p.sessionHolder.Save(ctx, store)
	return nil
}

// RegenerateSession regenerate a session ID.
func (p *Provider) RegenerateSession(ctx *fasthttp.RequestCtx) error {
	_, err := p.sessionHolder.Regenerate(ctx)
	return err
}

// DestroySession destroy a session ID and delete the cookie.
func (p *Provider) DestroySession(ctx *fasthttp.RequestCtx) error {
	return p.sessionHolder.Destroy(ctx)
}

// UpdateExpiration update the expiration of the cookie and session.
func (p *Provider) UpdateExpiration(ctx *fasthttp.RequestCtx, expiration time.Duration) error {
	store, err := p.sessionHolder.Get(ctx)

	if err != nil {
		return err
	}

	err = store.SetExpiration(expiration)

	if err != nil {
		return err
	}

	p.sessionHolder.Save(ctx, store)
	return nil
}

// GetExpiration get the expiration of the current session.
func (p *Provider) GetExpiration(ctx *fasthttp.RequestCtx) (time.Duration, error) {
	store, err := p.sessionHolder.Get(ctx)

	if err != nil {
		return time.Duration(0), err
	}

	return store.GetExpiration(), nil
}
