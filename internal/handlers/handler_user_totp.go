package handlers

import (
	"errors"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/storage"
)

// UserTOTPGet returns the users TOTP configuration.
func UserTOTPGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	config, err := ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username)
	if err != nil {
		if errors.Is(err, storage.ErrNoTOTPConfiguration) {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.Error(err, "No TOTP Configuration.")
		} else {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.Error(err, "Unknown Error.")
		}

		return
	}

	if err = ctx.SetJSONBody(config); err != nil {
		ctx.Logger.Errorf("Unable to perform TOTP configuration response: %s", err)
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}
