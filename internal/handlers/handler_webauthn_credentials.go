package handlers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func getWebAuthnCredentialIDFromContext(ctx *middlewares.AutheliaCtx) (int, error) {
	value := ctx.UserValue("credentialID")

	switch v := value.(type) {
	case nil:
		return 0, fmt.Errorf("error occurred retrieving WebAuthn Credential ID from context: the user value wasn't set")
	case string:
		credentialID, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("error occurred retrieving WebAuthn Credential ID from context: failed to parse '%s' as an integer: %w", v, err)
		}

		return credentialID, nil
	default:
		return 0, fmt.Errorf("error occurred retrieving WebAuthn Credential ID from context: the type '%T' is not a string", value)
	}
}

// WebAuthnCredentialsGET returns all credentials registered for the current user.
func WebAuthnCredentialsGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		origin      *url.URL
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred loading WebAuthn credentials: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Errorf("Error occurred loading WebAuthn credentials")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if origin, err = ctx.GetOrigin(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred loading WebAuthn credentials for user '%s': error occurred attempting to retrieve origin", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var credentials []model.WebAuthnCredential

	if credentials, err = ctx.Providers.StorageProvider.LoadWebAuthnCredentialsByUsername(ctx, origin.Hostname(), userSession.Username); err != nil && err != storage.ErrNoWebAuthnCredential {
		ctx.Logger.WithError(err).Errorf("Error occurred loading WebAuthn credentials for user '%s': error occurred loading credentials from the storage backend", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(credentials); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred loading WebAuthn credentials for user '%s': %s", userSession.Username, errStrRespBody)
	}
}

// WebAuthnCredentialPUT updates the description for a specific credential for the current user.
func WebAuthnCredentialPUT(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON bodyEditWebAuthnCredentialRequest

		id          int
		credential  *model.WebAuthnCredential
		userSession session.UserSession

		origin      *url.URL
		credentials []model.WebAuthnCredential

		err error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Errorf("Error occurred modifying WebAuthn credential")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential for user '%s': %s", userSession.Username, errStrReqBodyParse)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if len(bodyJSON.Description) == 0 {
		ctx.Logger.WithError(fmt.Errorf("description is empty")).Errorf("Error occurred modifying WebAuthn credential for user '%s", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if id, err = getWebAuthnCredentialIDFromContext(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential for user '%s': error occurred trying to determine the credential ID", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if credential, err = ctx.Providers.StorageProvider.LoadWebAuthnCredentialByID(ctx, id); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential for user '%s': error occurred loading the credential from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if credential.Username != userSession.Username {
		ctx.Logger.WithError(fmt.Errorf("user '%s' owns the credential with id '%d'", credential.Username, credential.ID)).Errorf("Error occurred modifying WebAuthn credential for user '%s'", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if origin, err = ctx.GetOrigin(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential for user '%s': error occurred determining the origin for the request", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if credentials, err = ctx.Providers.StorageProvider.LoadWebAuthnCredentialsByUsername(ctx, origin.Hostname(), userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential for user '%s': error occurred looking up existing credentials", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	for _, c := range credentials {
		if c.ID == id {
			continue
		}

		if c.Description == bodyJSON.Description {
			ctx.Logger.WithError(fmt.Errorf("credential with id '%d' also has the description '%s'", c.ID, bodyJSON.Description)).Errorf("Error occurred modifying WebAuthn credential for user '%s': error occurred ensuring the credentials had unique descriptions", userSession.Username)

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetJSONError(messageOperationFailed)

			return
		}
	}

	if err = ctx.Providers.StorageProvider.UpdateWebAuthnCredentialDescription(ctx, userSession.Username, id, bodyJSON.Description); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential for user '%s': error occurred while attempting to update the modified credential in the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}

// WebAuthnCredentialDELETE deletes a specific credential for the current user.
func WebAuthnCredentialDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		id          int
		credential  *model.WebAuthnCredential
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting WebAuthn credential: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Errorf("Error occurred modifying WebAuthn credential")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if id, err = getWebAuthnCredentialIDFromContext(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting WebAuthn credential for user '%s': error occurred trying to determine the credential ID", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if credential, err = ctx.Providers.StorageProvider.LoadWebAuthnCredentialByID(ctx, id); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting WebAuthn credential for user '%s': error occurred trying to load the credential from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if credential.Username != userSession.Username {
		ctx.Logger.WithError(fmt.Errorf("user '%s' owns the credential with id '%d'", credential.Username, credential.ID)).Errorf("Error occurred deleting WebAuthn credential for user '%s'", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.DeleteWebAuthnCredential(ctx, credential.KID.String()); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred delete WebAuthn credential for user '%s': error occurred while attempting to delete the credential from the storage backend", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctxLogEvent(ctx, userSession.Username, eventLogAction2FARemoved, map[string]any{eventLogKeyAction: eventLogAction2FARemoved, eventLogKeyCategory: eventLogCategoryWebAuthnCredential, eventLogKeyDescription: credential.Description})

	ctx.ReplyOK()
}
