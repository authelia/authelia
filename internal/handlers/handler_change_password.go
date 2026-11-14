// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

// ChangePasswordPOST handles the change password request.
func ChangePasswordPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		provider    session.Strategy
		err         error
	)
	if provider, err = ctx.GetSessionProvider(); err != nil {
		ctx.GetLogger().WithError(err).
			Error("Unable to change password for user: error occurred retrieving session provider")
		ctx.SetJSONError(messageUnableToChangePassword)
		ctx.SetStatusCode(http.StatusInternalServerError)

		return
	}

	var current *session.UserSession

	if current, err = provider.Get(ctx); err != nil {
		ctx.GetLogger().WithError(err).
			Error("Unable to change password for user: error occurred retrieving session for user")
		ctx.SetJSONError(messageUnableToChangePassword)
		ctx.SetStatusCode(http.StatusInternalServerError)

		return
	}

	if err = ctx.RegenerateSession(); err != nil {
		ctx.GetLogger().WithError(err).
			Error("Unable to change password for user: failed to regenerate session")
		ctx.SetJSONError(messageUnableToChangePassword)
		ctx.SetStatusCode(http.StatusInternalServerError)

		return
	}

	userSession = *current

	username := userSession.Username

	var requestBody changePasswordRequestBody

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.GetLogger().WithError(err).
			WithFields(map[string]any{"username": username}).
			Error("Unable to change password for user: unable to parse request body")
		ctx.SetJSONError(messageUnableToChangePassword)
		ctx.SetStatusCode(http.StatusBadRequest)

		return
	}

	if err = ctx.Providers.PasswordPolicy.Check(requestBody.NewPassword); err != nil {
		ctx.GetLogger().WithError(err).
			WithFields(map[string]any{"username": username}).
			Debug("Unable to change password for user as their new password was weak or empty")
		ctx.SetJSONError(messagePasswordWeak)
		ctx.SetStatusCode(http.StatusBadRequest)

		return
	}

	if err = ctx.Providers.UserProvider.ChangePassword(username, requestBody.OldPassword, requestBody.NewPassword); err != nil {
		respondChangePasswordError(ctx, username, err)

		return
	}

	ctx.GetLogger().
		WithFields(map[string]any{"username": username}).
		Debug("User has changed their password")

	// The password has changed whether or not the session is saved, so the user is still notified of it. The failure
	// is reported once the notification has been attempted.
	failed := false

	if err = provider.Save(ctx, &userSession); err != nil {
		ctx.GetLogger().WithError(err).
			WithFields(map[string]any{"username": username}).
			Error("Unable to update password change state")

		failed = true
	}

	userInfo, err := ctx.Providers.UserProvider.GetDetails(username)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred retrieving user details")

		ctx.EmitEvent(events.NewEvent(&events.DataUserPassword{
			Type:         events.TypeUserPasswordChanged,
			Username:     username,
			RemoteIP:     ctx.RemoteIP().String(),
			Notification: ctx.NewNotification(err, "Password changed successfully", nil, nil),
		}))

		replyPasswordModified(ctx, failed)

		return
	}

	data := templates.EmailEventValues{
		Title:       "Password changed successfully",
		DisplayName: userInfo.DisplayName,
		RemoteIP:    ctx.RemoteIP().String(),
		Details: map[string]any{
			"Action": "Password Change",
		},
		BodyPrefix: eventEmailActionPasswordModifyPrefix,
		BodyEvent:  eventEmailActionPasswordChange,
		BodySuffix: eventEmailActionPasswordModifySuffix,
	}

	var addresses []mail.Address

	notified := len(userInfo.Emails) != 0

	if !notified {
		err = fmt.Errorf("user has no email address configured")

		ctx.GetLogger().WithFields(map[string]any{"username": username}).
			Debug("user has no email address configured")
	} else {
		addresses = userInfo.Addresses()

		ctx.GetLogger().WithFields(map[string]any{
			"username": username,
			"email":    addresses[0].String(),
		}).
			Debug("Sending an email to inform user that their password has changed.")

		err = ctx.Providers.Notifier.Send(ctx, addresses[0], "Password changed successfully", ctx.Providers.Templates.GetEventEmailTemplate(), data)
	}

	ctx.EmitEvent(events.NewEvent(&events.DataUserPassword{
		Type:        events.TypeUserPasswordChanged,
		Username:    username,
		DisplayName: userInfo.DisplayName,
		Emails:      userInfo.Emails,
		RemoteIP:    ctx.RemoteIP().String(),
		Notification: ctx.NewNotification(err, data.Title, recipientsFromDetails(username, userInfo), &events.NotificationValues{
			BodyPrefix: data.BodyPrefix,
			BodyEvent:  data.BodyEvent,
			BodySuffix: data.BodySuffix,
			Details:    data.Details,
		}),
	}))

	if !notified {
		replyPasswordModified(ctx, failed)

		return
	}

	if err != nil {
		ctx.GetLogger().WithError(err).
			WithFields(map[string]any{
				"username": username,
				"email":    addresses[0].String(),
			}).
			Debug("Unable to notify user of password change")
		replyPasswordModified(ctx, failed)

		return
	}

	if failed {
		replyPasswordModified(ctx, failed)
	}
}

func respondChangePasswordError(ctx *middlewares.AutheliaCtx, username string, err error) {
	switch {
	case errors.Is(err, authentication.ErrIncorrectPassword):
		ctx.GetLogger().WithError(err).
			WithFields(map[string]any{"username": username}).
			Debug("Unable to change password for user as their old password was incorrect")
		ctx.SetJSONError(messageIncorrectPassword)
		ctx.SetStatusCode(http.StatusUnauthorized)
	case errors.Is(err, authentication.ErrPasswordWeak):
		ctx.GetLogger().WithError(err).
			WithFields(map[string]any{"username": username}).
			Debug("Unable to change password for user as their new password was weak or empty")
		ctx.SetJSONError(messagePasswordWeak)
		ctx.SetStatusCode(http.StatusBadRequest)
	case errors.Is(err, authentication.ErrAuthenticationFailed):
		ctx.GetLogger().WithError(err).
			WithFields(map[string]any{"username": username}).
			Error("Unable to change password for user as authentication failed for the user")
		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(http.StatusUnauthorized)
	default:
		ctx.GetLogger().WithError(err).
			WithFields(map[string]any{"username": username}).
			Error("Unable to change password for user for an unknown reason")
		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(http.StatusInternalServerError)
	}
}
