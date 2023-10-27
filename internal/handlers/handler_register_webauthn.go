package handlers

import (
	"bytes"
	"encoding/json"
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
		ctx.Logger.Errorf("Unable to create provider to generate %s registration challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.Errorf("Unable to parse %s registration request PUT data for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

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
		ctx.Logger.Errorf("Unable to load existing %s devices for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

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
		ctx.Logger.Errorf("Unable to load %s devices for registration challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

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
		ctx.Logger.Errorf("Unable to create %s registration challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	userSession.WebAuthn = &data

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "registration challenge", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if err = ctx.SetJSONBody(creation); err != nil {
		ctx.Logger.Errorf(logFmtErrWriteResponseBody, regulation.AuthTypeWebAuthn, userSession.Username, err)

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
		ctx.Logger.Errorf("Unable to configure %s during registration for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if response, err = protocol.ParseCredentialCreationResponseBody(bytes.NewReader(ctx.PostBody())); err != nil {
		switch e := err.(type) {
		case *protocol.Error:
			ctx.Logger.Errorf("Unable to parse %s registration for user '%s': %+v (%s)", regulation.AuthTypeWebAuthn, userSession.Username, err, e.DevInfo)
		default:
			ctx.Logger.Errorf("Unable to parse %s registration for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)
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
		switch e := err.(type) {
		case *protocol.Error:
			ctx.Logger.Errorf("Unable to create %s credential for user '%s': %+v (%s)", regulation.AuthTypeWebAuthn, userSession.Username, err, e.DevInfo)
		default:
			ctx.Logger.Errorf("Unable to create %s credential for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)
		}

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	device := model.NewWebAuthnCredential(w.Config.RPID, userSession.Username, userSession.WebAuthn.Description, credential)

	device.Discoverable = webauthnCredentialCreationIsDiscoverable(ctx, response)

	if err = ctx.Providers.StorageProvider.SaveWebAuthnCredential(ctx, device); err != nil {
		ctx.Logger.Errorf("Unable to save %s device registration for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	userSession.WebAuthn = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "removal of the registration challenge", regulation.AuthTypeWebAuthn, userSession.Username, err)
	}

	ctx.ReplyOK()
	ctx.SetStatusCode(fasthttp.StatusCreated)

	ctxLogEvent(ctx, userSession.Username, "Second Factor Method Added", map[string]any{"Action": "Second Factor Method Added", "Category": "WebAuthn Credential", "Credential Description": device.Description})
}
