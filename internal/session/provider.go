package session

import (
	"crypto/x509"
	"fmt"

	"github.com/fasthttp/session/v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// Provider contains a list of domain sessions.
type Provider struct {
	sessions map[string]*Session
}

// NewProvider instantiate a session provider given a configuration.
func NewProvider(config schema.Session, certPool *x509.CertPool) *Provider {
	log := logging.Logger()

	name, p, s, err := NewSessionProvider(config, certPool)
	if err != nil {
		log.Fatal(err)
	}

	provider := &Provider{
		sessions: map[string]*Session{},
	}

	var (
		holder *session.Session
	)

	for _, dconfig := range config.Cookies {
		if _, holder, err = NewProviderConfigAndSession(dconfig, name, s, p); err != nil {
			log.Fatal(err)
		}

		provider.sessions[dconfig.Domain] = &Session{
			Config:        dconfig,
			sessionHolder: holder,
		}
	}

	return provider
}

// Get returns session information for specified domain.
func (p *Provider) Get(domain string) (*Session, error) {
	if domain == "" {
		return nil, fmt.Errorf("can not get session from an undefined domain")
	}

	s, found := p.sessions[domain]

	if !found {
		return nil, fmt.Errorf("no session found for domain '%s'", domain)
	}

	return s, nil
}
