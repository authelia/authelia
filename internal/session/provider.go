package session

import (
	"crypto/x509"
	"fmt"
	"time"

	"github.com/fasthttp/session/v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Provider contains a list of domain sessions.
type Provider struct {
	sessions    map[string]*Session
	backend     session.Provider
	backendName string
	errStartup  error
}

// NewProvider instantiate a session provider given a configuration.
func NewProvider(config schema.Session, certPool *x509.CertPool) *Provider {
	name, p, s, err := NewSessionProvider(config, certPool)
	if err != nil {
		return &Provider{errStartup: fmt.Errorf("error initializing session backend: %w", err)}
	}

	provider := &Provider{
		sessions:    map[string]*Session{},
		backend:     p,
		backendName: name,
	}

	var (
		holder *session.Session
	)

	for _, dconfig := range config.Cookies {
		if _, holder, err = NewProviderConfigAndSession(dconfig, name, s, p); err != nil {
			provider.errStartup = fmt.Errorf("error initializing session for domain '%s': %w", dconfig.Domain, err)

			return provider
		}

		provider.sessions[dconfig.Domain] = &Session{
			Config:        dconfig,
			sessionHolder: holder,
		}
	}

	return provider
}

// StartupCheck implements the provider startup check interface.
func (p *Provider) StartupCheck() (err error) {
	if p.errStartup != nil {
		return p.errStartup
	}

	id := []byte("authelia-startup-probe")
	data := []byte("ok")

	for i := 0; i < 19; i++ {
		if err = p.backend.Save(id, data, time.Second); err == nil {
			break
		}

		time.Sleep(time.Millisecond * 500)
	}

	if err != nil {
		return fmt.Errorf("error writing startup probe to session backend '%s': %w", p.backendName, err)
	}

	if err = p.backend.Destroy(id); err != nil {
		return fmt.Errorf("error destroying startup probe in session backend '%s': %w", p.backendName, err)
	}

	return nil
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
