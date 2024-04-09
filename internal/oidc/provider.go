package oidc

import (
	"fmt"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
)

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(config *schema.IdentityProvidersOpenIDConnect, store storage.Provider, templates *templates.Provider) (provider *OpenIDConnectProvider) {
	if config == nil {
		return nil
	}

	signer := NewKeyManager(config)

	provider = &OpenIDConnectProvider{
		Store:      NewStore(config, store),
		KeyManager: signer,
		Config:     NewConfig(config, signer, templates),
	}

	provider.Provider = oauthelia2.New(provider.Store, provider.Config)

	provider.Config.LoadHandlers(provider.Store)

	provider.discovery = NewOpenIDConnectWellKnownConfiguration(config)

	return provider
}

// GetOAuth2WellKnownConfiguration returns the discovery document for the OAuth Configuration.
func (p *OpenIDConnectProvider) GetOAuth2WellKnownConfiguration(issuer string) OAuth2WellKnownConfiguration {
	options := p.discovery.OAuth2WellKnownConfiguration.Copy()

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
	options := p.discovery.Copy()

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
