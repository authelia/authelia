package handlers

import (
	"bytes"
	"fmt"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ResetPasswordPOST handler for resetting passwords.
func ResetPasswordPOST(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	// Those checks unsure that the identity verification process has been initiated and completed successfully
	// otherwise PasswordReset would not be set to true. We can improve the security of this check by making the
	// request expire at some point because here it only expires when the cookie expires.
	if userSession.PasswordResetUsername == nil {
		ctx.Error(fmt.Errorf("no identity verification process has been initiated"), messageUnableToResetPassword)
		return
	}

	username := *userSession.PasswordResetUsername

	var requestBody resetPasswordStep2RequestBody
	err := ctx.ParseBody(&requestBody)

	if err != nil {
		ctx.Error(err, messageUnableToResetPassword)
		return
	}

	if err = ctx.Providers.PasswordPolicy.Check(requestBody.Password); err != nil {
		ctx.Error(err, messagePasswordWeak)
		return
	}

	if err = ctx.Providers.UserProvider.UpdatePassword(username, requestBody.Password); err != nil {
		switch {
		case utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityCodes),
			utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityErrors):
			ctx.Error(err, ldapPasswordComplexityCode)
		default:
			ctx.Error(err, messageUnableToResetPassword)
		}

		return
	}

	ctx.Logger.Debugf("Password of user %s has been reset", username)

	// Reset the request.
	userSession.PasswordResetUsername = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Error(fmt.Errorf("unable to update password reset state: %s", err), messageOperationFailed)
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

	bufHTML := new(bytes.Buffer)

	disableHTML := false
	if ctx.Configuration.Notifier.SMTP != nil {
		disableHTML = ctx.Configuration.Notifier.SMTP.DisableHTMLEmails
	}

	values := templates.EmailPasswordResetValues{
		Title:       "Password changed successfully",
		DisplayName: userInfo.DisplayName,
		RemoteIP:    ctx.RemoteIP().String(),
	}

	if !disableHTML {
		if err = ctx.Providers.Templates.ExecuteEmailPasswordResetTemplate(bufHTML, values, templates.HTMLFormat); err != nil {
			ctx.Logger.Error(err)
			ctx.ReplyOK()

			return
		}
	}

	bufText := new(bytes.Buffer)

	if err = ctx.Providers.Templates.ExecuteEmailPasswordResetTemplate(bufText, values, templates.PlainTextFormat); err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}

	addresses := userInfo.Addresses()

	ctx.Logger.Debugf("Sending an email to user %s (%s) to inform that the password has changed.",
		username, addresses[0])

	if err = ctx.Providers.Notifier.Send(addresses[0], "Password changed successfully", bufText.String(), bufHTML.String()); err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}
}
