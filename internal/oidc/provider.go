package oidc

import (
	"fmt"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/herodot"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
)

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(config *schema.OpenIDConnectConfiguration, store storage.Provider, templates *templates.Provider) (provider *OpenIDConnectProvider, err error) {
	if config == nil {
		return nil, nil
	}

	provider = &OpenIDConnectProvider{
		JSONWriter: herodot.NewJSONWriter(nil),
		Store:      NewStore(config, store),
		Config:     NewConfig(config, templates),
	}

	provider.OAuth2Provider = fosite.NewOAuth2Provider(provider.Store, provider.Config)

	if provider.KeyManager, err = NewKeyManagerWithConfiguration(config); err != nil {
		return nil, err
	}

	provider.Config.Strategy.OpenID = &openid.DefaultStrategy{
		Signer: provider.KeyManager.Strategy(),
		Config: provider.Config,
	}

	provider.Config.LoadHandlers(provider.Store, provider.KeyManager.Strategy())

	provider.discovery = NewOpenIDConnectWellKnownConfiguration(config, provider.Store.clients)

	return provider, nil
}

// GetOAuth2WellKnownConfiguration returns the discovery document for the OAuth Configuration.
func (p *OpenIDConnectProvider) GetOAuth2WellKnownConfiguration(issuer string) OAuth2WellKnownConfiguration {
	options := OAuth2WellKnownConfiguration{
		CommonDiscoveryOptions: p.discovery.CommonDiscoveryOptions,
		OAuth2DiscoveryOptions: p.discovery.OAuth2DiscoveryOptions,
	}

	options.Issuer = issuer

	options.JWKSURI = fmt.Sprintf("%s%s", issuer, EndpointPathJWKs)
	options.AuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathAuthorization)
	options.PushedAuthorizationRequestEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathPushedAuthorizationRequest)
	options.TokenEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathToken)
	options.IntrospectionEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathIntrospection)
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
	options.AuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathAuthorization)
	options.PushedAuthorizationRequestEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathPushedAuthorizationRequest)
	options.TokenEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathToken)
	options.UserinfoEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathUserinfo)
	options.IntrospectionEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathIntrospection)
	options.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathRevocation)

	return options
}
