// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

type bodyAppAuthenticatorDevicePUT struct {
	Name string `json:"name"`
}

// AppAuthenticatorDevicesGET returns the Authelia app devices of the session user.
func AppAuthenticatorDevicesGET(ctx *middlewares.AutheliaCtx) {
	userSession, ok := appPortalSession(ctx, "retrieving the app authenticator devices", fasthttp.StatusForbidden, messageOperationFailed)
	if !ok {
		return
	}

	devices, err := ctx.Providers.StorageProvider.LoadAppAuthenticatorDevicesByUsername(ctx, userSession.Username)
	if err != nil {
		appPortalError(ctx, err, "Error occurred loading the app authenticator devices", fasthttp.StatusInternalServerError, messageOperationFailed)

		return
	}

	data := make([]model.AppAuthenticatorDeviceData, len(devices))

	for i, device := range devices {
		data[i] = device.ToData()
	}

	appPortalJSON(ctx, data, messageOperationFailed)
}

// AppAuthenticatorDevicePUT renames an Authelia app device of the session user.
func AppAuthenticatorDevicePUT(ctx *middlewares.AutheliaCtx) {
	userSession, device, ok := appPortalDevice(ctx, "renaming the app authenticator device")
	if !ok {
		return
	}

	body := bodyAppAuthenticatorDevicePUT{}

	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		appPortalError(ctx, err, "Error occurred renaming the app authenticator device: "+errStrReqBodyParse, fasthttp.StatusBadRequest, messageOperationFailed)

		return
	}

	name := strings.TrimSpace(body.Name)

	if n := utf8.RuneCountInString(name); n == 0 || n > appDeviceInfoMaxLength {
		ctx.Logger.WithField(logging.FieldUsername, userSession.Username).Errorf("Error occurred renaming the app authenticator device: the name must be between 1 and %d characters", appDeviceInfoMaxLength)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err := ctx.Providers.StorageProvider.UpdateAppAuthenticatorDeviceName(ctx, userSession.Username, device.ID, name); err != nil {
		appPortalError(ctx, err, "Error occurred renaming the app authenticator device", fasthttp.StatusInternalServerError, messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}

// AppAuthenticatorDeviceRemoveDELETE removes an Authelia app device of the session user.
func AppAuthenticatorDeviceRemoveDELETE(ctx *middlewares.AutheliaCtx) {
	userSession, device, ok := appPortalDevice(ctx, "removing the app authenticator device")
	if !ok {
		return
	}

	if err := ctx.Providers.StorageProvider.DeleteAppAuthenticatorDevice(ctx, device.DeviceID); err != nil {
		appPortalError(ctx, err, "Error occurred removing the app authenticator device", fasthttp.StatusInternalServerError, messageOperationFailed)

		return
	}

	ctx.Logger.WithFields(map[string]any{logging.FieldUsername: userSession.Username, appFieldDeviceID: device.DeviceID}).Info("App authenticator device removed by the user")

	ctxLogAppAuthenticatorEvent(ctx, device, false)

	ctx.ReplyOK()
}

func appPortalDevice(ctx *middlewares.AutheliaCtx, action string) (userSession session.UserSession, device *model.AppAuthenticatorDevice, ok bool) {
	if userSession, ok = appPortalSession(ctx, action, fasthttp.StatusForbidden, messageOperationFailed); !ok {
		return userSession, nil, false
	}

	id, err := strconv.Atoi(appUserValue(ctx, appUserValueID))
	if err != nil || id <= 0 {
		ctx.Logger.WithField(logging.FieldUsername, userSession.Username).Errorf("Error occurred %s: the device id is not valid", action)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, false
	}

	device, err = ctx.Providers.StorageProvider.LoadAppAuthenticatorDeviceByID(ctx, id)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		ctx.Logger.WithField(logging.FieldUsername, userSession.Username).Errorf("Error occurred %s: the device does not exist", action)
	case err != nil:
		appPortalError(ctx, err, "Error occurred "+action, fasthttp.StatusInternalServerError, messageOperationFailed)

		return userSession, nil, false
	case device.Username != userSession.Username:
		ctx.Logger.WithField(logging.FieldUsername, userSession.Username).Errorf("Error occurred %s: the device belongs to another user", action)
	default:
		return userSession, device, true
	}

	appPortalNotFound(ctx)

	return userSession, nil, false
}
