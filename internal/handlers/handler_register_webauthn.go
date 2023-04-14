package handlers

import (
	"bytes"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

// WebauthnIdentityStart the handler for initiating the identity validation.
var WebauthnIdentityStart = middlewares.IdentityVerificationStart(middlewares.IdentityVerificationStartArgs{
	MailTitle:             "Register your key",
	MailButtonContent:     "Register",
	TargetEndpoint:        "/webauthn/register",
	ActionClaim:           ActionWebAuthnRegistration,
	IdentityRetrieverFunc: identityRetrieverFromSession,
}, nil)

// WebauthnIdentityFinish the handler for finishing the identity validation.
var WebauthnIdentityFinish = middlewares.IdentityVerificationFinish(
	middlewares.IdentityVerificationFinishArgs{
		ActionClaim:          ActionWebAuthnRegistration,
		IsTokenUserValidFunc: isTokenUserValidFor2FARegistration,
	}, SecondFactorWebAuthnAttestationGET)

// SecondFactorWebAuthnAttestationGET returns the attestation challenge from the server.
func SecondFactorWebAuthnAttestationGET(ctx *middlewares.AutheliaCtx, _ string) {
	var (
		w           *webauthn.WebAuthn
		user        *model.WebAuthnUser
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s attestation challenge", regulation.AuthTypeWebAuthn)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if w, err = newWebAuthn(ctx); err != nil {
		ctx.Logger.Errorf("Unable to create %s attestation challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if user, err = getWebAuthnUser(ctx, userSession); err != nil {
		ctx.Logger.Errorf("Unable to load %s devices for assertion challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	var credentialCreation *protocol.CredentialCreation

	if credentialCreation, userSession.WebAuthn, err = w.BeginRegistration(user); err != nil {
		ctx.Logger.Errorf("Unable to create %s attestation challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "attestation challenge", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if err = ctx.SetJSONBody(credentialCreation); err != nil {
		ctx.Logger.Errorf(logFmtErrWriteResponseBody, regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}
}

// WebAuthnAttestationPOST processes the attestation challenge response from the client.
func WebAuthnAttestationPOST(ctx *middlewares.AutheliaCtx) {
	var (
		err  error
		w    *webauthn.WebAuthn
		user *model.WebAuthnUser

		userSession session.UserSession

		attestationResponse *protocol.ParsedCredentialCreationData
		credential          *webauthn.Credential
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s attestation response", regulation.AuthTypeWebAuthn)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if userSession.WebAuthn == nil {
		ctx.Logger.Errorf("WebAuthn session data is not present in order to handle attestation for user '%s'. This could indicate a user trying to POST to the wrong endpoint, or the session data is not present for the browser they used.", userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if w, err = newWebAuthn(ctx); err != nil {
		ctx.Logger.Errorf("Unable to configure %s during assertion challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageUnableToRegisterSecurityKey)

		return
	}

	if attestationResponse, err = protocol.ParseCredentialCreationResponseBody(bytes.NewReader(ctx.PostBody())); err != nil {
		ctx.Logger.Errorf("Unable to parse %s assertionfor user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if user, err = getWebAuthnUser(ctx, userSession); err != nil {
		ctx.Logger.Errorf("Unable to load %s devices for assertion challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if credential, err = w.CreateCredential(user, *userSession.WebAuthn, attestationResponse); err != nil {
		ctx.Logger.Errorf("Unable to load %s devices for assertion challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	device := model.NewWebAuthnDeviceFromCredential(w.Config.RPID, userSession.Username, "Primary", credential)

	if err = ctx.Providers.StorageProvider.SaveWebAuthnDevice(ctx, device); err != nil {
		ctx.Logger.Errorf("Unable to load %s devices for assertion challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	userSession.WebAuthn = nil
	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "removal of the attestation challenge", regulation.AuthTypeWebAuthn, userSession.Username, err)
	}

	ctx.ReplyOK()
	ctx.SetStatusCode(fasthttp.StatusCreated)

	ctxLogEvent(ctx, userSession.Username, "Second Factor Method Added", map[string]any{"Action": "Second Factor Method Added", "Category": "WebAuthn Credential", "Device Name": "Primary"})
}
