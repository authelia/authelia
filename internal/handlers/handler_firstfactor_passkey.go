package handlers

import (
	"bytes"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	iwebauthn "github.com/authelia/authelia/v4/internal/webauthn"
)

// FirstFactorPasskeyGET handler starts the passkey assertion ceremony.
func FirstFactorPasskeyGET(ctx *middlewares.AutheliaCtx) {
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

	if w, err = ctx.GetWebAuthnProvider(); err != nil {
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
		ctx.Logger.WithError(iwebauthn.FormatError(err)).Error("Error occurred generating a WebAuthn passkey authentication challenge: error occurred starting the authentication session")

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

// FirstFactorPasskeyPOST handler completes the assertion ceremony after verifying the challenge.
//
//nolint:gocyclo
func FirstFactorPasskeyPOST(ctx *middlewares.AutheliaCtx) {
	var (
		provider    *session.Session
		userSession session.UserSession

		err error

		w *webauthn.WebAuthn
		u webauthn.User
		c *webauthn.Credential

		bodyJSON bodySignPasskeyRequest

		response *protocol.ParsedCredentialAssertionData
	)

	if provider, err = ctx.GetSessionProvider(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession, err = provider.GetSession(ctx.RequestCtx); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: %s", errStrUserSessionData)

		return
	}

	if !userSession.IsAnonymous() {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(errUserIsAlreadyAuthenticated).Error("Error occurred validating a WebAuthn passkey authentication challenge")

		return
	}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: %s", errStrReqBodyParse)

		return
	}

	if response, err = protocol.ParseCredentialRequestResponseBody(bytes.NewReader(bodyJSON.Response)); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: %s", errStrReqBodyParse)

		return
	}

	if userSession.WebAuthn == nil || userSession.WebAuthn.SessionData == nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(fmt.Errorf("challenge session data is not present")).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: %s", errStrUserSessionData)

		return
	}

	if w, err = ctx.GetWebAuthnProvider(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge: error occurred provisioning the configuration")

		return
	}

	if u, c, err = w.ValidatePasskeyLogin(handlerWebAuthnDiscoverableLogin(ctx, w.Config.RPID), *userSession.WebAuthn.SessionData, response); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypePasskey, iwebauthn.FormatError(err))

		return
	}

	var (
		details *authentication.UserDetails
		user    *model.WebAuthnUser
		ok      bool
	)

	if user, ok = u.(*model.WebAuthnUser); !ok {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': the user object was not of the correct type", u.WebAuthnName())

		return
	}

	ok = false

	for _, credential := range user.Credentials {
		if bytes.Equal(credential.KID.Bytes(), c.ID) {
			credential.UpdateSignInInfo(w.Config, ctx.Clock.Now().UTC(), c.Authenticator)

			ok = true

			if err = ctx.Providers.StorageProvider.UpdateWebAuthnCredentialSignIn(ctx, credential); err != nil {
				ctx.SetStatusCode(fasthttp.StatusForbidden)
				ctx.SetJSONError(messageMFAValidationFailed)

				ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error occurred saving the credential sign-in information to the storage backend", u.WebAuthnName())

				return
			}

			break
		}
	}

	if !ok {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(fmt.Errorf("credential was not found")).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error occurred saving the credential sign-in information to storage", u.WebAuthnName())

		return
	}

	if c.Authenticator.CloneWarning {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(fmt.Errorf("authenticator sign count indicates that it is cloned")).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error occurred validating the authenticator response", u.WebAuthnName())

		return
	}

	if details, err = ctx.Providers.UserProvider.GetDetails(user.Username); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error retreiving user details", u.WebAuthnName())

		return
	}

	if err = ctx.RegenerateSession(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error regenerating the user session", details.Username)

		return
	}

	if err = markAuthenticationAttempt(ctx, true, nil, userSession.Username, regulation.AuthTypePasskey, nil); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn passkey authentication challenge for user '%s': error occurred recording the authentication attempt", details.Username)

		return
	}

	if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
		userSession.RefreshTTL = ctx.Clock.Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
	}

	// Check if bodyJSON.KeepMeLoggedIn can be deref'd and derive the value based on the configuration and JSON data.
	keepMeLoggedIn := !provider.Config.DisableRememberMe && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

	// Set the cookie to expire if remember me is enabled and the user has asked us to.
	if keepMeLoggedIn {
		err = provider.UpdateExpiration(ctx.RequestCtx, provider.Config.RememberMe)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetJSONError(messageMFAValidationFailed)

			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated expiration", regulation.AuthTypePasskey, logFmtActionAuthentication, details.Username)

			return
		}
	}

	userSession.SetOneFactorPasskey(
		ctx.Clock.Now(), details,
		false,
		response.ParsedPublicKeyCredential.AuthenticatorAttachment == protocol.CrossPlatform,
		response.Response.AuthenticatorData.Flags.HasUserPresent(),
		response.Response.AuthenticatorData.Flags.HasUserVerified(),
	)

	if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
		userSession.RefreshTTL = ctx.Clock.Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated profile", regulation.AuthTypePasskey, logFmtActionAuthentication, details.Username)

		return
	}

	if bodyJSON.Workflow == workflowOpenIDConnect {
		handleOIDCWorkflowResponse(ctx, &userSession, bodyJSON.TargetURL, bodyJSON.WorkflowID)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}
