package handlers

import (
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// WebauthnDevicesGet returns all devices registered for the current user.
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

// WebauthnDeviceDelete deletes a specific device for the current user.
func WebauthnDeviceDelete(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	deviceIDStr, ok := ctx.UserValue("deviceID").(string)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		ctx.Error(err, messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	devices, err := ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, userSession.Username)
	if err != nil {
		ctx.Error(err, messageOperationFailed)
	}

	for _, existingDevice := range devices {
		if existingDevice.ID == deviceID {
			if err := ctx.Providers.StorageProvider.DeleteWebauthnDeviceByUsernameAndID(ctx, userSession.Username, deviceID); err != nil {
				ctx.Error(err, messageOperationFailed)
			}

			break
		}
	}
}
