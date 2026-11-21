// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/url"
	"time"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/storage"
)

// DetachedContext is a Context which isn't bound to a request, with an issuer and providers given up front. The
// providers it returns are safe to use concurrently. It's used where there is no request, such as delivering the Logout
// Tokens of OpenID Connect Back-Channel Logout 1.0, which use the issuer recorded when the End-User logged out.
type DetachedContext struct {
	context.Context

	issuer   *url.URL
	clock    clock.Provider
	random   random.Provider
	config   *schema.Configuration
	storage  storage.Provider
	users    authentication.UserProvider
	resolver expression.UserAttributeResolver
}

// DetachedContextProviders are the providers a DetachedContext returns.
type DetachedContextProviders struct {
	Clock    clock.Provider
	Random   random.Provider
	Config   *schema.Configuration
	Storage  storage.Provider
	Users    authentication.UserProvider
	Resolver expression.UserAttributeResolver
}

// NewDetachedContextProviders returns the DetachedContextProviders of the Context.
func NewDetachedContextProviders(ctx Context) DetachedContextProviders {
	return DetachedContextProviders{
		Clock:    ctx.GetClock(),
		Random:   ctx.GetRandom(),
		Config:   ctx.GetConfiguration(),
		Storage:  ctx.GetProviderStorage(),
		Users:    ctx.GetUserProvider(),
		Resolver: ctx.GetProviderUserAttributeResolver(),
	}
}

// NewDetachedContextFromProviders returns a DetachedContext for the issuer and providers which is cancelled after the
// timeout or when the parent is done, along with its cancellation function. It's used where there is no request, such
// as delivering the Logout Tokens of OpenID Connect Back-Channel Logout 1.0 which were not delivered when the End-User
// logged out, which use the issuer recorded when the End-User logged out.
func NewDetachedContextFromProviders(parent context.Context, timeout time.Duration, issuer *url.URL, providers DetachedContextProviders) (detached *DetachedContext, cancel context.CancelFunc) {
	c, cancel := context.WithTimeout(parent, timeout)

	return &DetachedContext{
		Context:  c,
		issuer:   issuer,
		clock:    providers.Clock,
		random:   providers.Random,
		config:   providers.Config,
		storage:  providers.Storage,
		users:    providers.Users,
		resolver: providers.Resolver,
	}, cancel
}

// IssuerURL returns the issuer resolved when this context was created.
func (ctx *DetachedContext) IssuerURL() (issuerURL *url.URL, err error) {
	return ctx.issuer, nil
}

// Value returns this context for the Authelia context key, which is how a Config resolves the Context of a request,
// and otherwise defers to the context this one is bound by.
func (ctx *DetachedContext) Value(key any) (value any) {
	if key == model.CtxKeyAutheliaCtx {
		return ctx
	}

	return ctx.Context.Value(key)
}

// GetClock returns the clock provider.
func (ctx *DetachedContext) GetClock() (provider clock.Provider) {
	return ctx.clock
}

// GetRandom returns the random provider.
func (ctx *DetachedContext) GetRandom() (provider random.Provider) {
	return ctx.random
}

// GetConfiguration returns the configuration.
func (ctx *DetachedContext) GetConfiguration() (config *schema.Configuration) {
	return ctx.config
}

// GetProviderStorage returns the storage provider.
func (ctx *DetachedContext) GetProviderStorage() (provider storage.Provider) {
	return ctx.storage
}

// GetUserProvider returns the user provider.
func (ctx *DetachedContext) GetUserProvider() (provider authentication.UserProvider) {
	return ctx.users
}

// GetProviderUserAttributeResolver returns the user attribute resolver.
func (ctx *DetachedContext) GetProviderUserAttributeResolver() (resolver expression.UserAttributeResolver) {
	return ctx.resolver
}

var (
	_ Context = (*DetachedContext)(nil)
)
