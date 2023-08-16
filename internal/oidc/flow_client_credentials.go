package oidc

import (
	"context"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/x/errorsx"
)

var _ fosite.TokenEndpointHandler = (*ClientCredentialsGrantHandler)(nil)

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
		if clientCCFS, ok := client.(ClientCredentialsFlowScopeClient); ok && clientCCFS.GetClientCredentialsFlowGrantAllScopesWhenOmitted(ctx) {
			scopes = clientCCFS.GetScopes()
		}
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
