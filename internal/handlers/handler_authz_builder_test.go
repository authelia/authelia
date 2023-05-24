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
			RefreshInterval: "always",
		},
	})

	assert.Equal(t, time.Second*0, builder.config.RefreshInterval)

	builder.WithConfig(&schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			RefreshInterval: "disable",
		},
	})

	assert.Equal(t, time.Second*-1, builder.config.RefreshInterval)

	builder.WithConfig(&schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			RefreshInterval: "1m",
		},
	})

	assert.Equal(t, time.Minute, builder.config.RefreshInterval)

	builder.WithConfig(nil)

	assert.Equal(t, time.Minute, builder.config.RefreshInterval)
}

func TestAuthzBuilder_WithEndpointConfig(t *testing.T) {
	builder := NewAuthzBuilder()

	builder.WithEndpointConfig(schema.ServerAuthzEndpoint{
		Implementation: "ExtAuthz",
	})

	assert.Equal(t, AuthzImplExtAuthz, builder.implementation)

	builder.WithEndpointConfig(schema.ServerAuthzEndpoint{
		Implementation: "ForwardAuth",
	})

	assert.Equal(t, AuthzImplForwardAuth, builder.implementation)

	builder.WithEndpointConfig(schema.ServerAuthzEndpoint{
		Implementation: "AuthRequest",
	})

	assert.Equal(t, AuthzImplAuthRequest, builder.implementation)

	builder.WithEndpointConfig(schema.ServerAuthzEndpoint{
		Implementation: "Legacy",
	})

	assert.Equal(t, AuthzImplLegacy, builder.implementation)

	builder.WithEndpointConfig(schema.ServerAuthzEndpoint{
		Implementation: "ExtAuthz",
		AuthnStrategies: []schema.ServerAuthzEndpointAuthnStrategy{
			{Name: "HeaderProxyAuthorization"},
			{Name: "CookieSession"},
		},
	})

	assert.Len(t, builder.strategies, 2)

	builder.WithEndpointConfig(schema.ServerAuthzEndpoint{
		Implementation: "ExtAuthz",
		AuthnStrategies: []schema.ServerAuthzEndpointAuthnStrategy{
			{Name: "HeaderAuthorization"},
			{Name: "HeaderProxyAuthorization"},
			{Name: "HeaderAuthRequestProxyAuthorization"},
			{Name: "HeaderLegacy"},
			{Name: "CookieSession"},
		},
	})

	assert.Len(t, builder.strategies, 5)
}
