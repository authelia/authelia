package session

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

// NewProvider returns a new Provider with a Strategy for every configured session cookie domain, each sharing the
// Codec derived from the session secret and HMAC key.
func NewProvider(config *schema.Configuration, hmac []byte, clock clock.Provider, random random.Provider, repository Repository) (provider Provider, err error) {
	var (
		codec Codec
	)

	if codec, err = NewCodec(config.Session.Secret, hmac, random); err != nil {
		return nil, err
	}

	strategies := make(map[string]Strategy)

	for _, cookie := range config.Session.Cookies {
		strategies[cookie.Domain] = NewStrategy(cookie, clock, codec, repository)
	}

	return &DefaultProvider{codec: codec, strategies: strategies}, nil
}

// DefaultProvider is the default Provider which resolves a Strategy from the cookie domain of a request.
type DefaultProvider struct {
	codec      Codec
	strategies map[string]Strategy
}

// StartupCheck implements the Provider interface.
func (d DefaultProvider) StartupCheck() (err error) {
	return nil
}

// GetStrategy returns the Strategy configured for the given cookie domain, returning an error when the domain has no
// configured session cookie.
func (d DefaultProvider) GetStrategy(domain string) (strategy Strategy, err error) {
	var ok bool

	if strategy, ok = d.strategies[domain]; ok {
		return strategy, nil
	}

	return nil, fmt.Errorf("not found")
}

var (
	_ Provider = (*DefaultProvider)(nil)
)
