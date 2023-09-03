package handlers

import (
	"bytes"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

// WebAuthnAssertionGET handler starts the assertion ceremony.
func WebAuthnAssertionGET(ctx *middlewares.AutheliaCtx) {
	var (
		w           *webauthn.WebAuthn
		user        *model.WebAuthnUser
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if w, err = newWebAuthn(ctx); err != nil {
		ctx.Logger.Errorf("Unable to configure %s during authentication challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if user, err = getWebAuthnUserByRPID(ctx, userSession.Username, userSession.DisplayName, w.Config.RPID); err != nil {
		ctx.Logger.Errorf("Unable to load %s user details during authentication challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	extensions := map[string]any{}

	if user.HasFIDOU2F() {
		extensions["appid"] = w.Config.RPOrigins[0]
	}

	var opts = []webauthn.LoginOption{
		webauthn.WithAllowedCredentials(user.WebAuthnCredentialDescriptors()),
	}

	if len(extensions) != 0 {
		opts = append(opts, webauthn.WithAssertionExtensions(extensions))
	}

	var (
		assertion *protocol.CredentialAssertion
		data      session.WebAuthn
	)

	if assertion, data.SessionData, err = w.BeginLogin(user, opts...); err != nil {
		ctx.Logger.Errorf("Unable to create %s authentication challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	userSession.WebAuthn = &data

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "assertion challenge", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if err = ctx.SetJSONBody(assertion); err != nil {
		ctx.Logger.Errorf(logFmtErrWriteResponseBody, regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}
}

// WebAuthnAssertionPOST handler completes the assertion ceremony after verifying the challenge.
//
//nolint:gocyclo
func WebAuthnAssertionPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession

		err error
		w   *webauthn.WebAuthn

		bodyJSON bodySignWebAuthnRequest
	)

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.Errorf(logFmtErrParseRequestBody, regulation.AuthTypeWebAuthn, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if userSession.WebAuthn == nil || userSession.WebAuthn.SessionData == nil {
		ctx.Logger.Errorf("WebAuthn session data is not present in order to handle authentication challenge for user '%s'. This could indicate a user trying to POST to the wrong endpoint, or the session data is not present for the browser they used.", userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if w, err = newWebAuthn(ctx); err != nil {
		ctx.Logger.Errorf("Unable to configure %s during authentication challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	var (
		assertionResponse *protocol.ParsedCredentialAssertionData
		credential        *webauthn.Credential
		user              *model.WebAuthnUser
	)

	if assertionResponse, err = protocol.ParseCredentialRequestResponseBody(bytes.NewReader(bodyJSON.Response)); err != nil {
		ctx.Logger.Errorf("Unable to parse %s authentication challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if user, err = getWebAuthnUserByRPID(ctx, userSession.Username, userSession.DisplayName, w.Config.RPID); err != nil {
		ctx.Logger.Errorf("Unable to load %s credentials for authentication challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if credential, err = w.ValidateLogin(user, *userSession.WebAuthn.SessionData, assertionResponse); err != nil {
		_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypeWebAuthn, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	var found bool

	for _, device := range user.Credentials {
		if bytes.Equal(device.KID.Bytes(), credential.ID) {
			device.UpdateSignInInfo(w.Config, ctx.Clock.Now(), credential.Authenticator.SignCount)

			found = true

			if err = ctx.Providers.StorageProvider.UpdateWebAuthnCredentialSignIn(ctx, device); err != nil {
				ctx.Logger.Errorf("Unable to save %s device signin count for authentication challenge for user '%s': %+v", regulation.AuthTypeWebAuthn, userSession.Username, err)

				respondUnauthorized(ctx, messageMFAValidationFailed)

				return
			}

			break
		}
	}

	if !found {
		ctx.Logger.Errorf("Unable to save %s device signin count for authentication challenge for user '%s' device '%x' count '%d': unable to find device", regulation.AuthTypeWebAuthn, userSession.Username, credential.ID, credential.Authenticator.SignCount)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if err = ctx.RegenerateSession(); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionRegenerate, regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if err = markAuthenticationAttempt(ctx, true, nil, userSession.Username, regulation.AuthTypeWebAuthn, nil); err != nil {
		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	userSession.SetTwoFactorWebAuthn(ctx.Clock.Now(),
		assertionResponse.Response.AuthenticatorData.Flags.HasUserPresent(),
		assertionResponse.Response.AuthenticatorData.Flags.HasUserVerified())

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "removal of the authentiation challenge and authentication time", regulation.AuthTypeWebAuthn, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if bodyJSON.Workflow == workflowOpenIDConnect {
		handleOIDCWorkflowResponse(ctx, bodyJSON.TargetURL, bodyJSON.WorkflowID)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}
