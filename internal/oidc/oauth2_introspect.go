package oidc

import (
	"net/http"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/internal/middlewares"
)

func introspectEndpoint(oauth2 fosite.OAuth2Provider) middlewares.AutheliaHandlerFunc {
	return func(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
		oidcSession := newDefaultSession(ctx)

		ir, err := oauth2.NewIntrospectionRequest(ctx, req, oidcSession)
		if err != nil {
			ctx.Logger.Errorf("Error occurred in OIDC introspection: %+v", err)
			oauth2.WriteIntrospectionError(rw, err)

			return
		}

		oauth2.WriteIntrospectionResponse(rw, ir)
	}
}
