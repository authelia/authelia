package handlers

import (
	"net/http"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

func oidcRevocation(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var err error

	if err = ctx.Providers.OpenIDConnect.Fosite.NewRevocationRequest(ctx, req); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("Revocation Request failed with error: %+v", rfc)
	}

	ctx.Providers.OpenIDConnect.Fosite.WriteRevocationResponse(rw, err)
}
