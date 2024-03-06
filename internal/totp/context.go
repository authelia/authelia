package totp

import (
	"context"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/random"
)

type Context interface {
	context.Context

	GetClock() clock.Provider
	GetRandom() random.Provider
}

func NewContext(ctx context.Context, clock clock.Provider, random random.Provider) Context {
	return &SimpleContext{Context: ctx, clock: clock, random: random}
}

type SimpleContext struct {
	context.Context

	clock  clock.Provider
	random random.Provider
}

func (c *SimpleContext) GetClock() clock.Provider {
	return c.clock
}

func (c *SimpleContext) GetRandom() random.Provider {
	return c.random
}
