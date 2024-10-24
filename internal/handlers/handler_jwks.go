package handlers

import (
	"encoding/json"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// JSONWebKeySetGET returns the JSON Web Key Set. Used in OAuth 2.0 and OpenID Connect 1.0.
func JSONWebKeySetGET(ctx *middlewares.AutheliaCtx) {
	ctx.SetContentTypeApplicationJSON()

	if err := json.NewEncoder(ctx).Encode(ctx.Providers.OpenIDConnect.Issuer.GetPublicJSONWebKeys(ctx)); err != nil {
		ctx.Error(err, "failed to serve json web key set")
	}
}
