package oidc

import (
	"fmt"
	"strings"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
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
		Store:      NewStore(config, store),
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

	provider.Config.LoadHandlers(provider.Store, provider.KeyManager.Strategy())

	provider.discovery = NewOpenIDConnectWellKnownConfiguration(config.EnablePKCEPlainChallenge, provider.Store.clients)

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

// GetAuthorizationBearerConfiguration returns the client and schema.OpenIDConnectAuthorizationBearerConfiguration if the target URI matches.
func (p *OpenIDConnectProvider) GetAuthorizationBearerConfiguration(targetURI string) (client *Client, config *schema.OpenIDConnectAuthorizationBearerConfiguration) {
	if targetURI == "" || p.Config.AuthorizationBearers.RedirectURI == nil {
		return nil, nil
	}

	var err error

	for _, c := range p.Config.AuthorizationBearers.Configurations {
		for _, prefix := range c.URLPrefixes {
			if strings.HasPrefix(strings.ToLower(targetURI), strings.ToLower(prefix)) {
				if client, err = p.GetFullClient(c.ClientID); err != nil {
					return nil, nil
				}

				return client, &c
			}
		}
	}

	return nil, nil
}
