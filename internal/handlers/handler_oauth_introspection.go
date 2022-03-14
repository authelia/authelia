package handlers

import (
	"net/http"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// OAuthIntrospectionPOST handles POST requests to the OAuth 2.0 Introspection endpoint.
//
// https://datatracker.ietf.org/doc/html/rfc7662
func OAuthIntrospectionPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	oidcSession := newOpenIDSession("")

	ir, err := ctx.Providers.OpenIDConnect.Fosite.NewIntrospectionRequest(ctx, req, oidcSession)

	if err != nil {
		ctx.Logger.Errorf("Error occurred in NewIntrospectionRequest: %+v", err)
		ctx.Providers.OpenIDConnect.Fosite.WriteIntrospectionError(rw, err)

		return
	}

	ctx.Providers.OpenIDConnect.Fosite.WriteIntrospectionResponse(rw, ir)
}
