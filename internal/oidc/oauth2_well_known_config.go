package oidc

import (
	"encoding/json"
	"fmt"

	"github.com/authelia/authelia/internal/middlewares"
)

// WellKnownConfigurationJSON is the OIDC well known config struct.
type WellKnownConfigurationJSON struct {
	Issuer                 string   `json:"issuer"`
	AuthURL                string   `json:"authorization_endpoint"`
	TokenURL               string   `json:"token_endpoint"`
	JWKSURL                string   `json:"jwks_uri"`
	UserInfoURL            string   `json:"userinfo_endpoint"`
	Algorithms             []string `json:"id_token_signing_alg_values_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
}

// WellKnownConfigurationHandler handler serving the openid configuration.
func WellKnownConfigurationHandler(ctx *middlewares.AutheliaCtx) {
	var configuration WellKnownConfigurationJSON

	issuer, err := ctx.ForwardedProtoHost()

	if err != nil {
		issuer = fallbackOIDCIssuer
	}

	configuration.Issuer = issuer
	configuration.AuthURL = fmt.Sprintf("%s%s", issuer, authPath)
	configuration.TokenURL = fmt.Sprintf("%s%s", issuer, tokenPath)
	configuration.JWKSURL = fmt.Sprintf("%s%s", issuer, jwksPath)
	configuration.UserInfoURL = fmt.Sprintf("%s%s", issuer, userinfoPath)
	configuration.Algorithms = []string{"RS256"}
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

	if err := json.NewEncoder(ctx).Encode(configuration); err != nil {
		ctx.Error(err, "Failed to serve openid configuration")
	}
}
