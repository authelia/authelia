package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
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

// WebauthnDevicesGET returns all devices registered for the current user.
func WebauthnDevicesGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		origin      *url.URL
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.ReplyForbidden()

		return
	}

	if origin, err = ctx.GetOrigin(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving origin")

		ctx.ReplyForbidden()

		return
	}

	devices, err := ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, origin.Hostname(), userSession.Username)

	if err != nil && err != storage.ErrNoWebauthnDevice {
		ctx.Error(err, messageOperationFailed)
		return
	}

	if err = ctx.SetJSONBody(devices); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}
}

// WebauthnDevicePUT updates the description for a specific device for the current user.
func WebauthnDevicePUT(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON bodyEditWebauthnDeviceRequest

		id          int
		device      *model.WebauthnDevice
		userSession session.UserSession

		err error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.ReplyForbidden()

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.Errorf("Unable to parse %s update request data for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Error(err, messageOperationFailed)

		return
	}

	if id, err = getWebauthnDeviceIDFromContext(ctx); err != nil {
		return
	}

	if device, err = ctx.Providers.StorageProvider.LoadWebauthnDeviceByID(ctx, id); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}

	if device.Username != userSession.Username {
		ctx.Error(fmt.Errorf("user '%s' tried to delete device with id '%d' which belongs to '%s", userSession.Username, device.ID, device.Username), messageOperationFailed)
		return
	}

	if err = ctx.Providers.StorageProvider.UpdateWebauthnDeviceDescription(ctx, userSession.Username, id, bodyJSON.Description); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}
}

// WebauthnDeviceDELETE deletes a specific device for the current user.
func WebauthnDeviceDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		id          int
		device      *model.WebauthnDevice
		userSession session.UserSession
		err         error
	)

	if id, err = getWebauthnDeviceIDFromContext(ctx); err != nil {
		return
	}

	if device, err = ctx.Providers.StorageProvider.LoadWebauthnDeviceByID(ctx, id); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.ReplyForbidden()

		return
	}

	if device.Username != userSession.Username {
		ctx.Error(fmt.Errorf("user '%s' tried to delete device with id '%d' which belongs to '%s", userSession.Username, device.ID, device.Username), messageOperationFailed)
		return
	}

	if err = ctx.Providers.StorageProvider.DeleteWebauthnDevice(ctx, device.KID.String()); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}

	ctx.ReplyOK()
}
