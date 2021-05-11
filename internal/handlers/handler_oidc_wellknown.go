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
		Issuer:              issuer,
		AuthURL:             fmt.Sprintf("%s%s", issuer, oidcAuthorizePath),
		TokenURL:            fmt.Sprintf("%s%s", issuer, oidcTokenPath),
		RevocationURL:       fmt.Sprintf("%s%s", issuer, oidcRevokePath),
		UserinfoURL:         fmt.Sprintf("%s%s", issuer, oidcUserinfoPath),
		JWKSURL:             fmt.Sprintf("%s%s", issuer, oidcJWKsPath),
		IDTokenAlgorithms:   []string{"RS256"},
		UserinfoAlgorithms:  []string{"none", "RS256"},
		RequestURIParameter: false,
		GrantTypes:          []string{"authorization_code", "refresh_token"},
		SubjectTypes:        []string{"public"},
		ResponseTypes: []string{
			"none", "code", "token", "id_token", "code token", "code id_token", "token id_token", "code token id_token",
		},
		ResponseModes: []string{"form_post", "query", "fragment"},
		Scopes:        []string{"openid", "offline_access", "profile", "groups", "email"},
		Claims: []string{
			"aud", "exp", "iat", "iss", "jti", "rat", "sub", "auth_time", "nonce",
			"email", "email_verified", "alt_emails", "groups", "name",
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
