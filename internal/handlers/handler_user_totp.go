package handlers

import (
	"errors"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

// UserTOTPInfoGET returns the users TOTP configuration.
func UserTOTPInfoGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.ReplyForbidden()

		return
	}

	var config *model.TOTPConfiguration

	if config, err = ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username); err != nil {
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
