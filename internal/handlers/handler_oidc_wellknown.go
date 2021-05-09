package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/oidc"
)

func oidcWellKnown(ctx *middlewares.AutheliaCtx) {
	if ctx.Providers.OpenIDConnect.WellKnown == nil {
		ctx.Providers.OpenIDConnect.WellKnown = &oidc.WellKnownConfiguration{
			Issuer:             ctx.Configuration.ExternalURL,
			AuthURL:            fmt.Sprintf("%s%s", ctx.Configuration.ExternalURL, oidcAuthorizePath),
			TokenURL:           fmt.Sprintf("%s%s", ctx.Configuration.ExternalURL, oidcTokenPath),
			RevocationEndpoint: fmt.Sprintf("%s%s", ctx.Configuration.ExternalURL, oidcRevokePath),
			UserinfoEndpoint:   fmt.Sprintf("%s%s", ctx.Configuration.ExternalURL, oidcUserinfoPath),
			JWKSURL:            fmt.Sprintf("%s%s", ctx.Configuration.ExternalURL, oidcJWKsPath),
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
	}

	ctx.SetContentType("application/json")

	if err := json.NewEncoder(ctx).Encode(*ctx.Providers.OpenIDConnect.WellKnown); err != nil {
		ctx.Logger.Errorf("Error occurred in json Encode: %+v", err)
		// TODO: Determine if this is the appropriate error code here.
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
