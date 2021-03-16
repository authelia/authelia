package oidc

import (
	"net/http"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/internal/middlewares"
)

func revokeEndpoint(oauth2 fosite.OAuth2Provider) middlewares.AutheliaHandlerFunc {
	return func(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
		ctx.Logger.Debugf("Hit revoke endpoint")

		err := oauth2.NewRevocationRequest(ctx, req)

		oauth2.WriteRevocationResponse(rw, err)
	}
}
