package totp

import (
	"context"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/random"
)

// Context represents the context used by the TOTP provider.
type Context interface {
	context.Context

	GetClock() clock.Provider
	GetRandom() random.Provider
}

// NewContext returns a Context given the provided context, clock, and random providers.
func NewContext(ctx context.Context, clock clock.Provider, random random.Provider) Context {
	return &SimpleContext{Context: ctx, clock: clock, random: random}
}

// SimpleContext is a Context which wraps a [context.Context] along with the clock and random providers.
type SimpleContext struct {
	context.Context

	clock  clock.Provider
	random random.Provider
}

// GetClock returns the clock provider.
func (c *SimpleContext) GetClock() clock.Provider {
	return c.clock
}

// GetRandom returns the random provider.
func (c *SimpleContext) GetRandom() random.Provider {
	return c.random
}
