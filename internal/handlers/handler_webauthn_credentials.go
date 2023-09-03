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

func getWebAuthnCredentialIDFromContext(ctx *middlewares.AutheliaCtx) (int, error) {
	credentialIDStr, ok := ctx.UserValue("credentialID").(string)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return 0, errors.New("Invalid credential ID type")
	}

	deviceID, err := strconv.Atoi(credentialIDStr)
	if err != nil {
		ctx.Error(err, messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return 0, err
	}

	return deviceID, nil
}

// WebAuthnCredentialsGET returns all credentials registered for the current user.
func WebAuthnCredentialsGET(ctx *middlewares.AutheliaCtx) {
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

	devices, err := ctx.Providers.StorageProvider.LoadWebAuthnCredentialsByUsername(ctx, origin.Hostname(), userSession.Username)

	if err != nil && err != storage.ErrNoWebAuthnCredential {
		ctx.Error(err, messageOperationFailed)
		return
	}

	if err = ctx.SetJSONBody(devices); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}
}

// WebAuthnCredentialPUT updates the description for a specific credential for the current user.
func WebAuthnCredentialPUT(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON bodyEditWebAuthnCredentialRequest

		id          int
		device      *model.WebAuthnCredential
		userSession session.UserSession

		err error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.ReplyForbidden()

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.Errorf("Unable to parse %s update request data for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Error(err, messageOperationFailed)

		return
	}

	if id, err = getWebAuthnCredentialIDFromContext(ctx); err != nil {
		return
	}

	if device, err = ctx.Providers.StorageProvider.LoadWebAuthnCredentialByID(ctx, id); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}

	if device.Username != userSession.Username {
		ctx.Error(fmt.Errorf("user '%s' tried to delete device with id '%d' which belongs to '%s", userSession.Username, device.ID, device.Username), messageOperationFailed)
		return
	}

	if err = ctx.Providers.StorageProvider.UpdateWebAuthnCredentialDescription(ctx, userSession.Username, id, bodyJSON.Description); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}
}

// WebAuthnCredentialDELETE deletes a specific credential for the current user.
func WebAuthnCredentialDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		id          int
		device      *model.WebAuthnCredential
		userSession session.UserSession
		err         error
	)

	if id, err = getWebAuthnCredentialIDFromContext(ctx); err != nil {
		return
	}

	if device, err = ctx.Providers.StorageProvider.LoadWebAuthnCredentialByID(ctx, id); err != nil {
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

	if err = ctx.Providers.StorageProvider.DeleteWebAuthnCredential(ctx, device.KID.String()); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}

	ctx.ReplyOK()
}
