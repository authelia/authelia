package handlers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/oidc"
)

func oidcWellKnown(ctx *middlewares.AutheliaCtx) {
	issuerHost, err := ctx.ForwardedProtoHost()
	issuerPath := string(ctx.Path())

	// TODO: Remove testing log.
	ctx.Logger.Debugf("host: %s, path: %s", issuerHost, issuerPath)
	if err != nil {
		ctx.Logger.Errorf("Error occurred determining OpenID Connect issuer details: %+v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	issuerURL, err := url.Parse(issuerHost)
	if err != nil {
		ctx.Logger.Errorf("Error occurred determining OpenID Connect issuer details: %+v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	issuerURL.Path = strings.TrimSuffix(issuerPath, pathOpenIDConnectWellKnown)

	issuer := issuerURL.String()

	// TODO: Remove testing log.
	ctx.Logger.Debugf("Issuer Is: %s", issuer)
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
		ctx.Logger.Errorf("Error occurred in json Encode: %+v", err)
		// TODO: Determine if this is the appropriate error code here.
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
