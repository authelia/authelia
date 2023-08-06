package oidc

import (
	"context"

	"github.com/ory/fosite"
)

// FlowClientCredentialsTokenHandler handles the client credentials flow grants.
func (p *OpenIDConnectProvider) FlowClientCredentialsTokenHandler(ctx context.Context, requester fosite.AccessRequester, client fosite.Client) {
	scopes := requester.GetRequestedScopes()

	if len(scopes) == 0 {
		scopes = client.GetScopes()
	}

	for _, scope := range scopes {
		if p.GetScopeStrategy(ctx)(client.GetScopes(), scope) {
			requester.GrantScope(scope)
		}
	}

	audiences := requester.GetRequestedAudience()

	for _, audience := range audiences {
		if p.GetAudienceStrategy(ctx)(client.GetAudience(), []string{audience}) == nil {
			requester.GrantAudience(audience)
		}
	}
}
