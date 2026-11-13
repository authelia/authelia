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

// DetachedContext is a Context which answers for a request without holding it. The issuer is resolved up front, on the
// goroutine handling the request, as resolving it reads the request headers which are not safe for concurrent use. The
// providers it returns are safe to use concurrently. It's used where work outgrows the request, such as delivering the
// Logout Tokens of OpenID Connect Back-Channel Logout 1.0 to several Relying Parties at once.
type DetachedContext struct {
	context.Context

	ctx    Context
	issuer *url.URL
}

// NewDetachedContext returns a DetachedContext for the given Context which is cancelled after the timeout, along with
// its cancellation function. It must be called on the goroutine which handles the request.
func NewDetachedContext(ctx Context, timeout time.Duration) (detached *DetachedContext, cancel context.CancelFunc, err error) {
	var issuer *url.URL

	if issuer, err = ctx.IssuerURL(); err != nil {
		return nil, nil, err
	}

	// The parent is the background context rather than the request, whose cancellation fasthttp only signals on server
	// shutdown and whose values are answered by this context instead.
	c, cancel := context.WithTimeout(context.Background(), timeout)

	return &DetachedContext{Context: c, ctx: ctx, issuer: issuer}, cancel, nil
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
	return ctx.ctx.GetClock()
}

// GetRandom returns the random provider.
func (ctx *DetachedContext) GetRandom() (provider random.Provider) {
	return ctx.ctx.GetRandom()
}

// GetConfiguration returns the configuration.
func (ctx *DetachedContext) GetConfiguration() (config *schema.Configuration) {
	return ctx.ctx.GetConfiguration()
}

// GetProviderStorage returns the storage provider.
func (ctx *DetachedContext) GetProviderStorage() (provider storage.Provider) {
	return ctx.ctx.GetProviderStorage()
}

// GetUserProvider returns the user provider.
func (ctx *DetachedContext) GetUserProvider() (provider authentication.UserProvider) {
	return ctx.ctx.GetUserProvider()
}

// GetProviderUserAttributeResolver returns the user attribute resolver.
func (ctx *DetachedContext) GetProviderUserAttributeResolver() (resolver expression.UserAttributeResolver) {
	return ctx.ctx.GetProviderUserAttributeResolver()
}

var (
	_ Context = (*DetachedContext)(nil)
)
