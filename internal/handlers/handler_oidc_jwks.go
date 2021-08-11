package handlers

import (
	"encoding/json"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

func oidcJWKs(ctx *middlewares.AutheliaCtx) {
	ctx.SetContentType("application/json")

	if err := json.NewEncoder(ctx).Encode(ctx.Providers.OpenIDConnect.KeyManager.GetKeySet()); err != nil {
		ctx.Error(err, "failed to serve jwk set")
	}
}
