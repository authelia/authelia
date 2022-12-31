package oidc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/herodot"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
)

const (
	WriteFormPostResponseContextKey ContextKey = iota
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

// WriteAuthorizeResponse persists the AuthorizeSession in the store and redirects the user agent to the provided
// redirect url or returns an error if storage failed.
func (p *OpenIDConnectProvider) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, requester fosite.AuthorizeRequester, responder fosite.AuthorizeResponder) {
	writeFormPostResponseFn := getWriteFormPostResponseFn(ctx)
	if writeFormPostResponseFn == nil || requester.GetResponseMode() != fosite.ResponseModeFormPost {
		p.OAuth2Provider.WriteAuthorizeResponse(ctx, rw, requester, responder)
		return
	}

	wh := rw.Header()
	rh := responder.GetHeader()

	for k := range rh {
		wh.Set(k, rh.Get(k))
	}

	clientID := requester.GetClient().GetID()
	client, _ := p.Store.GetFullClient(clientID)

	data := map[string]any{
		"ClientDescription": client.Description,
		"RedirURL":          requester.GetRedirectURI(),
		"Parameters":        responder.GetParameters(),
	}
	writeFormPostResponseFn(data)
}

// WriteAuthorizeError returns the error codes to the redirection endpoint or shows the error to the user, if no valid
// redirect uri was given. Implements rfc6749#section-4.1.2.1.
func (p *OpenIDConnectProvider) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, requester fosite.AuthorizeRequester, err error) {
	writeFormPostResponseFn := getWriteFormPostResponseFn(ctx)
	if writeFormPostResponseFn == nil || requester.GetResponseMode() != fosite.ResponseModeFormPost {
		p.OAuth2Provider.WriteAuthorizeError(ctx, rw, requester, err)
		return
	}

	rfcerr := fosite.ErrorToRFC6749Error(err).
		WithLegacyFormat(p.Config.GetUseLegacyErrorFormat(ctx)).
		WithExposeDebug(p.Config.GetSendDebugMessagesToClients(ctx)).
		WithLocalizer(p.Config.GetMessageCatalog(ctx), getLangFromRequester(requester))

	errors := rfcerr.ToValues()
	errors.Set("state", requester.GetState())

	redirectURI := requester.GetRedirectURI()

	// The endpoint URI MUST NOT include a fragment component.
	redirectURI.Fragment = ""

	clientID := requester.GetClient().GetID()
	client, _ := p.Store.GetFullClient(clientID)

	data := map[string]any{
		"ClientDescription": client.Description,
		"RedirURL":          redirectURI.String(),
		"Parameters":        errors,
	}
	writeFormPostResponseFn(data)
}

func getWriteFormPostResponseFn(ctx context.Context) func(templateData map[string]any) {
	if fn := ctx.Value(WriteFormPostResponseContextKey); fn != nil {
		if fn, ok := fn.(func(templateData map[string]any)); ok {
			return fn
		}
	}

	return nil
}

func getLangFromRequester(requester fosite.Requester) language.Tag {
	lang := language.English
	if g11nContext, ok := requester.(fosite.G11NContext); ok {
		lang = g11nContext.GetLang()
	}

	return lang
}
