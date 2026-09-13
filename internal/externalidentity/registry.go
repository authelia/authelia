// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Providers is the registry of configured external identity providers.
type Providers struct {
	ordered []Provider
	byID    map[string]Provider
}

// NewProviders returns a new *Providers built from the configuration. No request is made to any of the configured
// providers here; anything a provider has to resolve is resolved the first time it is used.
func NewProviders(config *schema.AuthenticationBackendExternalIdentity, caCertPool *x509.CertPool) (providers *Providers) {
	if config == nil || len(config.Providers) == 0 {
		return nil
	}

	all := make([]Provider, 0, len(config.Providers))

	for i := range config.Providers {
		provider := &config.Providers[i]

		switch provider.Type {
		case ProviderTypeDiscord:
			all = append(all, newDiscordProvider(provider, newProviderClient(caCertPool)))
		case ProviderTypePlex:
			all = append(all, newPlexProvider(provider, newProviderClient(caCertPool)))
		case ProviderTypeGitHub:
			all = append(all, newGitHubProvider(provider, newProviderClient(caCertPool)))
		default:
			all = append(all, newOpenIDConnectProvider(provider, newProviderClient(caCertPool)))
		}
	}

	return NewProvidersWith(all...)
}

// NewProvidersWith returns a new *Providers holding the given providers in the given order.
func NewProvidersWith(all ...Provider) (providers *Providers) {
	providers = &Providers{
		ordered: make([]Provider, 0, len(all)),
		byID:    make(map[string]Provider, len(all)),
	}

	for _, provider := range all {
		providers.ordered = append(providers.ordered, provider)
		providers.byID[provider.ID()] = provider
	}

	return providers
}

// Get returns the provider with the given id.
func (p *Providers) Get(id string) (provider Provider, ok bool) {
	if p == nil {
		return nil, false
	}

	provider, ok = p.byID[id]

	return provider, ok
}

// All returns every configured provider in configuration order.
func (p *Providers) All() (providers []Provider) {
	if p == nil {
		return nil
	}

	return p.ordered
}
