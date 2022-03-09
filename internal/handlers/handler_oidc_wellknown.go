package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

func wellKnownOpenIDConnectConfigurationGET(ctx *middlewares.AutheliaCtx) {
	issuer, err := ctx.ExternalRootURL()
	if err != nil {
		ctx.Logger.Errorf("Error occurred determining OpenID Connect issuer details: %+v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	wellKnown := ctx.Providers.OpenIDConnect.GetOpenIDConnectWellKnownConfiguration(issuer)

	ctx.SetContentType("application/json")

	if err = json.NewEncoder(ctx).Encode(wellKnown); err != nil {
		ctx.Logger.Errorf("Error occurred in JSON encode: %+v", err)
		// TODO: Determine if this is the appropriate error code here.
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}

func wellKnownOAuthAuthorizationServerGET(ctx *middlewares.AutheliaCtx) {
	issuer, err := ctx.ExternalRootURL()
	if err != nil {
		ctx.Logger.Errorf("Error occurred determining OpenID Connect issuer details: %+v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	wellKnown := ctx.Providers.OpenIDConnect.GetOAuth2WellKnownConfiguration(issuer)

	ctx.SetContentType("application/json")

	if err = json.NewEncoder(ctx).Encode(wellKnown); err != nil {
		ctx.Logger.Errorf("Error occurred in JSON encode: %+v", err)
		// TODO: Determine if this is the appropriate error code here.
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
