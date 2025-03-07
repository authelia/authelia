package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

func ChangePasswordPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Error(fmt.Errorf("error occurred retrieving session for user: %w", err), messageUnableToChangePassword)
		return
	}

	username := userSession.Username

	var requestBody changePasswordRequestBody

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Error(err, messageUnableToChangePassword)
		return
	}

	if err = ctx.Providers.PasswordPolicy.Check(requestBody.NewPassword); err != nil {
		ctx.Error(err, messagePasswordWeak)
		return
	}

	if err = ctx.Providers.UserProvider.ChangePassword(username, requestBody.OldPassword, requestBody.NewPassword); err != nil {
		switch {
		case errors.Is(err, authentication.ErrIncorrectPassword):
			ctx.Logger.WithError(err).
				WithFields(map[string]any{"username": username}).
				Debug("Unable to change password for user as their old password was incorrect")
			ctx.SetJSONError(messageIncorrectPassword)
			ctx.SetStatusCode(http.StatusUnauthorized)
		case errors.Is(err, authentication.ErrPasswordWeak):
			ctx.Logger.WithError(err).
				WithFields(map[string]any{"username": username}).
				Debug("Unable to change password for user as their new password was weak or empty")
			ctx.SetJSONError(messagePasswordWeak)
			ctx.SetStatusCode(http.StatusBadRequest)
		case errors.Is(err, authentication.ErrAuthenticationFailed):
			ctx.Logger.WithError(err).
				WithFields(map[string]any{"username": username}).
				Error("Unable to change password for user as authentication failed for the user")
			ctx.SetJSONError(messageOperationFailed)
			ctx.SetStatusCode(http.StatusUnauthorized)
		default:
			ctx.Logger.WithError(err).
				WithFields(map[string]any{"username": username}).
				Error("Unable to change password for user for an unknown reason")
			ctx.SetJSONError(messageOperationFailed)
			ctx.SetStatusCode(http.StatusInternalServerError)
		}

		return
	}

	ctx.Logger.Debugf("User %s has changed their password", username)

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Error(fmt.Errorf("unable to update password reset state: %w", err), messageOperationFailed)
		return
	}

	userInfo, err := ctx.Providers.UserProvider.GetDetails(username)
	if err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}

	if len(userInfo.Emails) == 0 {
		ctx.Logger.Error(fmt.Errorf("user %s has no email address configured", username))
		ctx.ReplyOK()

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

	addresses := userInfo.Addresses()

	ctx.Logger.Debugf("Sending an email to user %s (%s) to inform that the password has changed.",
		username, addresses[0].String())

	if err = ctx.Providers.Notifier.Send(ctx, addresses[0], "Password changed successfully", ctx.Providers.Templates.GetEventEmailTemplate(), data); err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}
}
