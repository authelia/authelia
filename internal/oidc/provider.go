package oidc

import (
	"fmt"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/handler/par"
	"github.com/ory/fosite/handler/pkce"
	"github.com/ory/herodot"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
)

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(config *schema.OpenIDConnectConfiguration, store storage.Provider) (provider *OpenIDConnectProvider, err error) {
	if config == nil {
		return nil, nil
	}

	provider = &OpenIDConnectProvider{
		JSONWriter: herodot.NewJSONWriter(nil),
		Store:      NewOpenIDConnectStore(config, store),
		Config:     NewConfig(config),
	}

	provider.OAuth2Provider = fosite.NewOAuth2Provider(provider.Store, provider.Config)

	if provider.KeyManager, err = NewKeyManagerWithConfiguration(config); err != nil {
		return nil, err
	}

	provider.Config.Strategy.OpenID = &openid.DefaultStrategy{
		Signer: provider.KeyManager.Strategy(),
		Config: provider.Config,
	}

	provider.LoadHandlers()

	provider.discovery = NewOpenIDConnectWellKnownConfiguration(config.EnablePKCEPlainChallenge, provider.Store.clients)

	return provider, nil
}

func (p *OpenIDConnectProvider) LoadHandlers() {
	validator := openid.NewOpenIDConnectRequestValidator(p.KeyManager.Strategy(), p.Config)

	handlers := []any{
		&oauth2.AuthorizeExplicitGrantHandler{
			AccessTokenStrategy:    p.Strategy.Core,
			RefreshTokenStrategy:   p.Strategy.Core,
			AuthorizeCodeStrategy:  p.Strategy.Core,
			CoreStorage:            p.Store,
			TokenRevocationStorage: p.Store,
			Config:                 p.Config,
		},
		&oauth2.AuthorizeImplicitGrantTypeHandler{
			AccessTokenStrategy: p.Strategy.Core,
			AccessTokenStorage:  p.Store,
			Config:              p.Config,
		},
		&oauth2.ClientCredentialsGrantHandler{
			HandleHelper: &oauth2.HandleHelper{
				AccessTokenStrategy: p.Strategy.Core,
				AccessTokenStorage:  p.Store,
				Config:              p.Config,
			},
			Config: p.Config,
		},
		&oauth2.RefreshTokenGrantHandler{
			AccessTokenStrategy:    p.Strategy.Core,
			RefreshTokenStrategy:   p.Strategy.Core,
			TokenRevocationStorage: p.Store,
			Config:                 p.Config,
		},
		&openid.OpenIDConnectExplicitHandler{
			IDTokenHandleHelper: &openid.IDTokenHandleHelper{
				IDTokenStrategy: p.Strategy.OpenID,
			},
			OpenIDConnectRequestValidator: validator,
			OpenIDConnectRequestStorage:   p.Store,
			Config:                        p.Config,
		},
		&openid.OpenIDConnectImplicitHandler{
			AuthorizeImplicitGrantTypeHandler: &oauth2.AuthorizeImplicitGrantTypeHandler{
				AccessTokenStrategy: p.Strategy.Core,
				AccessTokenStorage:  p.Store,
				Config:              p.Config,
			},
			IDTokenHandleHelper: &openid.IDTokenHandleHelper{
				IDTokenStrategy: p.Strategy.OpenID,
			},
			OpenIDConnectRequestValidator: validator,
			Config:                        p.Config,
		},
		&openid.OpenIDConnectHybridHandler{
			AuthorizeExplicitGrantHandler: &oauth2.AuthorizeExplicitGrantHandler{
				AccessTokenStrategy:   p.Strategy.Core,
				RefreshTokenStrategy:  p.Strategy.Core,
				AuthorizeCodeStrategy: p.Strategy.Core,
				CoreStorage:           p.Store,
				Config:                p.Config,
			},
			Config: p.Config,
			AuthorizeImplicitGrantTypeHandler: &oauth2.AuthorizeImplicitGrantTypeHandler{
				AccessTokenStrategy: p.Strategy.Core,
				AccessTokenStorage:  p.Store,
				Config:              p.Config,
			},
			IDTokenHandleHelper: &openid.IDTokenHandleHelper{
				IDTokenStrategy: p.Strategy.OpenID,
			},
			OpenIDConnectRequestValidator: validator,
			OpenIDConnectRequestStorage:   p.Store,
		},
		&openid.OpenIDConnectRefreshHandler{
			IDTokenHandleHelper: &openid.IDTokenHandleHelper{
				IDTokenStrategy: p.Strategy.OpenID,
			},
			Config: p.Config,
		},
		&oauth2.CoreValidator{
			CoreStrategy: p.Strategy.Core,
			CoreStorage:  p.Store,
			Config:       p.Config,
		},
		&oauth2.TokenRevocationHandler{
			AccessTokenStrategy:    p.Strategy.Core,
			RefreshTokenStrategy:   p.Strategy.Core,
			TokenRevocationStorage: p.Store,
		},
		&pkce.Handler{
			AuthorizeCodeStrategy: p.Strategy.Core,
			Storage:               p.Store,
			Config:                p.Config,
		},
		&par.PushedAuthorizeHandler{
			Storage: p.Store,
			Config:  p.Config,
		},
	}

	c := HandlersConfig{}

	for _, handler := range handlers {
		if h, ok := handler.(fosite.AuthorizeEndpointHandler); ok {
			c.AuthorizeEndpoint.Append(h)
		}
		if h, ok := handler.(fosite.TokenEndpointHandler); ok {
			c.TokenEndpoint.Append(h)
		}
		if h, ok := handler.(fosite.TokenIntrospector); ok {
			c.TokenIntrospection.Append(h)
		}
		if h, ok := handler.(fosite.RevocationHandler); ok {
			c.Revocation.Append(h)
		}
		if h, ok := handler.(fosite.PushedAuthorizeEndpointHandler); ok {
			c.PushedAuthorizeEndpoint.Append(h)
		}
	}

	p.Config.Handlers = c
}

// GetOAuth2WellKnownConfiguration returns the discovery document for the OAuth Configuration.
func (p *OpenIDConnectProvider) GetOAuth2WellKnownConfiguration(issuer string) OAuth2WellKnownConfiguration {
	options := OAuth2WellKnownConfiguration{
		CommonDiscoveryOptions: p.discovery.CommonDiscoveryOptions,
		OAuth2DiscoveryOptions: p.discovery.OAuth2DiscoveryOptions,
	}

	options.Issuer = issuer
	options.JWKSURI = fmt.Sprintf("%s%s", issuer, EndpointPathJWKs)

	options.IntrospectionEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathIntrospection)
	options.TokenEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathToken)

	options.AuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathAuthorization)
	options.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathRevocation)

	return options
}

// GetOpenIDConnectWellKnownConfiguration returns the discovery document for the OpenID Configuration.
func (p *OpenIDConnectProvider) GetOpenIDConnectWellKnownConfiguration(issuer string) OpenIDConnectWellKnownConfiguration {
	options := OpenIDConnectWellKnownConfiguration{
		CommonDiscoveryOptions:                          p.discovery.CommonDiscoveryOptions,
		OAuth2DiscoveryOptions:                          p.discovery.OAuth2DiscoveryOptions,
		OpenIDConnectDiscoveryOptions:                   p.discovery.OpenIDConnectDiscoveryOptions,
		OpenIDConnectFrontChannelLogoutDiscoveryOptions: p.discovery.OpenIDConnectFrontChannelLogoutDiscoveryOptions,
		OpenIDConnectBackChannelLogoutDiscoveryOptions:  p.discovery.OpenIDConnectBackChannelLogoutDiscoveryOptions,
	}

	options.Issuer = issuer
	options.JWKSURI = fmt.Sprintf("%s%s", issuer, EndpointPathJWKs)

	options.IntrospectionEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathIntrospection)
	options.TokenEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathToken)

	options.AuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathAuthorization)
	options.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathRevocation)
	options.UserinfoEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathUserinfo)

	return options
}
