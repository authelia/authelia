package session

import (
	"crypto/x509"
	"fmt"

	fasthttpsession "github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	"github.com/fasthttp/session/v2/providers/redis"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// Provider contains a list of domain sessions.
type Provider struct {
	sessions map[string]*Session
}

// NewProvider instantiate a session provider given a configuration.
func NewProvider(config schema.SessionConfiguration, certPool *x509.CertPool) *Provider {
	c := NewProviderConfig(config, certPool)

	provider := new(Provider)
	provider.sessions = make(map[string]*Session)

	logger := logging.Logger()

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

	// configuring extra root domains.
	for _, domain := range config.Domains {
		c.config.Domain = domain.Domain
		provider.sessions[domain.Domain] = &Session{
			sessionHolder: fasthttpsession.New(c.config),
			RememberMe:    domain.RememberMeDuration,
			Inactivity:    domain.Inactivity,
		}

		if err := provider.sessions[domain.Domain].sessionHolder.SetProvider(providerImpl); err != nil {
			logger.Fatal(err)
		}
	}

	return provider
}

// Get returns session information for specified domain.
func (p *Provider) Get(domain string) (*Session, error) {
	if domain == "" {
		return nil, fmt.Errorf("can not get session from an undefined domain")
	}

	session, found := p.sessions[domain]

	if !found {
		return nil, fmt.Errorf("no session found for domain '%s'", domain)
	}

	return session, nil
}
