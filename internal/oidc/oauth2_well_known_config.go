package oidc

import (
	"encoding/json"

	"github.com/authelia/authelia/internal/middlewares"
)

type configurationJSON struct {
	Issuer                 string   `json:"issuer"`
	AuthURL                string   `json:"authorization_endpoint"`
	TokenURL               string   `json:"token_endpoint"`
	JWKSURL                string   `json:"jwks_uri"`
	UserInfoURL            string   `json:"userinfo_endpoint"`
	Algorithms             []string `json:"id_token_signing_alg_values_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
}

// WellKnownConfigurationGet handler serving the openid configuration.
func WellKnownConfigurationGet(req *middlewares.AutheliaCtx) {
	var configuration configurationJSON

	configuration.Issuer = "https://login.example.com:8080"
	configuration.AuthURL = "https://login.example.com:8080/api/oidc/auth"
	configuration.TokenURL = "https://login.example.com:8080/api/oidc/token"
	configuration.JWKSURL = "https://login.example.com:8080/api/oidc/jwks"
	configuration.UserInfoURL = "https://login.example.com:8080/api/oidc/userinfo"
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

	if err := json.NewEncoder(req).Encode(configuration); err != nil {
		req.Error(err, "Failed to serve openid configuration")
	}
}
