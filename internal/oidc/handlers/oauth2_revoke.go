package handlers

import (
	"net/http"

	"github.com/authelia/authelia/internal/middlewares"
)

func revokeHandler(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	err := ctx.Providers.OpenIDConnect.Fosite.NewRevocationRequest(ctx, req)

	ctx.Providers.OpenIDConnect.Fosite.WriteRevocationResponse(rw, err)
}
