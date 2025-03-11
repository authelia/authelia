package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	iwebauthn "github.com/authelia/authelia/v4/internal/webauthn"
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
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn registration challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred generating a WebAuthn registration challenge")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn registration challenge for user '%s': %s", userSession.Username, errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if length := len(bodyJSON.Description); length == 0 || length > webauthnCredentialDescriptionMaxLen {
		ctx.Logger.WithError(fmt.Errorf("description has a length of %d but must be between 1 and 64", length)).Errorf("Error occurred generating a WebAuthn registration challenge for user '%s': error occurred validating the description chosen by the user", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if w, err = ctx.GetWebAuthnProvider(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn registration challenge for user '%s': error occurred provisioning the configuration", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if user, err = handleGetWebAuthnUserByRPID(ctx, userSession.Username, userSession.DisplayName, w.Config.RPID); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn registration challenge for user '%s': error occurred retrieving the WebAuthn user configuration from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	for _, credential := range user.Credentials {
		if strings.EqualFold(credential.Description, bodyJSON.Description) {
			ctx.Logger.WithError(fmt.Errorf("the description '%s' already exists for the user", bodyJSON.Description)).Errorf("Error occurred generating a WebAuthn registration challenge for user '%s': error occurred validating the description chosen by the user", userSession.Username)

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetJSONError(messageSecurityKeyDuplicateName)

			return
		}
	}

	var (
		creation *protocol.CredentialCreation
	)

	opts := []webauthn.RegistrationOption{
		webauthn.WithExclusions(user.WebAuthnCredentialDescriptors()),
		webauthn.WithExtensions(map[string]any{"credProps": true}),
	}

	data := session.WebAuthn{
		Description: bodyJSON.Description,
	}

	if creation, data.SessionData, err = w.BeginRegistration(user, opts...); err != nil {
		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf("Error occurred generating a WebAuthn registration challenge for user '%s': error occurred starting the registration session", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	userSession.WebAuthn = &data

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn registration challenge for user '%s': %s", userSession.Username, errStrUserSessionDataSave)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if err = ctx.SetJSONBody(creation); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred generating a WebAuthn registration challenge for user '%s': %s", userSession.Username, errStrRespBody)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

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

		c *webauthn.Credential
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn registration challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred validating a WebAuthn registration challenge")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if userSession.WebAuthn == nil || userSession.WebAuthn.SessionData == nil {
		ctx.Logger.WithError(fmt.Errorf("registration challenge session data is not present")).Errorf("Error occurred validating a WebAuthn registration challenge for user '%s': error occurred retrieving the user session data", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	defer func() {
		userSession.WebAuthn = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn registration challenge for user '%s': %s", userSession.Username, errStrUserSessionDataSave)
		}
	}()

	if response, err = protocol.ParseCredentialCreationResponseBody(bytes.NewReader(ctx.PostBody())); err != nil {
		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf("Error occurred validating a WebAuthn registration challenge for user '%s': %s", userSession.Username, errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if w, err = ctx.GetWebAuthnProvider(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn registration challenge for user '%s': error occurred provisioning the configuration", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if user, err = handleGetWebAuthnUserByRPID(ctx, userSession.Username, userSession.DisplayName, w.Config.RPID); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn registration challenge for user '%s': error occurred retrieving the WebAuthn user configuration from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if c, err = w.CreateCredential(user, *userSession.WebAuthn.SessionData, response); err != nil {
		ctx.Logger.WithError(iwebauthn.FormatError(err)).Errorf("Error occurred validating a WebAuthn registration challenge for user '%s': error comparing the response to the WebAuthn session data", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	credential := model.NewWebAuthnCredential(ctx, w.Config.RPID, userSession.Username, userSession.WebAuthn.Description, c)

	credential.Discoverable = iwebauthn.IsCredentialCreationDiscoverable(ctx.Logger, response)

	if err = iwebauthn.ValidateCredentialAllowed(&ctx.Configuration.WebAuthn, &credential); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn registration challenge for user '%s': error occurred processing the credential filtering", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveWebAuthnCredential(ctx, credential); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating a WebAuthn registration challenge for user '%s': error occurred saving the credential to the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageUnableToRegisterSecurityKey)

		return
	}

	ctx.ReplyOK()
	ctx.SetStatusCode(fasthttp.StatusCreated)

	body := emailEventBody{
		Prefix: eventEmailAction2FAPrefix,
		Body:   eventEmailAction2FABody,
		Suffix: eventEmailAction2FAAddedSuffix,
	}

	ctxLogEvent(ctx, userSession.Username, eventLogAction2FAAdded, body, map[string]any{eventLogKeyAction: eventLogAction2FAAdded, eventLogKeyCategory: eventLogCategoryWebAuthnCredential, eventLogKeyDescription: credential.Description})
}

// WebAuthnRegistrationDELETE deletes any active WebAuthn registration session..
func WebAuthnRegistrationDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting a WebAuthn registration challenge: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred deleting a WebAuthn registration challenge")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.WebAuthn != nil {
		userSession.WebAuthn = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred deleting a WebAuthn registration challenge for user '%s': %s", userSession.Username, errStrUserSessionDataSave)

			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetJSONError(messageOperationFailed)

			return
		}
	}

	ctx.ReplyOK()
}
