package oidc

import (
	"fmt"
	"net/http"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
)

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(config *schema.Configuration, store storage.Provider, templates *templates.Provider) (provider *OpenIDConnectProvider) {
	if config == nil || config.IdentityProviders.OIDC == nil {
		return nil
	}

	issuer := NewIssuer(config.IdentityProviders.OIDC.JSONWebKeys)

	provider = &OpenIDConnectProvider{
		Store:  NewStore(config, store),
		Issuer: issuer,
		Config: NewConfig(config.IdentityProviders.OIDC, issuer, templates),
	}

	provider.Provider = oauthelia2.New(provider.Store, provider.Config)

	provider.LoadHandlers(provider.Store)

	provider.discovery = NewOpenIDConnectWellKnownConfiguration(config.IdentityProviders.OIDC)

	return provider
}

// GetOAuth2WellKnownConfiguration returns the discovery document for the OAuth Configuration.
func (p *OpenIDConnectProvider) GetOAuth2WellKnownConfiguration(issuer string) OAuth2WellKnownConfiguration {
	options := p.discovery.OAuth2WellKnownConfiguration.Copy()

	options.Issuer = issuer

	options.JWKSURI = fmt.Sprintf("%s%s", issuer, EndpointPathJWKs)
	options.AuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathAuthorization)
	options.DeviceAuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathDeviceAuthorization)
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
	options.DeviceAuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathDeviceAuthorization)
	options.PushedAuthorizationRequestEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathPushedAuthorizationRequest)
	options.TokenEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathToken)
	options.UserinfoEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathUserinfo)
	options.IntrospectionEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathIntrospection)
	options.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathRevocation)

	return options
}

func (p *OpenIDConnectProvider) WriteDynamicAuthorizeError(ctx Context, rw http.ResponseWriter, requester oauthelia2.Requester, err error) {
	switch r := requester.(type) {
	case oauthelia2.DeviceAuthorizeRequester:
		p.WriteRFC8628UserAuthorizeError(ctx, rw, r, err)
	case oauthelia2.AuthorizeRequester:
		p.WriteAuthorizeError(ctx, rw, r, err)
	}
}
