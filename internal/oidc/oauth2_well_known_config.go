package oidc

import (
	"encoding/json"
	"fmt"

	"github.com/authelia/authelia/internal/middlewares"
)

// WellKnownConfigurationHandler handler serving the openid configuration.
func WellKnownConfigurationHandler(ctx *middlewares.AutheliaCtx) {
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
	}
	configuration.ClaimsSupported = []string{
		"auth_time",
		"exp",
		"iss",
		"sub",
		"aud",
		"ist",
		"jti",
		"rat",
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
