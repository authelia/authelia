package session

import (
	"crypto/x509"
	"encoding/json"
	"time"

	"github.com/fasthttp/session/v2"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/utils"
)

// Provider a session provider.
type Provider struct {
	manager *session.Session
	config  *session.Config

	store Store

	RememberMe time.Duration
	Inactivity time.Duration
}

// NewProvider instantiate a session provider given a configuration.
func NewProvider(configuration schema.SessionConfiguration, certPool *x509.CertPool) (provider *Provider) {
	provider = new(Provider)

	provider.config = NewSessionConfig(configuration)
	provider.manager = session.New(*provider.config)

	logger := logging.Logger()

	duration, err := utils.ParseDurationString(configuration.RememberMeDuration)
	if err != nil {
		logger.Fatal(err)
	}

	provider.RememberMe = duration

	duration, err = utils.ParseDurationString(configuration.Inactivity)
	if err != nil {
		logger.Fatal(err)
	}

	provider.Inactivity = duration

	switch {
	case configuration.Redis != nil && configuration.Redis.HighAvailability != nil:
		provider.store = NewRedisFailoverStore(configuration.Redis, certPool, logger)
	case configuration.Redis != nil:
		provider.store = NewRedisStandaloneStore(configuration.Redis, certPool, logger)
	default:
		provider.store = NewMemoryStore()
	}

	err = provider.manager.SetProvider(provider.store)
	if err != nil {
		logger.Fatal(err)
	}

	return provider
}

// GetSession return the user session from a request.
func (p *Provider) GetSession(ctx *fasthttp.RequestCtx) (UserSession, error) {
	store, err := p.manager.Get(ctx)

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
	store, err := p.manager.Get(ctx)

	if err != nil {
		return err
	}

	userSessionJSON, err := json.Marshal(userSession)

	if err != nil {
		return err
	}

	store.Set(userSessionStorerKey, userSessionJSON)

	err = p.manager.Save(ctx, store)

	if err != nil {
		return err
	}

	return nil
}

// RegenerateSession regenerate a session ID.
func (p *Provider) RegenerateSession(ctx *fasthttp.RequestCtx) error {
	err := p.manager.Regenerate(ctx)

	return err
}

// DestroySession destroy a session ID and delete the cookie.
func (p *Provider) DestroySession(ctx *fasthttp.RequestCtx) error {
	return p.manager.Destroy(ctx)
}

// UpdateExpiration update the expiration of the cookie and session.
func (p *Provider) UpdateExpiration(ctx *fasthttp.RequestCtx, expiration time.Duration) error {
	store, err := p.manager.Get(ctx)

	if err != nil {
		return err
	}

	err = store.SetExpiration(expiration)

	if err != nil {
		return err
	}

	return p.manager.Save(ctx, store)
}

// GetExpiration get the expiration of the current session.
func (p *Provider) GetExpiration(ctx *fasthttp.RequestCtx) (time.Duration, error) {
	store, err := p.manager.Get(ctx)

	if err != nil {
		return time.Duration(0), err
	}

	return store.GetExpiration(), nil
}
