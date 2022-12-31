package middlewares

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewBridgeBuilder creates a new BridgeBuilder.
func NewBridgeBuilder(config schema.Configuration, providers Providers) *BridgeBuilder {
	return &BridgeBuilder{
		config:            config,
		providers:         providers,
		autheliaCtxValues: make(map[any]func(ctx *AutheliaCtx) any),
	}
}

// WithConfig sets the schema.Configuration used with this BridgeBuilder.
func (b *BridgeBuilder) WithConfig(config schema.Configuration) *BridgeBuilder {
	b.config = config

	return b
}

// WithProviders sets the Providers used with this BridgeBuilder.
func (b *BridgeBuilder) WithProviders(providers Providers) *BridgeBuilder {
	b.providers = providers

	return b
}

// WithPreMiddlewares sets the Middleware's used with this BridgeBuilder which are applied before the actual Bridge.
func (b *BridgeBuilder) WithPreMiddlewares(middlewares ...Middleware) *BridgeBuilder {
	b.preMiddlewares = middlewares

	return b
}

// WithAutheliaCtxValue sets the value associated with key to `AutheliaCtx` which can then be obtained from this context by the key.
// This makes it easy to pass various values initialized by `AutheliaCtx` to providers.
// For example, passing `WriteFormPostResponse` func to `OpenIDConnectProvider` to display the authorization page with custom `form_post` response method.
func (b *BridgeBuilder) WithAutheliaCtxValue(key any, val func(ctx *AutheliaCtx) any) *BridgeBuilder {
	b.autheliaCtxValues[key] = val

	return b
}

// WithPostMiddlewares sets the AutheliaMiddleware's used with this BridgeBuilder which are applied after the actual
// Bridge.
func (b *BridgeBuilder) WithPostMiddlewares(middlewares ...AutheliaMiddleware) *BridgeBuilder {
	b.postMiddlewares = middlewares

	return b
}

// Build and return the Bridge configured by this BridgeBuilder.
func (b *BridgeBuilder) Build() Bridge {
	return func(next RequestHandler) fasthttp.RequestHandler {
		for i := len(b.postMiddlewares) - 1; i >= 0; i-- {
			next = b.postMiddlewares[i](next)
		}

		bridge := func(requestCtx *fasthttp.RequestCtx) {
			ctx := NewAutheliaCtx(requestCtx, b.config, b.providers)
			for key, val := range b.autheliaCtxValues {
				ctx = AutheliaCtxWithValue(ctx, key, val(ctx))
			}

			next(ctx)
		}

		for i := len(b.preMiddlewares) - 1; i >= 0; i-- {
			bridge = b.preMiddlewares[i](bridge)
		}

		return bridge
	}
}
