package handlers

import (
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/valyala/fasthttp"
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

// WithImplementationLegacy configures this builder to output an Authz which is used with the Legacy
// implementation which is a mix of the other implementations and usually works with most proxies.
func (b *AuthzBuilder) WithImplementationLegacy() *AuthzBuilder {
	b.implementation = AuthzImplLegacy

	return b
}

// WithImplementationForwardAuth configures this builder to output an Authz which is used with the ForwardAuth
// implementation traditionally used by Traefik, Caddy, and Skipper.
func (b *AuthzBuilder) WithImplementationForwardAuth() *AuthzBuilder {
	b.implementation = AuthzImplForwardAuth

	return b
}

// WithImplementationAuthRequest configures this builder to output an Authz which is used with the AuthRequest
// implementation traditionally used by NGINX.
func (b *AuthzBuilder) WithImplementationAuthRequest() *AuthzBuilder {
	b.implementation = AuthzImplAuthRequest

	return b
}

// WithImplementationExtAuthz configures this builder to output an Authz which is used with the ExtAuthz
// implementation traditionally used by Envoy.
func (b *AuthzBuilder) WithImplementationExtAuthz() *AuthzBuilder {
	b.implementation = AuthzImplExtAuthz

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
	}

	return b
}

// WithEndpointConfig configures the AuthzBuilder with a *schema.ServerAuthzEndpointConfig. Should be called AFTER
// WithConfig or WithAuthzConfig.
func (b *AuthzBuilder) WithEndpointConfig(config schema.ServerAuthzEndpoint) *AuthzBuilder {
	switch config.Implementation {
	case AuthzImplForwardAuth.String():
		b.WithImplementationForwardAuth()
	case AuthzImplAuthRequest.String():
		b.WithImplementationAuthRequest()
	case AuthzImplExtAuthz.String():
		b.WithImplementationExtAuthz()
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
			b.strategies = append(b.strategies, NewHeaderProxyAuthorizationAuthRequestAuthnStrategy())
		case AuthnStrategyHeaderLegacy:
			b.strategies = append(b.strategies, NewHeaderLegacyAuthnStrategy())
		}
	}

	return b
}

// Build returns a new Authz from the currently configured options in this builder.
func (b *AuthzBuilder) Build() (authz *Authz) {
	authz = &Authz{
		config:           b.config,
		strategies:       b.strategies,
		handleAuthorized: handleAuthzAuthorizedStandard,
		implementation:   b.implementation,
	}

	authz.config.StatusCodeBadRequest = fasthttp.StatusBadRequest

	if len(authz.strategies) == 0 {
		switch b.implementation {
		case AuthzImplLegacy:
			authz.strategies = []AuthnStrategy{NewHeaderLegacyAuthnStrategy(), NewCookieSessionAuthnStrategy(b.config.RefreshInterval)}
		case AuthzImplAuthRequest:
			authz.strategies = []AuthnStrategy{NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(), NewCookieSessionAuthnStrategy(b.config.RefreshInterval)}
		default:
			authz.strategies = []AuthnStrategy{NewHeaderProxyAuthorizationAuthnStrategy(), NewCookieSessionAuthnStrategy(b.config.RefreshInterval)}
		}
	}

	switch b.implementation {
	case AuthzImplLegacy:
		authz.config.StatusCodeBadRequest = fasthttp.StatusUnauthorized
		authz.handleGetObject = handleAuthzGetObjectLegacy
		authz.handleUnauthorized = handleAuthzUnauthorizedLegacy
		authz.handleGetAutheliaURL = handleAuthzPortalURLLegacy
	case AuthzImplForwardAuth:
		authz.handleGetObject = handleAuthzGetObjectForwardAuth
		authz.handleUnauthorized = handleAuthzUnauthorizedForwardAuth
		authz.handleGetAutheliaURL = handleAuthzPortalURLFromQuery
	case AuthzImplAuthRequest:
		authz.handleGetObject = handleAuthzGetObjectAuthRequest
		authz.handleUnauthorized = handleAuthzUnauthorizedAuthRequest
		authz.handleGetAutheliaURL = handleAuthzPortalURLFromQuery
	case AuthzImplExtAuthz:
		authz.handleGetObject = handleAuthzGetObjectExtAuthz
		authz.handleUnauthorized = handleAuthzUnauthorizedExtAuthz
		authz.handleGetAutheliaURL = handleAuthzPortalURLFromHeader
	}

	return authz
}
