package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func oidcWellKnown(ctx *middlewares.AutheliaCtx) {
	issuer, err := ctx.ExternalRootURL()
	if err != nil {
		ctx.Logger.Errorf("error occurred determining OpenID Connect issuer details: %+v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	wellKnown := oidc.WellKnownConfiguration{
		Issuer:  issuer,
		JWKSURI: fmt.Sprintf("%s%s", issuer, pathOpenIDConnectJWKs),

		AuthorizationEndpoint: fmt.Sprintf("%s%s", issuer, pathOpenIDConnectAuthorization),
		TokenEndpoint:         fmt.Sprintf("%s%s", issuer, pathOpenIDConnectToken),
		RevocationEndpoint:    fmt.Sprintf("%s%s", issuer, pathOpenIDConnectRevocation),
		UserinfoEndpoint:      fmt.Sprintf("%s%s", issuer, pathOpenIDConnectUserinfo),

		Algorithms:         []string{"RS256"},
		UserinfoAlgorithms: []string{"none", "RS256"},

		SubjectTypesSupported: []string{
			"public",
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
		ResponseModesSupported: []string{
			"form_post",
			"query",
			"fragment",
		},
		ScopesSupported: []string{
			"openid",
			"offline_access",
			"profile",
			"groups",
			"email",
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

		RequestURIParameterSupported:       false,
		BackChannelLogoutSupported:         false,
		FrontChannelLogoutSupported:        false,
		BackChannelLogoutSessionSupported:  false,
		FrontChannelLogoutSessionSupported: false,
	}

	ctx.SetContentType("application/json")

	if err := json.NewEncoder(ctx).Encode(wellKnown); err != nil {
		ctx.Logger.Errorf("error occurred in JSON encode: %+v", err)
		// TODO: Determine if this is the appropriate error code here.
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
