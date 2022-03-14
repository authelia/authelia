package handlers

import (
	"net/http"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// OAuthRevocationPOST handles POST requests to the OAuth 2.0 Revocation endpoint.
//
// https://datatracker.ietf.org/doc/html/rfc7009
func OAuthRevocationPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	err := ctx.Providers.OpenIDConnect.Fosite.NewRevocationRequest(ctx, req)

	ctx.Providers.OpenIDConnect.Fosite.WriteRevocationResponse(rw, err)
}
