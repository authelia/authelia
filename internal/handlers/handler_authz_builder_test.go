package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestAuthzBuilder_WithConfig(t *testing.T) {
	builder := NewAuthzBuilder()

	builder.WithConfig(&schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			RefreshInterval: schema.NewRefreshIntervalDurationAlways(),
		},
	})

	assert.Equal(t, schema.NewRefreshIntervalDurationAlways(), builder.config.RefreshInterval)

	builder.WithConfig(&schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			RefreshInterval: schema.NewRefreshIntervalDurationNever(),
		},
	})

	assert.Equal(t, schema.NewRefreshIntervalDurationNever(), builder.config.RefreshInterval)

	builder.WithConfig(&schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			RefreshInterval: schema.NewRefreshIntervalDuration(time.Minute),
		},
	})

	assert.Equal(t, schema.NewRefreshIntervalDuration(time.Minute), builder.config.RefreshInterval)

	builder.WithConfig(nil)

	assert.Equal(t, schema.NewRefreshIntervalDuration(time.Minute), builder.config.RefreshInterval)
}

func TestAuthzBuilder_WithEndpointConfig(t *testing.T) {
	builder := NewAuthzBuilder()

	builder.WithEndpointConfig(schema.ServerEndpointsAuthz{
		Implementation: "ExtAuthz",
	})

	assert.Equal(t, AuthzImplExtAuthz, builder.implementation)

	builder.WithEndpointConfig(schema.ServerEndpointsAuthz{
		Implementation: "ForwardAuth",
	})

	assert.Equal(t, AuthzImplForwardAuth, builder.implementation)

	builder.WithEndpointConfig(schema.ServerEndpointsAuthz{
		Implementation: "AuthRequest",
	})

	assert.Equal(t, AuthzImplAuthRequest, builder.implementation)

	builder.WithEndpointConfig(schema.ServerEndpointsAuthz{
		Implementation: "Legacy",
	})

	assert.Equal(t, AuthzImplLegacy, builder.implementation)

	builder.WithEndpointConfig(schema.ServerEndpointsAuthz{
		Implementation: "ExtAuthz",
		AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{
			{Name: "HeaderProxyAuthorization"},
			{Name: "CookieSession"},
		},
	})

	assert.Len(t, builder.strategies, 2)

	builder.WithEndpointConfig(schema.ServerEndpointsAuthz{
		Implementation: "ExtAuthz",
		AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{
			{Name: "HeaderAuthorization"},
			{Name: "HeaderProxyAuthorization"},
			{Name: "HeaderAuthRequestProxyAuthorization"},
			{Name: "HeaderLegacy"},
			{Name: "CookieSession"},
		},
	})

	assert.Len(t, builder.strategies, 5)
}
