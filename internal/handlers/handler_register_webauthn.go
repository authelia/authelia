package handlers

import (
	"bytes"
	"encoding/json"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

// WebauthnRegistrationGET returns the attestation challenge from the server.
func WebauthnRegistrationGET(ctx *middlewares.AutheliaCtx) {
	var (
		w           *webauthn.WebAuthn
		user        *model.WebauthnUser
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s attestation challenge", regulation.AuthTypeWebauthn)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if w, err = newWebauthn(ctx); err != nil {
		ctx.Logger.Errorf("Unable to create %s attestation challenge for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if user, err = getWebAuthnUser(ctx, userSession); err != nil {
		ctx.Logger.Errorf("Unable to load %s devices for assertion challenge for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	var credentialCreation *protocol.CredentialCreation

	if credentialCreation, userSession.Webauthn, err = w.BeginRegistration(user, webauthn.WithExclusions(user.WebAuthnCredentialDescriptors())); err != nil {
		ctx.Logger.Errorf("Unable to create %s attestation challenge for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "attestation challenge", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if err = ctx.SetJSONBody(credentialCreation); err != nil {
		ctx.Logger.Errorf(logFmtErrWriteResponseBody, regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}
}

// WebauthnRegistrationPOST processes the attestation challenge response from the client.
func WebauthnRegistrationPOST(ctx *middlewares.AutheliaCtx) {
	var (
		err  error
		w    *webauthn.WebAuthn
		user *model.WebauthnUser

		userSession session.UserSession

		response *protocol.ParsedCredentialCreationData

		credential *webauthn.Credential
		bodyJSON   bodyRegisterWebauthnRequest
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration response", regulation.AuthTypeWebauthn)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if userSession.Webauthn == nil {
		ctx.Logger.Errorf("Webauthn session data is not present in order to handle %s registration for user '%s'. This could indicate a user trying to POST to the wrong endpoint, or the session data is not present for the browser they used.", regulation.AuthTypeWebauthn, userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if w, err = newWebauthn(ctx); err != nil {
		ctx.Logger.Errorf("Unable to configure %s during registration for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.Errorf("Unable to parse %s registration request POST data for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if response, err = protocol.ParseCredentialCreationResponseBody(bytes.NewReader(bodyJSON.Response)); err != nil {
		ctx.Logger.Errorf("Unable to parse %s registration for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	ctx.Logger.WithField("att_format", response.Response.AttestationObject.Format).Debug("Response Data")

	if user, err = getWebAuthnUser(ctx, userSession); err != nil {
		ctx.Logger.Errorf("Unable to load %s user details for registration for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if credential, err = w.CreateCredential(user, *userSession.Webauthn, response); err != nil {
		ctx.Logger.Errorf("Unable to create %s credential for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	ctx.Logger.WithField("att_type", credential.AttestationType).Debug("Credential Data")

	devices, err := ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, userSession.Username)
	if err != nil && err != storage.ErrNoWebauthnDevice {
		ctx.Logger.Errorf("Unable to load existing %s devices for for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	for _, existingDevice := range devices {
		if existingDevice.Description == bodyJSON.Description {
			ctx.Logger.Errorf("%s device for for user '%s' with name '%s' already exists", regulation.AuthTypeWebauthn, userSession.Username, bodyJSON.Description)

			respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)
			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetJSONError(messageSecurityKeyDuplicateName)

			return
		}
	}

	device := model.NewWebauthnDeviceFromCredential(w.Config.RPID, userSession.Username, bodyJSON.Description, credential)

	if err = ctx.Providers.StorageProvider.SaveWebauthnDevice(ctx, device); err != nil {
		ctx.Logger.Errorf("Unable to save %s device registration for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	userSession.Webauthn = nil
	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "removal of the registration challenge", regulation.AuthTypeWebauthn, userSession.Username, err)
	}

	ctx.ReplyOK()
	ctx.SetStatusCode(fasthttp.StatusCreated)

	ctxLogEvent(ctx, userSession.Username, "Second Factor Method Added", map[string]any{"Action": "Second Factor Method Added", "Category": "Webauthn Credential", "Credential Description": bodyJSON.Description})
}
