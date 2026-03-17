package session

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

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

type DefaultProvider struct {
	codec      Codec
	strategies map[string]Strategy
}

func (d DefaultProvider) StartupCheck() (err error) {
	return nil
}

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
