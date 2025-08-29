package handlers

import (
	"bytes"
	"errors"
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
		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeGenerate, errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if !userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserIsAlreadyAuthenticated).Errorf(logFmtErrPasskeyAuthenticationChallengeGenerate, errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if w, err = ctx.GetWebAuthnProvider(); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeGenerate, "error occurred provisioning the configuration")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	var opts []webauthn.LoginOption

	var (
		assertion *protocol.CredentialAssertion
		data      session.WebAuthn
	)

	if assertion, data.SessionData, err = w.BeginDiscoverableLogin(opts...); err != nil {
		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf(logFmtErrPasskeyAuthenticationChallengeGenerate, "error occurred starting the authentication session")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	userSession.WebAuthn = &data

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeGenerate, errStrUserSessionDataSave)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if err = ctx.SetJSONBody(assertion); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeGenerate, errStrRespBody)

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
		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeValidate, errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession, err = provider.GetSession(ctx.RequestCtx); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeValidate, errStrUserSessionData)

		return
	}

	if !userSession.IsAnonymous() {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(errUserIsAlreadyAuthenticated).Error("Error occurred validating a WebAuthn passkey authentication challenge")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	defer func() {
		userSession.WebAuthn = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeValidateUser, userSession.Username, errStrUserSessionDataSave)
		}
	}()

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeValidate, errStrReqBodyParse)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	if response, err = protocol.ParseCredentialRequestResponseBody(bytes.NewReader(bodyJSON.Response)); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf(logFmtErrPasskeyAuthenticationChallengeValidate, errStrReqBodyParse)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	if userSession.WebAuthn == nil || userSession.WebAuthn.SessionData == nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(fmt.Errorf("challenge session data is not present")).Errorf(logFmtErrPasskeyAuthenticationChallengeValidate, errStrUserSessionData)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	if w, err = ctx.GetWebAuthnProvider(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf(logFmtErrPasskeyAuthenticationChallengeValidate, "error occurred provisioning the configuration")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	if u, c, err = w.ValidatePasskeyLogin(handlerWebAuthnDiscoverableLogin(ctx, w.Config.RPID), *userSession.WebAuthn.SessionData, response); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf(logFmtErrPasskeyAuthenticationChallengeValidate, "error performing the login validation")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

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

		ctx.Logger.Errorf(logFmtErrPasskeyAuthenticationChallengeValidateUser, "the user object was not of the correct type", u.WebAuthnName())

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	ok = false

	for _, credential := range user.Credentials {
		if bytes.Equal(credential.KID.Bytes(), c.ID) {
			credential.UpdateSignInInfo(w.Config, ctx.GetClock().Now().UTC(), c.Authenticator)

			if !credential.Discoverable {
				credential.Discoverable = true

				ctx.Logger.WithFields(map[string]any{"kid": credential.KID.String(), "rpid": credential.RPID, "aaguid": credential.AAGUID.UUID.String(), "username": credential.Username, "description": credential.Description}).Debug("WebAuthn Credential Passively Upgraded to a Passkey")
			}

			ok = true

			if err = ctx.Providers.StorageProvider.UpdateWebAuthnCredentialSignIn(ctx, credential); err != nil {
				ctx.SetStatusCode(fasthttp.StatusForbidden)
				ctx.SetJSONError(messageMFAValidationFailed)

				ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeValidateUser, u.WebAuthnName(), "error occurred saving the credential sign-in information to the storage backend")

				doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

				return
			}

			break
		}
	}

	if !ok {
		err = fmt.Errorf("credential was not found")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeValidateUser, u.WebAuthnName(), "error occurred saving the credential sign-in information to storage")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, err)

		return
	}

	if c.Authenticator.CloneWarning {
		err = fmt.Errorf("authenticator sign count indicates that it is cloned")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeValidateUser, u.WebAuthnName(), "error occurred validating the authenticator response")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, err)

		return
	}

	if details, err = ctx.Providers.UserProvider.GetDetails(user.Username); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeValidateUser, u.WebAuthnName(), "error retrieving user details")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	if ban, _, expires, err := ctx.Providers.Regulator.BanCheck(ctx, details.Username); err != nil {
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		ctx.SetJSONError(messageMFAValidationFailed)

		if errors.Is(err, regulation.ErrUserIsBanned) {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(ban, details.Username, expires), regulation.AuthTypePasskey, nil)
		} else {
			ctx.Logger.WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthTypePasskey, details.Username)
		}

		return
	}

	if err = ctx.RegenerateSession(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrPasskeyAuthenticationChallengeValidateUser, details.Username, "error regenerating the user session")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthTypePasskey, nil)

		return
	}

	doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthTypePasskey, nil)

	if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
		userSession.RefreshTTL = ctx.GetClock().Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
	}

	// Check if bodyJSON.KeepMeLoggedIn can be deref'd and derive the value based on the configuration and JSON data.
	keepMeLoggedIn := !provider.Config.DisableRememberMe && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

	// Set the cookie to expire if remember me is enabled and the user has asked us to.
	if keepMeLoggedIn {
		if err = provider.UpdateExpiration(ctx.RequestCtx, provider.Config.RememberMe); err != nil {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetJSONError(messageMFAValidationFailed)

			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated expiration", regulation.AuthTypePasskey, logFmtActionAuthentication, details.Username)

			return
		}
	}

	ctx.Logger.WithFields(map[string]any{
		"hardware": response.ParsedPublicKeyCredential.AuthenticatorAttachment == protocol.CrossPlatform,
		"presence": response.Response.AuthenticatorData.Flags.HasUserPresent(),
		"verified": response.Response.AuthenticatorData.Flags.HasUserVerified(),
	}).Debug("Passkey Login")

	userSession.SetOneFactorPasskey(
		ctx.GetClock().Now(), details,
		keepMeLoggedIn,
		response.AuthenticatorAttachment == protocol.CrossPlatform,
		response.Response.AuthenticatorData.Flags.HasUserPresent(),
		response.Response.AuthenticatorData.Flags.HasUserVerified(),
	)

	if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
		userSession.RefreshTTL = ctx.GetClock().Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
	}

	if len(bodyJSON.Flow) > 0 {
		handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
	} else {
		HandlePasskeyResponse(ctx, bodyJSON.TargetURL, bodyJSON.RequestMethod, userSession.Username, userSession.Groups, userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA) == authentication.TwoFactor)
	}
}
