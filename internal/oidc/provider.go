// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
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
	options.EndSessionEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathEndSession)

	return options
}

// WriteDynamicAuthorizeError writes the authorization error appropriate for the given requester.
func (p *OpenIDConnectProvider) WriteDynamicAuthorizeError(ctx Context, rw http.ResponseWriter, requester oauthelia2.Requester, err error) {
	switch r := requester.(type) {
	case oauthelia2.DeviceAuthorizeRequester:
		p.WriteRFC8628UserAuthorizeError(ctx, rw, r, err)
	case oauthelia2.AuthorizeRequester:
		p.WriteAuthorizeError(ctx, rw, r, err)
	}
}

// RPInitiatedLogoutProvider is implemented by an oauthelia2.Provider which supports parsing and validating
// OpenID Connect RP-Initiated Logout 1.0 end session requests.
type RPInitiatedLogoutProvider interface {
	NewRPInitiatedLogoutRequest(ctx context.Context, r *http.Request) (requester oauthelia2.RPInitiatedLogoutRequester, err error)
}

// NewRPInitiatedLogoutRequest parses and validates an OpenID Connect RP-Initiated Logout 1.0 end session request.
//
// It does not authenticate the client, end any session, or write a response. On error the returned requester never
// has a post logout redirect URI, so an error can't be redirected to an unvalidated URI.
func (p *OpenIDConnectProvider) NewRPInitiatedLogoutRequest(ctx context.Context, r *http.Request) (requester oauthelia2.RPInitiatedLogoutRequester, err error) {
	provider, ok := p.Provider.(RPInitiatedLogoutProvider)
	if !ok {
		return oauthelia2.NewRPInitiatedLogoutRequest(), oauthelia2.ErrServerError.
			WithDebug("The OpenID Connect 1.0 Provider does not support RP-Initiated Logout.")
	}

	return provider.NewRPInitiatedLogoutRequest(ctx, r)
}

// BackChannelLogoutProvider is implemented by an oauthelia2.Provider which supports delivering Logout Tokens to
// Relying Parties as per OpenID Connect Back-Channel Logout 1.0.
type BackChannelLogoutProvider interface {
	SendBackChannelLogout(ctx context.Context, requester oauthelia2.BackChannelLogoutRequester) (results []oauthelia2.BackChannelLogoutResult, err error)
}

// SendBackChannelLogout delivers a Logout Token to each of the clients named by the requester, as per OpenID
// Connect Back-Channel Logout 1.0.
//
// The clients are supplied by the caller: neither this provider nor the underlying library records which clients
// participated in a session, and neither ends any session. Delivery is best effort, so a Relying Party which is
// unreachable or which rejects its Logout Token is reported in its own result rather than failing the call. Only
// whole-operation failures are returned as err, and a result is returned for every client supplied, in the order
// supplied.
//
// OpenID Connect Back-Channel Logout 1.0 (https://openid.net/specs/openid-connect-backchannel-1_0.html)
func (p *OpenIDConnectProvider) SendBackChannelLogout(ctx context.Context, requester oauthelia2.BackChannelLogoutRequester) (results []oauthelia2.BackChannelLogoutResult, err error) {
	provider, ok := p.Provider.(BackChannelLogoutProvider)
	if !ok {
		return nil, oauthelia2.ErrServerError.
			WithDebug("The OpenID Connect 1.0 Provider does not support Back-Channel Logout.")
	}

	return provider.SendBackChannelLogout(ctx, requester)
}
