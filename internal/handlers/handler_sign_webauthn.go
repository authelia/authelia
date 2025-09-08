package handlers

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	iwebauthn "github.com/authelia/authelia/v4/internal/webauthn"
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
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge: %s", errStrUserSessionData)

		return
	}

	if userSession.IsAnonymous() {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred generating a WebAuthn authentication challenge")

		return
	}

	var origin *url.URL

	if origin, err = ctx.GetOrigin(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': error occurred provisioning the configuration", userSession.Username)

		return
	}

	if w, err = ctx.GetWebAuthnProvider(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': error occurred provisioning the configuration", userSession.Username)

		return
	}

	rpid := origin.Hostname()

	if user, err = handleGetWebAuthnUserByRPID(ctx, userSession.Username, userSession.DisplayName, rpid); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': error occurred retrieving the WebAuthn user configuration from the storage backend", userSession.Username)

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
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': error occurred starting the authentication session", userSession.Username)

		return
	}

	userSession.WebAuthn = &data

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': %s", userSession.Username, errStrUserSessionDataSave)

		return
	}

	if err = ctx.SetJSONBody(assertion); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn authentication challenge for user '%s': %s", userSession.Username, errStrRespBody)

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
		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf("Error occurred validating a WebAuthn authentication challenge for user '%s': %s", userSession.Username, errStrReqBodyParse)

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

	if w, err = ctx.GetWebAuthnProvider(); err != nil {
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
		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeWebAuthn, iwebauthn.FormatError(err))

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
			credential.UpdateSignInInfo(w.Config, ctx.GetClock().Now().UTC(), c.Authenticator)

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

	doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeWebAuthn, nil)

	userSession.SetTwoFactorWebAuthn(ctx.GetClock().Now().UTC(),
		response.AuthenticatorAttachment == protocol.CrossPlatform,
		response.Response.AuthenticatorData.Flags.HasUserPresent(),
		response.Response.AuthenticatorData.Flags.HasUserVerified())

	if len(bodyJSON.Flow) > 0 {
		handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}
