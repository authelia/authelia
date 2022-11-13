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
)

// WebauthnIdentityStart the handler for initiating the identity validation.
var WebauthnIdentityStart = middlewares.IdentityVerificationStart(middlewares.IdentityVerificationStartArgs{
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
		w    *webauthn.WebAuthn
		user *model.WebauthnUser
		err  error
	)

	userSession := ctx.GetSession()

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

	if credentialCreation, userSession.Webauthn, err = w.BeginRegistration(user); err != nil {
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

		attestationResponse *protocol.ParsedCredentialCreationData
		credential          *webauthn.Credential
		postData            *requestPostData
	)

	userSession := ctx.GetSession()

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

	device := model.NewWebauthnDeviceFromCredential(w.Config.RPID, userSession.Username, postData.Description, credential)

	if err = ctx.Providers.StorageProvider.SaveWebauthnDevice(ctx, device); err != nil {
		ctx.Logger.Errorf("Unable to load %s devices for assertion challenge for user '%s': %+v", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	userSession.Webauthn = nil
	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "removal of the attestation challenge", regulation.AuthTypeWebauthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	ctx.ReplyOK()
	ctx.SetStatusCode(fasthttp.StatusCreated)
}
