package oidc

import (
	"net/http"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/internal/middlewares"
)

func tokenEndpoint(oauth2 fosite.OAuth2Provider) middlewares.AutheliaHandlerFunc {
	return func(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
		oidcSession := newDefaultSession(ctx)
		accessRequest, err := oauth2.NewAccessRequest(ctx, req, oidcSession)

		if err != nil {
			ctx.Logger.Errorf("Error occurred in NewAccessRequest: %+v", err)
			oauth2.WriteAccessError(rw, accessRequest, err)

			return
		}

		// If this is a client_credentials grant, grant all scopes the client is allowed to perform.
		if accessRequest.GetGrantTypes().ExactOne("client_credentials") {
			for _, scope := range accessRequest.GetRequestedScopes() {
				if fosite.HierarchicScopeStrategy(accessRequest.GetClient().GetScopes(), scope) {
					accessRequest.GrantScope(scope)
				}
			}
		}

		response, err := oauth2.NewAccessResponse(ctx, accessRequest)
		if err != nil {
			ctx.Logger.Errorf("Error occurred in NewAccessResponse: %+v", err)
			oauth2.WriteAccessError(rw, accessRequest, err)

			return
		}

		oauth2.WriteAccessResponse(rw, accessRequest, response)
	}
}
