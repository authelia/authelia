package oidc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/i18n"
	"authelia.com/provider/oauth2/x/errorsx"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
)

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(config *schema.Configuration, store storage.Provider, templates *templates.Provider) (provider *OpenIDConnectProvider) {
	if config == nil || config.IdentityProviders.OIDC == nil {
		return nil
	}

	signer := NewKeyManager(config.IdentityProviders.OIDC)

	provider = &OpenIDConnectProvider{
		Store:      NewStore(config, store),
		KeyManager: signer,
		Config:     NewConfig(config.IdentityProviders.OIDC, signer, templates),
	}

	provider.Provider = oauthelia2.New(provider.Store, provider.Config)

	provider.Config.LoadHandlers(provider.Store)

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

type bodyDeviceCodeUserAuthorizeRequest struct {
	ID       string `json:"id"`
	UserCode string `json:"user_code"`
}

func (p *OpenIDConnectProvider) NewRFC8628UserAuthorizeRequest(ctx Context, req *http.Request) (requester oauthelia2.DeviceAuthorizeRequester, err error) {
	body := &bodyDeviceCodeUserAuthorizeRequest{}

	if err = json.NewDecoder(req.Body).Decode(body); err != nil {
		return nil, errorsx.WithStack(oauthelia2.ErrInvalidRequest.WithHint("Unable to parse HTTP body, make sure to send a properly formatted json body body.").WithWrap(err).WithDebugError(err))
	}

	request := oauthelia2.NewDeviceAuthorizeRequest()
	request.Lang = i18n.GetLangFromRequest(p.Config.GetMessageCatalog(ctx), req)

	request.Form = url.Values{
		"user_code": []string{body.UserCode},
		"id":        []string{body.ID},
	}

	for _, h := range p.Config.GetRFC8628UserAuthorizeEndpointHandlers(ctx) {
		if err = h.HandleRFC8628UserAuthorizeEndpointRequest(ctx, request); err != nil && !errors.Is(err, oauthelia2.ErrUnknownRequest) {
			return nil, err
		}
	}

	return request, nil
}

func (p *OpenIDConnectProvider) WriteDynamicAuthorizeError(ctx Context, rw http.ResponseWriter, requester oauthelia2.Requester, err error) {
	switch r := requester.(type) {
	case oauthelia2.DeviceAuthorizeRequester:
		p.WriteRFC8628UserAuthorizeError(ctx, rw, r, err)
	case oauthelia2.AuthorizeRequester:
		p.WriteAuthorizeError(ctx, rw, r, err)
	}
}
