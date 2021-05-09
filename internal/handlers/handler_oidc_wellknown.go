package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/oidc"
)

func oidcWellKnown(ctx *middlewares.AutheliaCtx) {
	// TODO (james-d-elliott): append the server.path here for path based installs. Also check other instances in OIDC.
	issuer, err := ctx.ForwardedProtoHost()
	if err != nil {
		ctx.Logger.Errorf("Error occurred in ForwardedProtoHost: %+v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	wellKnown := oidc.WellKnownConfiguration{
		Issuer:             issuer,
		AuthURL:            fmt.Sprintf("%s%s", issuer, oidcAuthorizePath),
		TokenURL:           fmt.Sprintf("%s%s", issuer, oidcTokenPath),
		RevocationEndpoint: fmt.Sprintf("%s%s", issuer, oidcRevokePath),
		JWKSURL:            fmt.Sprintf("%s%s", issuer, oidcJWKsPath),
		Algorithms:         []string{"RS256"},
		ScopesSupported: []string{
			"openid",
			"profile",
			"groups",
			"email",
			// Determine if this is really mandatory knowing the RP can request for a refresh token through the authorize
			// endpoint anyway.
			"offline_access",
		},
		ClaimsSupported: []string{
			"aud",
			"exp",
			"iat",
			"iss",
			"jti",
			"rat",
			"sub",
			"auth_time",
			"nonce",
			"email",
			"email_verified",
			"alt_emails",
			"groups",
			"name",
		},
		ResponseModesSupported: []string{
			"form_post",
			"query",
			"fragment",
		},
		ResponseTypesSupported: []string{
			"code",
			"token",
			"id_token",
			"code token",
			"code id_token",
			"token id_token",
			"code token id_token",
			"none",
		},
	}

	ctx.SetContentType("application/json")

	if err := json.NewEncoder(ctx).Encode(wellKnown); err != nil {
		ctx.Logger.Errorf("Error occurred in json Encode: %+v", err)
		// TODO: Determine if this is the appropriate error code here.
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
