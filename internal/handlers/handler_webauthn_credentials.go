package handlers

import (
	"encoding/json"
	"errors"
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
	credentialIDStr, ok := ctx.UserValue("credentialID").(string)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return 0, errors.New("Invalid credential ID type")
	}

	credentialID, err := strconv.Atoi(credentialIDStr)
	if err != nil {
		ctx.Error(err, messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return 0, err
	}

	return credentialID, nil
}

// WebAuthnCredentialsGET returns all credentials registered for the current user.
func WebAuthnCredentialsGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		origin      *url.URL
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred loading WebAuthn credentials: error occurred loading session data")

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(fmt.Errorf("user is anonymous")).Errorf("Error occurred loading WebAuthn credentials")

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
		ctx.Logger.WithError(err).Errorf("Error occurred loading WebAuthn credentials for user '%s'", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(credentials); err != nil {
		ctx.Logger.WithError(err).Errorf("Error ccurred loading WebAuthn credentials for user '%s': error occurred attempting to write the response body", userSession.Username)
	}
}

// WebAuthnCredentialPUT updates the description for a specific credential for the current user.
func WebAuthnCredentialPUT(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON bodyEditWebAuthnCredentialRequest

		id          int
		credential  *model.WebAuthnCredential
		userSession session.UserSession

		err error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred modifying WebAuthn credential: error occurred loading session data")

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(fmt.Errorf("user is anonymous")).Errorf("Error occurred modifying WebAuthn credential")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred modifying WebAuthn credential: error occurred parsing the form data")

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	if id, err = getWebAuthnCredentialIDFromContext(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential for user '%s': error occurred trying to determine the credential ID", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if len(bodyJSON.Description) == 0 {
		ctx.Logger.WithError(fmt.Errorf("description is empty")).Errorf("Error occurred modifying WebAuthn credential for user '%s", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if credential, err = ctx.Providers.StorageProvider.LoadWebAuthnCredentialByID(ctx, id); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential for user '%s': error occurred trying to load the credential", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if credential.Username != userSession.Username {
		ctx.Logger.WithError(fmt.Errorf("user '%s' owns the credential with id '%d'", credential.Username, credential.ID)).Errorf("Error occurred modifying WebAuthn credential for user '%s'", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.UpdateWebAuthnCredentialDescription(ctx, userSession.Username, id, bodyJSON.Description); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying WebAuthn credential for user '%s': error occurred while attempting to save the modified credential in storage", userSession.Username)

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
		ctx.Logger.WithError(err).Error("Error occurred deleting WebAuthn credential: error occurred loading session data")

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(fmt.Errorf("user is anonymous")).Errorf("Error occurred modifying WebAuthn credential")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if id, err = getWebAuthnCredentialIDFromContext(ctx); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting WebAuthn credential: error occurred trying to determine the credential ID")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if credential, err = ctx.Providers.StorageProvider.LoadWebAuthnCredentialByID(ctx, id); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting WebAuthn credential for user '%s': error occurred trying to load the credential", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if credential.Username != userSession.Username {
		ctx.Logger.WithError(fmt.Errorf("user '%s' owns the credential with id '%d'", credential.Username, credential.ID)).Errorf("Error occurred deleting WebAuthn credential for user '%s'", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.DeleteWebAuthnCredential(ctx, credential.KID.String()); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred delete WebAuthn credential for user '%s': error occurred while attempting to delete the credential from storage", userSession.Username)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctxLogEvent(ctx, userSession.Username, eventLogAction2FARemoved, map[string]any{eventLogKeyAction: eventLogAction2FARemoved, eventLogKeyCategory: eventLogCategoryWebAuthnCredential, eventLogKeyDescription: credential.Description})

	ctx.ReplyOK()
}
