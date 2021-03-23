package handlers

import (
	"net/http"

	"github.com/authelia/authelia/internal/middlewares"
)

func oidcIntrospect(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	oidcSession := newDefaultOIDCSession(ctx)

	ir, err := ctx.Providers.OpenIDConnect.Fosite.NewIntrospectionRequest(ctx, req, oidcSession)
	if err != nil {
		ctx.Logger.Errorf("Error occurred in OIDC introspection: %+v", err)
		ctx.Providers.OpenIDConnect.Fosite.WriteIntrospectionError(rw, err)

		return
	}

	ctx.Providers.OpenIDConnect.Fosite.WriteIntrospectionResponse(rw, ir)
}
