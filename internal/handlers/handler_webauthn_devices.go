package handlers

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/storage"
)

func getWebauthnDeviceIDFromContext(ctx *middlewares.AutheliaCtx) (int, error) {
	deviceIDStr, ok := ctx.UserValue("deviceID").(string)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return 0, errors.New("Invalid device ID type")
	}

	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		ctx.Error(err, messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return 0, err
	}

	return deviceID, nil
}

// WebauthnDevicesGet returns all devices registered for the current user.
func WebauthnDevicesGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	devices, err := ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, userSession.Username)

	if err != nil && err != storage.ErrNoWebauthnDevice {
		ctx.Error(err, messageOperationFailed)
		return
	}

	if err = ctx.SetJSONBody(devices); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}
}

// WebauthnDeviceUpdate updates the description for a specific device for the current user.
func WebauthnDeviceUpdate(ctx *middlewares.AutheliaCtx) {
	type requestPostData struct {
		Description string `json:"description"`
	}

	var postData *requestPostData

	userSession := ctx.GetSession()

	err := json.Unmarshal(ctx.PostBody(), &postData)
	if err != nil {
		ctx.Logger.Errorf("Unable to parse %s update request data for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Error(err, messageOperationFailed)

		return
	}

	deviceID, err := getWebauthnDeviceIDFromContext(ctx)
	if err != nil {
		return
	}

	if err := ctx.Providers.StorageProvider.UpdateWebauthnDeviceDescription(ctx, userSession.Username, deviceID, postData.Description); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}
}

// WebauthnDeviceDelete deletes a specific device for the current user.
func WebauthnDeviceDelete(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	deviceID, err := getWebauthnDeviceIDFromContext(ctx)
	if err != nil {
		return
	}

	if err := ctx.Providers.StorageProvider.DeleteWebauthnDeviceByUsernameAndID(ctx, userSession.Username, deviceID); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}
}
