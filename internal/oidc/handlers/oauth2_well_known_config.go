package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/authelia/authelia/internal/middlewares"
)

func wellKnownConfigurationHandler(ctx *middlewares.AutheliaCtx) {
	var configuration WellKnownConfigurationJSON

	issuer, err := ctx.ForwardedProtoHost()

	if err != nil {
		issuer = fallbackOIDCIssuer
	}

	configuration.Issuer = issuer
	configuration.AuthURL = fmt.Sprintf("%s%s", issuer, authorizePath)
	configuration.TokenURL = fmt.Sprintf("%s%s", issuer, tokenPath)
	configuration.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, revokePath)
	configuration.JWKSURL = fmt.Sprintf("%s%s", issuer, jwksPath)
	configuration.Algorithms = []string{"RS256"}
	configuration.ScopesSupported = []string{
		"openid",
		"profile",
		"groups",
		"email",
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
		ctx.Error(err, "Failed to serve openid configuration")
	}
}
