package middlewares

import (
	"github.com/go-spop/spop/request"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewBridgeSPOPBuilder creates a new BridgeSPOPBuilder.
func NewBridgeSPOPBuilder(config schema.Configuration, providers Providers) *BridgeSPOPBuilder {
	return &BridgeSPOPBuilder{
		config:    config,
		providers: providers,
	}
}

// WithConfig sets the schema.Configuration used with this BridgeSPOPBuilder.
func (b *BridgeSPOPBuilder) WithConfig(config schema.Configuration) *BridgeSPOPBuilder {
	b.config = config

	return b
}

// WithProviders sets the Providers used with this BridgeSPOPBuilder.
func (b *BridgeSPOPBuilder) WithProviders(providers Providers) *BridgeSPOPBuilder {
	b.providers = providers

	return b
}

// Build and return the Bridge configured by this BridgeSPOPBuilder.
func (b *BridgeSPOPBuilder) Build() BridgeSPOP {
	return func(next RequestSPOPHandler) func(request *request.Request) {
		bridge := func(request *request.Request) {
			message, err := request.Messages.GetByName("authelia-authz")
			if err != nil {
				// TODO: Log Error.
				return
			} else if message == nil {
				// TODO: Log Error.
				return
			}

			next(NewAutheliaSPOPCtx(request, message, b.config, b.providers))
		}

		return bridge
	}
}
