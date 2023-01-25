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

// WebauthnIdentityStart the handler for initiating the identity validation.
var WebauthnIdentityStart = middlewares.IdentityVerificationStart(
	middlewares.IdentityVerificationStartArgs{
		MailTitle:             "Register your key",
		MailButtonContent:     "Register",
		TargetEndpoint:        "/webauthn/register",
		ActionClaim:           ActionWebauthnRegistration,
		IdentityRetrieverFunc: identityRetrieverFromSession,
	}, nil)

// WebauthnIdentityFinish the handler for finishing the identity validation.
var WebauthnIdentityFinish = middlewares.IdentityVerificationFinish(
	middlewares.IdentityVerificationFinishArgs{
		ActionClaim:          ActionWebauthnRegistration,
		IsTokenUserValidFunc: isTokenUserValidFor2FARegistration,
	}, SecondFactorWebauthnAttestationGET)

// SecondFactorWebauthnAttestationGET returns the attestation challenge from the server.
func SecondFactorWebauthnAttestationGET(ctx *middlewares.AutheliaCtx, _ string) {
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

// WebauthnAttestationPOST processes the attestation challenge response from the client.
func WebauthnAttestationPOST(ctx *middlewares.AutheliaCtx) {
	type requestPostData struct {
		Credential  json.RawMessage `json:"credential"`
		Description string          `json:"description"`
	}

	var (
		err  error
		w    *webauthn.WebAuthn
		user *model.WebauthnUser

		userSession session.UserSession

		attestationResponse *protocol.ParsedCredentialCreationData
		credential          *webauthn.Credential
		postData            *requestPostData
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s attestation response", regulation.AuthTypeWebauthn)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if userSession.Webauthn == nil {
		ctx.Logger.Errorf("Webauthn session data is not present in order to handle attestation for user '%s'. This could indicate a user trying to POST to the wrong endpoint, or the session data is not present for the browser they used.", userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if w, err = newWebauthn(ctx); err != nil {
		ctx.Logger.Errorf("Unable to configure %s during assertion challenge for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	err = json.Unmarshal(ctx.PostBody(), &postData)
	if err != nil {
		ctx.Logger.Errorf("Unable to parse %s assertion request data for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if attestationResponse, err = protocol.ParseCredentialCreationResponseBody(bytes.NewReader(postData.Credential)); err != nil {
		ctx.Logger.Errorf("Unable to parse %s assertion for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if user, err = getWebAuthnUser(ctx, userSession); err != nil {
		ctx.Logger.Errorf("Unable to load %s devices for assertion challenge for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if credential, err = w.CreateCredential(user, *userSession.Webauthn, attestationResponse); err != nil {
		ctx.Logger.Errorf("Unable to load %s devices for assertion challenge for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	devices, err := ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, userSession.Username)
	if err != nil && err != storage.ErrNoWebauthnDevice {
		ctx.Logger.Errorf("Unable to load existing %s devices for for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	for _, existingDevice := range devices {
		if existingDevice.Description == postData.Description {
			ctx.Logger.Errorf("%s device for for user '%s' with name '%s' already exists", regulation.AuthTypeWebauthn, userSession.Username, postData.Description)

			respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)
			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetJSONError(messageSecurityKeyDuplicateName)

			return
		}
	}

	device := model.NewWebauthnDeviceFromCredential(w.Config.RPID, userSession.Username, postData.Description, credential)

	if err = ctx.Providers.StorageProvider.SaveWebauthnDevice(ctx, device); err != nil {
		ctx.Logger.Errorf("Unable to load %s devices for assertion challenge for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	userSession.Webauthn = nil
	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "removal of the attestation challenge", regulation.AuthTypeWebauthn, userSession.Username, err)
	}

	ctx.ReplyOK()
	ctx.SetStatusCode(fasthttp.StatusCreated)

	ctxLogEvent(ctx, userSession.Username, "Second Factor Method Added", map[string]any{"Action": "Second Factor Method Added", "Category": "Webauthn Credential", "Credential Description": postData.Description})
}
