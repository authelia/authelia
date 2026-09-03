package oidcrp

import (
	"context"
	"time"

	"authelia.com/provider/oauth2/token/jose"
)

// Discovery represents the subset of the OpenID Connect Discovery 1.0 document Authelia consumes.
type Discovery struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	JWKSURI               string   `json:"jwks_uri"`
	IDTokenSigningAlgs    []string `json:"id_token_signing_alg_values_supported"`
}

// IdentityClaims represents the validated claims Authelia consumes from an ID Token.
type IdentityClaims struct {
	Issuer                         string
	Subject                        string
	PreferredUsername              string
	Name                           string
	Email                          string
	AuthenticationMethodsReference []string
}

// KeySet resolves a JSON Web Key Set for a location, optionally bypassing any cache.
type KeySet interface {
	Resolve(ctx context.Context, location string, ignoreCache bool) (jwks *jose.JSONWebKeySet, err error)
}

// ValidateOptions represents the parameters an ID Token is validated against.
type ValidateOptions struct {
	Issuer   string
	ClientID string
	Nonce    string
	Alg      string
	JWKSURI  string
	Now      time.Time
	Leeway   time.Duration
}
