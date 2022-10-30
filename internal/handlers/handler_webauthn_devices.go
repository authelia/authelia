package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func WebauthnDevicesGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	devices, err := ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, userSession.Username)

	if err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}

	if err = ctx.SetJSONBody(devices); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}
}
