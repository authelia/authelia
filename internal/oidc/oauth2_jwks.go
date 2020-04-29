package oidc

import (
	"encoding/json"

	"github.com/authelia/authelia/internal/middlewares"
	"gopkg.in/square/go-jose.v2"
)

// JWKsGet handler serving the jwks used to verify the JWT tokens.
func JWKsGet(req *middlewares.AutheliaCtx) {
	key := jose.JSONWebKey{}
	key.Key = &privateKey.PublicKey
	key.KeyID = "main-key"
	key.Algorithm = "RS256"
	key.Use = "sig"

	keySet := new(jose.JSONWebKeySet)
	keySet.Keys = append(keySet.Keys, key)

	if err := json.NewEncoder(req).Encode(keySet); err != nil {
		req.Error(err, "failed to serve jwk set")
	}
}
