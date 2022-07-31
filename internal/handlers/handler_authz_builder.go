package handlers

import (
	"fmt"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func NewAuthzBuilder() *AuthzBuilder {
	return &AuthzBuilder{
		config: AuthzConfig{RefreshInterval: time.Second * -1},
	}
}

func (b *AuthzBuilder) WithStrategies(strategies ...AuthnStrategy) *AuthzBuilder {
	b.strategies = strategies

	return b
}

func (b *AuthzBuilder) WithStrategyCookie(refreshInterval time.Duration) *AuthzBuilder {
	b.strategies = append(b.strategies, NewCookieAuthnStrategy(refreshInterval))

	return b
}

func (b *AuthzBuilder) WithStrategyAuthorization() *AuthzBuilder {
	b.strategies = append(b.strategies, NewAuthorizationHeaderAuthnStrategy())

	return b
}

func (b *AuthzBuilder) WithStrategyProxyAuthorization() *AuthzBuilder {
	b.strategies = append(b.strategies, NewProxyAuthorizationHeaderAuthnStrategy())

	return b
}

func (b *AuthzBuilder) WithImplementationLegacy() *AuthzBuilder {
	b.impl = AuthzImplLegacy

	return b
}

func (b *AuthzBuilder) WithImplementationForwardAuth() *AuthzBuilder {
	b.impl = AuthzImplForwardAuth

	return b
}

func (b *AuthzBuilder) WithImplementationAuthRequest() *AuthzBuilder {
	b.impl = AuthzImplForwardAuth

	return b
}

func (b *AuthzBuilder) WithConfig(config *schema.Configuration) *AuthzBuilder {
	var refreshInterval time.Duration

	switch config.AuthenticationBackend.RefreshInterval {
	case schema.ProfileRefreshDisabled:
		refreshInterval = time.Second * -1
	case schema.ProfileRefreshAlways:
		refreshInterval = time.Second * 0
	default:
		refreshInterval, _ = utils.ParseDurationString(config.AuthenticationBackend.RefreshInterval)
	}

	fmt.Printf("refresh setting: %v\n", refreshInterval)
	fmt.Printf("refresh config: %v\n", config.AuthenticationBackend.RefreshInterval)

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

func (b *AuthzBuilder) WithAuthzConfig(config AuthzConfig) *AuthzBuilder {
	b.config = config

	return b
}

func (b AuthzBuilder) Build() (authz Authz) {
	authz = Authz{
		config:            b.config,
		strategies:        b.strategies,
		fObjectVerify:     authzObjectVerifyStandard,
		fHandleAuthorized: authzHandleAuthorizedStandard,
	}

	if len(authz.strategies) == 0 {
		switch b.impl {
		case AuthzImplLegacy:
			authz.strategies = []AuthnStrategy{NewLegacyHeaderAuthnStrategy(), NewCookieAuthnStrategy(b.config.RefreshInterval)}
		case AuthzImplAuthRequest:
			authz.strategies = []AuthnStrategy{
				HeaderAuthnStrategy{
					authn:           AuthnTypeProxyAuthorization,
					headerAuthorize: headerProxyAuthorization,
					handle:          true,
					status:          fasthttp.StatusUnauthorized,
				},
				NewCookieAuthnStrategy(0),
			}
		default:
			authz.strategies = []AuthnStrategy{NewProxyAuthorizationHeaderAuthnStrategy(), NewCookieAuthnStrategy(b.config.RefreshInterval)}
		}
	}

	switch b.impl {
	case AuthzImplLegacy:
		authz.fObjectGet = authzGetObjectImplLegacy
		authz.fHandleUnauthorized = authzHandleUnauthorizedImplLegacy
	case AuthzImplForwardAuth:
		authz.fObjectGet = authzGetObjectImplForwardAuth
		authz.fHandleUnauthorized = authzHandleUnauthorizedImplForwardAuth
	case AuthzImplAuthRequest:
		authz.fObjectGet = authzGetObjectImplAuthRequest
		authz.fHandleUnauthorized = authzHandleUnauthorizedImplAuthRequest
	}

	return authz
}
