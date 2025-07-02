package handlers

import (
	"errors"
	"net/http"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

func ChangePasswordPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		provider    *session.Session
		err         error
	)
	if provider, err = ctx.GetSessionProvider(); err != nil {
		ctx.Logger.WithError(err).
			Error("Unable to change password for user: error occurred retrieving session provider")
		ctx.SetJSONError(messageUnableToChangePassword)
		ctx.SetStatusCode(http.StatusInternalServerError)

		return
	}

	if userSession, err = provider.GetSession(ctx.RequestCtx); err != nil {
		ctx.Logger.WithError(err).
			Error("Unable to change password for user: error occurred retrieving session for user")
		ctx.SetJSONError(messageUnableToChangePassword)
		ctx.SetStatusCode(http.StatusInternalServerError)

		return
	}

	username := userSession.Username

	var requestBody changePasswordRequestBody

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Logger.WithError(err).
			WithFields(map[string]any{"username": username}).
			Error("Unable to change password for user: unable to parse request body")
		ctx.SetJSONError(messageUnableToChangePassword)
		ctx.SetStatusCode(http.StatusBadRequest)

		return
	}

	if err = ctx.Providers.PasswordPolicy.Check(requestBody.NewPassword); err != nil {
		ctx.Logger.WithError(err).
			WithFields(map[string]any{"username": username}).
			Debug("Unable to change password for user as their new password was weak or empty")
		ctx.SetJSONError(messagePasswordWeak)
		ctx.SetStatusCode(http.StatusBadRequest)

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

	ctx.Logger.
		WithFields(map[string]any{"username": username}).
		Debug("User has changed their password")

	if err = provider.SaveSession(ctx.RequestCtx, userSession); err != nil {
		ctx.Logger.WithError(err).
			WithFields(map[string]any{"username": username}).
			Error("Unable to update password change state")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	userInfo, err := ctx.Providers.UserProvider.GetDetails(username)
	if err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}

	if len(userInfo.Emails) == 0 {
		ctx.Logger.WithFields(map[string]any{"username": username}).
			Debug("user has no email address configured")
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

	ctx.Logger.WithFields(map[string]any{
		"username": username,
		"email":    addresses[0].String(),
	}).
		Debug("Sending an email to inform user that their password has changed.")

	if err = ctx.Providers.Notifier.Send(ctx, addresses[0], "Password changed successfully", ctx.Providers.Templates.GetEventEmailTemplate(), data); err != nil {
		ctx.Logger.WithError(err).
			WithFields(map[string]any{
				"username": username,
				"email":    addresses[0].String(),
			}).
			Debug("Unable to notify user of password change")
		ctx.ReplyOK()

		return
	}
}
