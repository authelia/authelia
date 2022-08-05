package handlers

import (
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewAuthzBuilder creates a new AuthzBuilder.
func NewAuthzBuilder() *AuthzBuilder {
	return &AuthzBuilder{
		config: AuthzConfig{RefreshInterval: time.Second * -1},
	}
}

// WithStrategies replaces all strategies in this builder with the provided value.
func (b *AuthzBuilder) WithStrategies(strategies ...AuthnStrategy) *AuthzBuilder {
	b.strategies = strategies

	return b
}

// WithStrategyCookie adds the Cookie header strategy to the strategies in this builder.
func (b *AuthzBuilder) WithStrategyCookie(refreshInterval time.Duration) *AuthzBuilder {
	b.strategies = append(b.strategies, NewCookieSessionAuthnStrategy(refreshInterval))

	return b
}

// WithStrategyAuthorization adds the Authorization header strategy to the strategies in this builder.
func (b *AuthzBuilder) WithStrategyAuthorization() *AuthzBuilder {
	b.strategies = append(b.strategies, NewHeaderAuthorizationAuthnStrategy())

	return b
}

// WithStrategyProxyAuthorization adds the Proxy-Authorization header strategy to the strategies in this builder.
func (b *AuthzBuilder) WithStrategyProxyAuthorization() *AuthzBuilder {
	b.strategies = append(b.strategies, NewHeaderProxyAuthorizationAuthnStrategy())

	return b
}

// WithImplementationLegacy configures this builder to output an Authz which is used with the Legacy
// implementation which is a mix of the other implementations and usually works with most proxies.
func (b *AuthzBuilder) WithImplementationLegacy() *AuthzBuilder {
	b.impl = AuthzImplLegacy

	return b
}

// WithImplementationForwardAuth configures this builder to output an Authz which is used with the ForwardAuth
// implementation traditionally used by Traefik, Caddy, and Skipper.
func (b *AuthzBuilder) WithImplementationForwardAuth() *AuthzBuilder {
	b.impl = AuthzImplForwardAuth

	return b
}

// WithImplementationAuthRequest configures this builder to output an Authz which is used with the AuthRequest
// implementation traditionally used by NGINX.
func (b *AuthzBuilder) WithImplementationAuthRequest() *AuthzBuilder {
	b.impl = AuthzImplAuthRequest

	return b
}

// WithImplementationExtAuthz configures this builder to output an Authz which is used with the ExtAuthz
// implementation traditionally used by Envoy.
func (b *AuthzBuilder) WithImplementationExtAuthz() *AuthzBuilder {
	b.impl = AuthzImplExtAuthz

	return b
}

// WithConfig allows configuring the Authz config by providing a *schema.Configuration. This function converts it to
// an AuthzConfig and assigns it to the builder.
func (b *AuthzBuilder) WithConfig(config *schema.Configuration) *AuthzBuilder {
	if config == nil {
		return b
	}

	var refreshInterval time.Duration

	switch config.AuthenticationBackend.RefreshInterval {
	case schema.ProfileRefreshDisabled:
		refreshInterval = time.Second * -1
	case schema.ProfileRefreshAlways:
		refreshInterval = time.Second * 0
	default:
		refreshInterval, _ = utils.ParseDurationString(config.AuthenticationBackend.RefreshInterval)
	}

	b.config = AuthzConfig{
		RefreshInterval: refreshInterval,
		Domains: []AuthzDomain{
			{
				Name:      fmt.Sprintf(".%s", config.Session.Domain),
				PortalURL: nil,
			},
		},
	}

	return b
}

// WithEndpointConfig configures the AuthzBuilder with a *schema.ServerAuthzEndpointConfig. Should be called AFTER
// WithConfig or WithAuthzConfig.
func (b *AuthzBuilder) WithEndpointConfig(config schema.ServerAuthzEndpointConfig) *AuthzBuilder {
	switch config.Implementation {
	case AuthzImplForwardAuth.String():
		b.WithImplementationForwardAuth()
	case AuthzImplAuthRequest.String():
		b.WithImplementationAuthRequest()
	default:
		b.WithImplementationLegacy()
	}

	b.WithStrategies()

	for _, strategy := range config.AuthnStrategies {
		switch strategy.Name {
		case AuthnStrategyCookieSession:
			b.strategies = append(b.strategies, NewCookieSessionAuthnStrategy(b.config.RefreshInterval))
		case AuthnStrategyHeaderAuthorization:
			b.strategies = append(b.strategies, NewHeaderAuthorizationAuthnStrategy())
		case AuthnStrategyHeaderProxyAuthorization:
			b.strategies = append(b.strategies, NewHeaderProxyAuthorizationAuthnStrategy())
		case AuthnStrategyHeaderAuthRequestProxyAuthorization:
			b.strategies = append(b.strategies, NewHeaderAuthRequestProxyAuthorizationAuthnStrategy())
		case AuthnStrategyHeaderLegacy:
			b.strategies = append(b.strategies, NewHeaderLegacyAuthnStrategy())
		}
	}

	return b
}

// WithAuthzConfig allows configuring the Authz config by providing a AuthzConfig directly. Recommended this is only
// used in testing and WithConfig is used instead.
func (b *AuthzBuilder) WithAuthzConfig(config AuthzConfig) *AuthzBuilder {
	b.config = config

	return b
}

// Build returns a new Authz from the currently configured options in this builder.
func (b *AuthzBuilder) Build() (authz *Authz) {
	authz = &Authz{
		config:            b.config,
		strategies:        b.strategies,
		fObjectVerify:     authzObjectVerifyStandard,
		fHandleAuthorized: authzHandleAuthorizedStandard,
	}

	if len(authz.strategies) == 0 {
		switch b.impl {
		case AuthzImplLegacy:
			authz.strategies = []AuthnStrategy{NewHeaderLegacyAuthnStrategy(), NewCookieSessionAuthnStrategy(b.config.RefreshInterval)}
		case AuthzImplAuthRequest:
			authz.strategies = []AuthnStrategy{NewHeaderAuthRequestProxyAuthorizationAuthnStrategy(), NewCookieSessionAuthnStrategy(b.config.RefreshInterval)}
		default:
			authz.strategies = []AuthnStrategy{NewHeaderProxyAuthorizationAuthnStrategy(), NewCookieSessionAuthnStrategy(b.config.RefreshInterval)}
		}
	}

	switch b.impl {
	case AuthzImplLegacy:
		authz.fObjectGet = authzGetObjectImplLegacy
		authz.fHandleUnauthorized = authzHandleUnauthorizedImplLegacy
		authz.fPortalURL = authzPortalURLFromQuery
	case AuthzImplForwardAuth:
		authz.fObjectGet = authzGetObjectImplForwardAuth
		authz.fHandleUnauthorized = authzHandleUnauthorizedImplForwardAuth
		authz.fPortalURL = authzPortalURLFromQuery
	case AuthzImplAuthRequest:
		authz.fObjectGet = authzGetObjectImplAuthRequest
		authz.fHandleUnauthorized = authzHandleUnauthorizedImplAuthRequest
	case AuthzImplExtAuthz:
		authz.fObjectGet = authzGetObjectImplExtAuthz
		authz.fHandleUnauthorized = authzHandleUnauthorizedImplExtAuthz
		authz.fPortalURL = authzPortalURLFromHeader
	}

	return authz
}
