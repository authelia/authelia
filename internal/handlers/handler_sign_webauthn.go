package handlers

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
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
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred generating a WebAuthn authentication challenge")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	var origin *url.URL

	if origin, err = ctx.GetOrigin(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': error occurred provisioning the configuration", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if w, err = handleNewWebAuthn(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': error occurred provisioning the configuration", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	rpid := origin.Hostname()

	if user, err = handleGetWebAuthnUserByRPID(ctx, userSession.Username, userSession.DisplayName, rpid); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': error occurred retrieving the WebAuthn user configuration from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	extensions := map[string]any{}

	if user.HasFIDOU2F() {
		extensions["appid"] = w.Config.RPOrigins[0]
	}

	var opts = []webauthn.LoginOption{
		webauthn.WithAllowedCredentials(user.WebAuthnCredentialDescriptors()),
		webauthn.WithLoginRelyingPartyID(rpid),
	}

	if len(extensions) != 0 {
		opts = append(opts, webauthn.WithAssertionExtensions(extensions))
	}

	var (
		assertion *protocol.CredentialAssertion
		data      session.WebAuthn
	)

	if assertion, data.SessionData, err = w.BeginLogin(user, opts...); err != nil {
		ctx.Logger.WithError(formatWebAuthnError(err)).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': error occurred starting the authentication session", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	userSession.WebAuthn = &data

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': %s", userSession.Username, errStrUserSessionDataSave)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.SetJSONBody(assertion); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': %s", userSession.Username, errStrRespBody)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

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

		w    *webauthn.WebAuthn
		c    *webauthn.Credential
		user *model.WebAuthnUser

		bodyJSON bodySignWebAuthnRequest

		response *protocol.ParsedCredentialAssertionData
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn authentication challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred validating a WebAuthn authentication challenge")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': %s", userSession.Username, errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if response, err = protocol.ParseCredentialRequestResponseBody(bytes.NewReader(bodyJSON.Response)); err != nil {
		ctx.Logger.WithError(formatWebAuthnError(err)).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': %s", userSession.Username, errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession.WebAuthn == nil || userSession.WebAuthn.SessionData == nil {
		ctx.Logger.WithError(fmt.Errorf("challenge session data is not present")).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': %s", userSession.Username, errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if w, err = handleNewWebAuthn(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': error occurred provisioning the configuration", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if user, err = handleGetWebAuthnUserByRPID(ctx, userSession.Username, userSession.DisplayName, w.Config.RPID); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': error occurred retrieving the WebAuthn user configuration from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if c, err = w.ValidateLogin(user, *userSession.WebAuthn.SessionData, response); err != nil {
		_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypeWebAuthn, formatWebAuthnError(err))

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	defer func() {
		userSession.WebAuthn = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': %s", userSession.Username, errStrUserSessionDataSave)
		}
	}()

	var found bool

	for _, credential := range user.Credentials {
		if bytes.Equal(credential.KID.Bytes(), c.ID) {
			credential.UpdateSignInInfo(w.Config, ctx.Clock.Now().UTC(), c.Authenticator)

			found = true

			if err = ctx.Providers.StorageProvider.UpdateWebAuthnCredentialSignIn(ctx, credential); err != nil {
				ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': error occurred saving the credential sign-in information to the storage backend", userSession.Username)

				ctx.SetStatusCode(fasthttp.StatusForbidden)
				ctx.SetJSONError(messageMFAValidationFailed)

				return
			}

			break
		}
	}

	if !found {
		ctx.Logger.WithError(fmt.Errorf("credential was not found")).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': error occurred saving the credential sign-in information to storage", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if c.Authenticator.CloneWarning {
		ctx.Logger.WithError(fmt.Errorf("authenticator sign count indicates that it is cloned")).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': error occurred validating the authenticator response", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.RegenerateSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': error regenerating the user session", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = markAuthenticationAttempt(ctx, true, nil, userSession.Username, regulation.AuthTypeWebAuthn, nil); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	userSession.SetTwoFactorWebAuthn(ctx.Clock.Now(),
		response.ParsedPublicKeyCredential.AuthenticatorAttachment == protocol.CrossPlatform,
		response.Response.AuthenticatorData.Flags.HasUserPresent(),
		response.Response.AuthenticatorData.Flags.HasUserVerified())

	if bodyJSON.Workflow == workflowOpenIDConnect {
		handleOIDCWorkflowResponse(ctx, &userSession, bodyJSON.TargetURL, bodyJSON.WorkflowID)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}

// WebAuthnPasskeyAssertionGET handler starts the passkey assertion ceremony.
func WebAuthnPasskeyAssertionGET(ctx *middlewares.AutheliaCtx) {
	var (
		w           *webauthn.WebAuthn
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn passkey authentication challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if !userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserIsAlreadyAuthenticated).Errorf("Error occurred generating a WebAuthn passkey authentication challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)
	}

	if w, err = handleNewWebAuthn(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge: error occurred provisioning the configuration")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	extensions := map[string]any{}

	var opts []webauthn.LoginOption

	if len(extensions) != 0 {
		opts = append(opts, webauthn.WithAssertionExtensions(extensions))
	}

	var (
		assertion *protocol.CredentialAssertion
		data      session.WebAuthn
	)

	if assertion, data.SessionData, err = w.BeginDiscoverableLogin(opts...); err != nil {
		ctx.Logger.WithError(formatWebAuthnError(err)).Error("Error occurred generating a WebAuthn passkey authentication challenge: error occurred starting the authentication session")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	userSession.WebAuthn = &data

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn passkey authentication challenge: %s", errStrUserSessionDataSave)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.SetJSONBody(assertion); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn passkey authentication challenge: %s", errStrRespBody)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}
}

// WebAuthnPasskeyAssertionPOST handler completes the assertion ceremony after verifying the challenge.
//
//nolint:gocyclo
func WebAuthnPasskeyAssertionPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession

		err error

		w *webauthn.WebAuthn
		u webauthn.User
		c *webauthn.Credential

		bodyJSON bodySignWebAuthnRequest

		response *protocol.ParsedCredentialAssertionData
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if !userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserIsAlreadyAuthenticated).Error("Error occurred validating a WebAuthn passkey authentication challenge")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: %s", errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if response, err = protocol.ParseCredentialRequestResponseBody(bytes.NewReader(bodyJSON.Response)); err != nil {
		ctx.Logger.WithError(formatWebAuthnError(err)).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: %s", errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession.WebAuthn == nil || userSession.WebAuthn.SessionData == nil {
		ctx.Logger.WithError(fmt.Errorf("challenge session data is not present")).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if w, err = handleNewWebAuthn(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: error occurred provisioning the configuration")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if u, c, err = w.ValidatePasskeyLogin(handlerWebAuthnDiscoverableLogin(ctx, w.Config.RPID), *userSession.WebAuthn.SessionData, response); err != nil {
		_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypeWebAuthn, formatWebAuthnError(err))

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	defer func() {
		userSession.WebAuthn = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': %s", u.WebAuthnName(), errStrUserSessionDataSave)
		}
	}()

	var (
		details *authentication.UserDetails
		user    *model.WebAuthnUser
		ok      bool
	)

	if user, ok = u.(*model.WebAuthnUser); !ok {
		ctx.Logger.Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': the user object was not of the correct type", u.WebAuthnName())
	}

	ok = false

	for _, credential := range user.Credentials {
		if bytes.Equal(credential.KID.Bytes(), c.ID) {
			credential.UpdateSignInInfo(w.Config, ctx.Clock.Now().UTC(), c.Authenticator)

			ok = true

			if err = ctx.Providers.StorageProvider.UpdateWebAuthnCredentialSignIn(ctx, credential); err != nil {
				ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error occurred saving the credential sign-in information to the storage backend", u.WebAuthnName())

				ctx.SetStatusCode(fasthttp.StatusForbidden)
				ctx.SetJSONError(messageMFAValidationFailed)

				return
			}

			break
		}
	}

	if !ok {
		ctx.Logger.WithError(fmt.Errorf("credential was not found")).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error occurred saving the credential sign-in information to storage", u.WebAuthnName())

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if c.Authenticator.CloneWarning {
		ctx.Logger.WithError(fmt.Errorf("authenticator sign count indicates that it is cloned")).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error occurred validating the authenticator response", u.WebAuthnName())

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if details, err = ctx.Providers.UserProvider.GetDetails(user.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error retreiving user details", u.WebAuthnName())

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.RegenerateSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error regenerating the user session", u.WebAuthnName())

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = markAuthenticationAttempt(ctx, true, nil, userSession.Username, regulation.AuthTypeWebAuthn, nil); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	userSession.SetOneFactor(ctx.Clock.Now(), details, false)
	userSession.SetTwoFactorWebAuthn(ctx.Clock.Now(),
		response.ParsedPublicKeyCredential.AuthenticatorAttachment == protocol.CrossPlatform,
		response.Response.AuthenticatorData.Flags.HasUserPresent(),
		response.Response.AuthenticatorData.Flags.HasUserVerified())

	if bodyJSON.Workflow == workflowOpenIDConnect {
		handleOIDCWorkflowResponse(ctx, &userSession, bodyJSON.TargetURL, bodyJSON.WorkflowID)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}
