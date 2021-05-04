package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/middlewares"
)

func oidcWellKnown(ctx *middlewares.AutheliaCtx) {
	var configuration WellKnownConfigurationJSON

	issuer, err := ctx.ForwardedProtoHost()
	if err != nil {
		ctx.Logger.Errorf("Error occurred in ForwardedProtoHost: %+v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	configuration.Issuer = issuer
	configuration.AuthURL = fmt.Sprintf("%s%s", issuer, oidcAuthorizePath)
	configuration.TokenURL = fmt.Sprintf("%s%s", issuer, oidcTokenPath)
	configuration.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, oidcRevokePath)
	configuration.JWKSURL = fmt.Sprintf("%s%s", issuer, oidcJWKsPath)
	configuration.Algorithms = []string{"RS256"}
	configuration.ScopesSupported = []string{
		"openid",
		"profile",
		"groups",
		"email",
		// Determine if this is really mandatory knowing the RP can request for a refresh token through the authorize
		// endpoint anyway.
		"offline_access",
	}
	configuration.ClaimsSupported = []string{
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
		"groups",
		"name",
	}
	configuration.ResponseTypesSupported = []string{
		"code",
		"token",
		"id_token",
		"code token",
		"code id_token",
		"token id_token",
		"code token id_token",
		"none",
	}

	ctx.SetContentType("application/json")

	if err := json.NewEncoder(ctx).Encode(configuration); err != nil {
		ctx.Logger.Errorf("Error occurred in json Encode: %+v", err)
		// TODO: Determine if this is the appropriate error code here.
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
