package handlers

import (
	"encoding/json"

	"github.com/authelia/authelia/internal/middlewares"
)

func jwksHandler(ctx *middlewares.AutheliaCtx) {
	ctx.SetContentType("application/json")

	if err := json.NewEncoder(ctx).Encode(ctx.Providers.OpenIDConnect.GetKeySet()); err != nil {
		ctx.Error(err, "failed to serve jwk set")
	}
}
