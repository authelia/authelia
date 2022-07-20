package handler

import (
	"errors"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middleware"
	"github.com/authelia/authelia/v4/internal/storage"
)

// UserTOTPInfoGET returns the users TOTP configuration.
func UserTOTPInfoGET(ctx *middleware.AutheliaCtx) {
	userSession := ctx.GetSession()

	config, err := ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username)
	if err != nil {
		if errors.Is(err, storage.ErrNoTOTPConfiguration) {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetJSONError("Could not find TOTP Configuration for user.")
			ctx.Logger.Errorf("Failed to lookup TOTP configuration for user '%s'", userSession.Username)
		} else {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetJSONError("Could not find TOTP Configuration for user.")
			ctx.Logger.Errorf("Failed to lookup TOTP configuration for user '%s' with unknown error: %v", userSession.Username, err)
		}

		return
	}

	if err = ctx.SetJSONBody(config); err != nil {
		ctx.Logger.Errorf("Unable to perform TOTP configuration response: %s", err)
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}
