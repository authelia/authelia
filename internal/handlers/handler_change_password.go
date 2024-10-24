package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// these are redefined in /internal/authentication/const.go.
var ErrIncorrectPassword = errors.New("incorrect password")
var ErrReusePassword = errors.New("you cannot reuse your old password")

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func ChangePasswordPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Error(fmt.Errorf("error occurred retrieving session for user: %w", err), messageUnableToChangePassword)
		return
	}

	tempUsername := userSession.Username

	userSession.PasswordResetUsername = &tempUsername

	if userSession.PasswordResetUsername == nil {
		ctx.Error(fmt.Errorf("elevated session required for password change"), messageUnableToChangePassword)
	}

	username := *userSession.PasswordResetUsername

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
		case strings.Contains(err.Error(), ErrIncorrectPassword.Error()):
			ctx.Error(err, messageIncorrectPassword)
		case strings.Contains(err.Error(), ErrReusePassword.Error()):
			ctx.Error(err, messageCannotReusePassword)
		case utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityCodes),
			utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityErrors):
			ctx.Error(err, ldapPasswordComplexityCode)
		default:
			ctx.Error(err, messageUnableToChangePassword)
		}

		return
	}

	ctx.Logger.Debugf("User %s has changed their password", username)

	if err = ctx.Providers.StorageProvider.UpdatePasswordChangedDateByUsername(ctx, userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Could not update %s, for user '%s'", "last password change time", userSession.Username)
	}

	userSession.PasswordResetUsername = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Error(fmt.Errorf("unable to update password reset state: %w", err), messageOperationFailed)
		return
	}

	// Send Notification.
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
