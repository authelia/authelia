package oidc

import (
	"context"
	"net/url"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/x/errorsx"
)

// ClientCredentialsGrantHandler handles access requests for the Client Credentials Flow.
type ClientCredentialsGrantHandler struct {
	*oauth2.HandleHelper
	Config interface {
		fosite.ScopeStrategyProvider
		fosite.AudienceStrategyProvider
		fosite.AccessTokenLifespanProvider
	}
}

// HandleTokenEndpointRequest implements https://tools.ietf.org/html/rfc6749#section-6 and the
// fosite.TokenEndpointHandler for the Client Credentials Flow.
func (c *ClientCredentialsGrantHandler) HandleTokenEndpointRequest(ctx context.Context, request fosite.AccessRequester) error {
	if !c.CanHandleTokenEndpointRequest(ctx, request) {
		return errorsx.WithStack(fosite.ErrUnknownRequest)
	}

	client := request.GetClient()

	// The client MUST authenticate with the authorization server as described in Section 3.2.1.
	// This requirement is already fulfilled because fosite requires all token requests to be authenticated as described
	// in https://tools.ietf.org/html/rfc6749#section-3.2.1
	if client.IsPublic() {
		return errorsx.WithStack(fosite.ErrInvalidGrant.WithHint("The OAuth 2.0 Client is marked as public and is thus not allowed to use authorization grant 'client_credentials'."))
	}

	scopes := request.GetRequestedScopes()

	if len(scopes) == 0 {
		scopes = client.GetScopes()
	}

	for _, scope := range scopes {
		if !c.Config.GetScopeStrategy(ctx)(client.GetScopes(), scope) {
			return errorsx.WithStack(fosite.ErrInvalidScope.WithHintf("The OAuth 2.0 Client is not allowed to request scope '%s'.", scope))
		}

		request.GrantScope(scope)
	}

	if err := c.Config.GetAudienceStrategy(ctx)(client.GetAudience(), request.GetRequestedAudience()); err != nil {
		return err
	}

	// if the client is not public, he has already been authenticated by the access request handler.
	atLifespan := fosite.GetEffectiveLifespan(client, fosite.GrantTypeClientCredentials, fosite.AccessToken, c.Config.GetAccessTokenLifespan(ctx))
	request.GetSession().SetExpiresAt(fosite.AccessToken, time.Now().UTC().Add(atLifespan))

	return nil
}

// PopulateTokenEndpointResponse implements https://tools.ietf.org/html/rfc6749#section-4.4.3 and the
// fosite.TokenEndpointHandler for the Client Credentials Flow.
func (c *ClientCredentialsGrantHandler) PopulateTokenEndpointResponse(ctx context.Context, request fosite.AccessRequester, response fosite.AccessResponder) error {
	if !c.CanHandleTokenEndpointRequest(ctx, request) {
		return errorsx.WithStack(fosite.ErrUnknownRequest)
	}

	// TODO: remove?
	if !request.GetClient().GetGrantTypes().Has("client_credentials") {
		return errorsx.WithStack(fosite.ErrUnauthorizedClient.WithHint("The OAuth 2.0 Client is not allowed to use authorization grant 'client_credentials'."))
	}

	atLifespan := fosite.GetEffectiveLifespan(request.GetClient(), fosite.GrantTypeClientCredentials, fosite.AccessToken, c.Config.GetAccessTokenLifespan(ctx))

	return c.IssueAccessToken(ctx, atLifespan, request, response)
}

// CanSkipClientAuth implements the fosite.TokenEndpointHandler for the Client Credentials Flow.
func (c *ClientCredentialsGrantHandler) CanSkipClientAuth(ctx context.Context, requester fosite.AccessRequester) bool {
	return false
}

// CanHandleTokenEndpointRequest implements the fosite.TokenEndpointHandler for the Client Credentials Flow.
func (c *ClientCredentialsGrantHandler) CanHandleTokenEndpointRequest(ctx context.Context, requester fosite.AccessRequester) bool {
	// grant_type REQUIRED.
	// Value MUST be set to "client_credentials".
	return requester.GetGrantTypes().ExactOne(GrantTypeClientCredentials)
}

var (
	_ fosite.TokenEndpointHandler = (*ClientCredentialsGrantHandler)(nil)
)

// PopulateClientCredentialsFlowSessionWithAccessRequest is used to configure a session when performing a client credentials grant.
func PopulateClientCredentialsFlowSessionWithAccessRequest(ctx Context, client fosite.Client, session *Session) (err error) {
	var (
		issuer *url.URL
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		return fosite.ErrServerError.WithWrap(err).WithDebugf("Failed to determine the issuer with error: %s.", err.Error())
	}

	if client == nil {
		return fosite.ErrServerError.WithDebug("Failed to get the client for the request.")
	}

	session.Subject = ""
	session.Claims.Subject = client.GetID()
	session.ClientID = client.GetID()
	session.DefaultSession.Claims.Issuer = issuer.String()
	session.DefaultSession.Claims.IssuedAt = ctx.GetClock().Now().UTC()
	session.DefaultSession.Claims.RequestedAt = ctx.GetClock().Now().UTC()
	session.ClientCredentials = true

	return nil
}

// PopulateClientCredentialsFlowRequester is used to grant the authorized scopes and audiences when performing a client
// credentials grant.
func PopulateClientCredentialsFlowRequester(ctx Context, config fosite.Configurator, client fosite.Client, requester fosite.Requester) (err error) {
	if client == nil || config == nil || requester == nil {
		return fosite.ErrServerError.WithDebug("Failed to get the client, configuration, or requester for the request.")
	}

	scopes := requester.GetRequestedScopes()
	audience := requester.GetRequestedAudience()

	var authz, nauthz bool

	strategy := config.GetScopeStrategy(ctx)

	for _, scope := range scopes {
		switch scope {
		case ScopeOffline, ScopeOfflineAccess:
			break
		case ScopeAutheliaBearerAuthz:
			authz = true
		default:
			nauthz = true
		}

		if strategy(client.GetScopes(), scope) {
			requester.GrantScope(scope)
		} else {
			return fosite.ErrInvalidScope.WithDebugf("The scope '%s' is not authorized on client with id '%s'.", scope, client.GetID())
		}
	}

	if authz && nauthz {
		return fosite.ErrInvalidScope.WithDebugf("The scope '%s' must only be requested by itself or with the '%s' scope, no other scopes are permitted.", ScopeAutheliaBearerAuthz, ScopeOfflineAccess)
	}

	if authz && len(audience) == 0 {
		return fosite.ErrInvalidRequest.WithDebugf("The scope '%s' requires the request also include an audience.", ScopeAutheliaBearerAuthz)
	}

	if err = config.GetAudienceStrategy(ctx)(client.GetAudience(), audience); err != nil {
		return err
	}

	for _, aud := range audience {
		requester.GrantAudience(aud)
	}

	return nil
}
