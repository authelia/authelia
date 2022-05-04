package middlewares

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewBridgeBuilder creates a new BridgeBuilder.
func NewBridgeBuilder(config schema.Configuration, providers Providers) *BridgeBuilder {
	return &BridgeBuilder{
		config:    config,
		providers: providers,
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

// WithMiddlewares sets the Middleware's used with this BridgeBuilder.
func (b *BridgeBuilder) WithMiddlewares(middlewares ...Middleware) *BridgeBuilder {
	b.middlewares = middlewares

	return b
}

// WithAutheliaMiddlewares sets the AutheliaMiddleware's used with this BridgeBuilder.
func (b *BridgeBuilder) WithAutheliaMiddlewares(middlewares ...AutheliaMiddleware) *BridgeBuilder {
	b.autheliaMiddlewares = middlewares

	return b
}

// Build and return the Bridge configured by this BridgeBuilder.
func (b *BridgeBuilder) Build() Bridge {
	return func(next RequestHandler) fasthttp.RequestHandler {
		bridge := func(ctx *fasthttp.RequestCtx) {
			autheliaCtx, err := NewAutheliaCtx(ctx, b.config, b.providers)
			if err != nil {
				autheliaCtx.Error(err, messageOperationFailed)
				return
			}

			for i := len(b.autheliaMiddlewares) - 1; i >= 0; i-- {
				next = b.autheliaMiddlewares[i](next)
			}

			next(autheliaCtx)
		}

		for i := len(b.middlewares) - 1; i >= 0; i-- {
			bridge = b.middlewares[i](bridge)
		}

		return bridge
	}
}
