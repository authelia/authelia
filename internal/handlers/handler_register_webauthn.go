package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

// WebAuthnRegistrationPUT returns the attestation challenge from the server.
func WebAuthnRegistrationPUT(ctx *middlewares.AutheliaCtx) {
	var (
		w           *webauthn.WebAuthn
		user        *model.WebAuthnUser
		userSession session.UserSession
		bodyJSON    bodyRegisterWebAuthnPUTRequest
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration challenge", regulation.AuthTypeWebAuthn)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if w, err = newWebAuthn(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to create provider to generate %s registration challenge for user '%s'", regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to parse %s registration request PUT data for user '%s'", regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if length := len(bodyJSON.Description); length == 0 || length > 64 {
		ctx.Logger.Errorf("Failed to validate the user chosen display name for during %s registration for user '%s': the value has a length of %d but must be between 1 and 64", regulation.AuthTypeWebAuthn, userSession.Username, length)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	devices, err := ctx.Providers.StorageProvider.LoadWebAuthnCredentialsByUsername(ctx, w.Config.RPID, userSession.Username)
	if err != nil && err != storage.ErrNoWebAuthnCredential {
		ctx.Logger.WithError(err).Errorf("Unable to load existing %s devices for user '%s'", regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	for _, device := range devices {
		if strings.EqualFold(device.Description, bodyJSON.Description) {
			ctx.Logger.Errorf("Unable to generate %s registration challenge: device for for user '%s' with display name '%s' already exists", regulation.AuthTypeWebAuthn, userSession.Username, bodyJSON.Description)

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetJSONError(messageSecurityKeyDuplicateName)

			return
		}
	}

	if user, err = getWebAuthnUserByRPID(ctx, userSession.Username, userSession.DisplayName, w.Config.RPID); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to load %s devices for registration challenge for user '%s'", regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	var (
		creation *protocol.CredentialCreation
	)

	opts := []webauthn.RegistrationOption{
		webauthn.WithExclusions(user.WebAuthnCredentialDescriptors()),
		webauthn.WithExtensions(map[string]any{"credProps": true}),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementDiscouraged),
	}

	data := session.WebAuthn{
		Description: bodyJSON.Description,
	}

	if creation, data.SessionData, err = w.BeginRegistration(user, opts...); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to create %s registration challenge for user '%s'", regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	userSession.WebAuthn = &data

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "registration challenge", regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if err = ctx.SetJSONBody(creation); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrWriteResponseBody, regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}
}

// WebAuthnRegistrationPOST processes the attestation challenge response from the client.
func WebAuthnRegistrationPOST(ctx *middlewares.AutheliaCtx) {
	var (
		err  error
		w    *webauthn.WebAuthn
		user *model.WebAuthnUser

		userSession session.UserSession

		response *protocol.ParsedCredentialCreationData

		credential *webauthn.Credential
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration response", regulation.AuthTypeWebAuthn)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if userSession.WebAuthn == nil || userSession.WebAuthn.SessionData == nil {
		ctx.Logger.Errorf("WebAuthn session data is not present in order to handle %s registration for user '%s'. This could indicate a user trying to POST to the wrong endpoint, or the session data is not present for the browser they used.", regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if w, err = newWebAuthn(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to configure %s during registration for user '%s'", regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if response, err = protocol.ParseCredentialCreationResponseBody(bytes.NewReader(ctx.PostBody())); err != nil {
		var e *protocol.Error

		switch {
		case errors.As(err, &e):
			ctx.Logger.WithError(e).Errorf("Unable to parse %s registration for user '%s': %+v (%s)", regulation.AuthTypeWebAuthn, userSession.Username, err, e.DevInfo)
		default:
			ctx.Logger.WithError(err).Errorf("Unable to parse %s registration for user '%s'", regulation.AuthTypeWebAuthn, userSession.Username)
		}

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if user, err = getWebAuthnUserByRPID(ctx, userSession.Username, userSession.DisplayName, w.Config.RPID); err != nil {
		ctx.Logger.Errorf("Unable to load %s user details for registration for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if credential, err = w.CreateCredential(user, *userSession.WebAuthn.SessionData, response); err != nil {
		var e *protocol.Error

		switch {
		case errors.As(err, &e):
			ctx.Logger.WithError(e).Errorf("Unable to create %s credential for user '%s': %s", regulation.AuthTypeWebAuthn, userSession.Username, e.DevInfo)
		default:
			ctx.Logger.WithError(err).Errorf("Unable to create %s credential for user '%s'", regulation.AuthTypeWebAuthn, userSession.Username)
		}

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	device := model.NewWebAuthnCredential(w.Config.RPID, userSession.Username, userSession.WebAuthn.Description, credential)

	device.Discoverable = webauthnCredentialCreationIsDiscoverable(ctx, response)

	if err = ctx.Providers.StorageProvider.SaveWebAuthnCredential(ctx, device); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to save %s device registration for user '%s'", regulation.AuthTypeWebAuthn, userSession.Username)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	userSession.WebAuthn = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "removal of the registration challenge", regulation.AuthTypeWebAuthn, userSession.Username)
	}

	ctx.ReplyOK()
	ctx.SetStatusCode(fasthttp.StatusCreated)

	ctxLogEvent(ctx, userSession.Username, eventLogAction2FAAdded, map[string]any{eventLogKeyAction: eventLogAction2FAAdded, eventLogKeyCategory: eventLogCategoryWebAuthnCredential, eventLogKeyDescription: device.Description})
}

// WebAuthnRegistrationDELETE deletes any active WebAuthn registration session..
func WebAuthnRegistrationDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration deletion", regulation.AuthTypeWebAuthn)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.WebAuthn != nil {
		userSession.WebAuthn = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred attempting to save the updated session while attempting to delete the WebAuthn registration data")

			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetJSONError(messageOperationFailed)

			return
		}
	}

	ctx.ReplyOK()
}
