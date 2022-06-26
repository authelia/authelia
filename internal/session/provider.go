package session

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"time"

	fasthttpsession "github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// Provider a session provider.
type Provider struct {
	sessionHolder map[string]*fasthttpsession.Session
	RememberMe    time.Duration
	Inactivity    time.Duration
}

// NewProvider instantiate a session provider given a configuration.
func NewProvider(config schema.SessionConfiguration, certPool *x509.CertPool) *Provider {
	c := NewProviderConfig(config, certPool)

	provider := new(Provider)
	provider.sessionHolder = make(map[string]*fasthttpsession.Session)

	// configure default root domain.
	if config.Domain != "" {
		provider.sessionHolder[config.Domain] = fasthttpsession.New(c.config)
	}
	// configuring extra root domains.

	for _, domain := range config.DomainList {
		c.config.Domain = domain
		provider.sessionHolder[domain] = fasthttpsession.New(c.config)
	}

	logger := logging.Logger()

	provider.Inactivity, provider.RememberMe = config.Inactivity, config.RememberMeDuration

	var (
		providerImpl fasthttpsession.Provider
		err          error
	)

	switch {
	case c.redisConfig != nil:
		providerImpl, err = redis.New(*c.redisConfig)
		if err != nil {
			logger.Fatal(err)
		}
	case c.redisSentinelConfig != nil:
		providerImpl, err = redis.NewFailoverCluster(*c.redisSentinelConfig)
		if err != nil {
			logger.Fatal(err)
		}
	default:
		providerImpl, err = memory.New(memory.Config{})
		if err != nil {
			logger.Fatal(err)
		}
	}

	for domain := range provider.sessionHolder {
		err = provider.sessionHolder[domain].SetProvider(providerImpl)
		if err != nil {
			logger.Fatal(err)
		}
	}

	return provider
}

// GetSession return the user session from a request.
func (p *Provider) GetSession(ctx *fasthttp.RequestCtx, domain string) (UserSession, error) {
	sessionHolder, found := p.sessionHolder[domain]
	if !found {
		return NewDefaultUserSession(), fmt.Errorf("no session for domain %s found", domain)
	}

	store, err := sessionHolder.Get(ctx)

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
func (p *Provider) SaveSession(ctx *fasthttp.RequestCtx, userSession UserSession, domain string) error {
	sessionHolder, found := p.sessionHolder[domain]
	if !found {
		return fmt.Errorf("no session for domain %s found", domain)
	}

	store, err := sessionHolder.Get(ctx)

	if err != nil {
		return err
	}

	userSessionJSON, err := json.Marshal(userSession)

	if err != nil {
		return err
	}

	store.Set(userSessionStorerKey, userSessionJSON)

	err = sessionHolder.Save(ctx, store)

	if err != nil {
		return err
	}

	return nil
}

// RegenerateSession regenerate a session ID.
func (p *Provider) RegenerateSession(ctx *fasthttp.RequestCtx, domain string) error {
	sessionHolder, found := p.sessionHolder[domain]
	if !found {
		return fmt.Errorf("no session for domain %s found", domain)
	}

	err := sessionHolder.Regenerate(ctx)

	return err
}

// DestroySession destroy a session ID and delete the cookie.
func (p *Provider) DestroySession(ctx *fasthttp.RequestCtx, domain string) error {
	sessionHolder, found := p.sessionHolder[domain]
	if !found {
		return fmt.Errorf("no session for domain %s found", domain)
	}

	return sessionHolder.Destroy(ctx)
}

// UpdateExpiration update the expiration of the cookie and session.
func (p *Provider) UpdateExpiration(ctx *fasthttp.RequestCtx, expiration time.Duration, domain string) error {
	sessionHolder, found := p.sessionHolder[domain]
	if !found {
		return fmt.Errorf("no session for domain %s found", domain)
	}

	store, err := sessionHolder.Get(ctx)

	if err != nil {
		return err
	}

	err = store.SetExpiration(expiration)

	if err != nil {
		return err
	}

	return sessionHolder.Save(ctx, store)
}

// GetExpiration get the expiration of the current session.
func (p *Provider) GetExpiration(ctx *fasthttp.RequestCtx, domain string) (time.Duration, error) {
	sessionHolder, found := p.sessionHolder[domain]
	if !found {
		return time.Duration(0), fmt.Errorf("no session for domain %s found", domain)
	}

	store, err := sessionHolder.Get(ctx)

	if err != nil {
		return time.Duration(0), err
	}

	return store.GetExpiration(), nil
}
