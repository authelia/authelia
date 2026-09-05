package oidcrp

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Providers is the registry of configured external OpenID Connect 1.0 Providers.
type Providers struct {
	ordered []*Provider
	byID    map[string]*Provider
}

// NewProviders returns a new *Providers built from the configuration. No request is made to any of the configured
// providers here; the discovery document is resolved lazily the first time a provider is used.
func NewProviders(config *schema.AuthenticationBackendOpenIDConnect, caCertPool *x509.CertPool) (providers *Providers) {
	if config == nil || len(config.Providers) == 0 {
		return nil
	}

	providers = &Providers{
		ordered: make([]*Provider, 0, len(config.Providers)),
		byID:    make(map[string]*Provider, len(config.Providers)),
	}

	for i := range config.Providers {
		provider := newProvider(&config.Providers[i], caCertPool)

		providers.ordered = append(providers.ordered, provider)
		providers.byID[provider.ID] = provider
	}

	return providers
}

// Get returns the provider with the given id.
func (p *Providers) Get(id string) (provider *Provider, ok bool) {
	if p == nil {
		return nil, false
	}

	provider, ok = p.byID[id]

	return provider, ok
}

// All returns every configured provider in configuration order.
func (p *Providers) All() (providers []*Provider) {
	if p == nil {
		return nil
	}

	return p.ordered
}
