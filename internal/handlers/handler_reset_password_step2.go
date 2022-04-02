package handlers

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ResetPasswordPost handler for resetting passwords.
func ResetPasswordPost(ctx *middlewares.AutheliaCtx) {
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

	if err := validatePassword(ctx, requestBody.Password); err != nil {
		ctx.Error(err, messagePasswordWeak)
		return
	}

	err = ctx.Providers.UserProvider.UpdatePassword(username, requestBody.Password)

	if err != nil {
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
	err = ctx.SaveSession(userSession)

	if err != nil {
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
	if ctx.Configuration.Notifier != nil && ctx.Configuration.Notifier.SMTP != nil {
		disableHTML = ctx.Configuration.Notifier.SMTP.DisableHTMLEmails
	}

	if !disableHTML {
		htmlParams := map[string]interface{}{
			"title":       "Password changed successfully",
			"displayName": userInfo.DisplayName,
			"remoteIP":    ctx.RemoteIP().String(),
		}

		err = templates.HTMLEmailTemplateStep2.Execute(bufHTML, htmlParams)

		if err != nil {
			ctx.Logger.Error(err)
			ctx.ReplyOK()

			return
		}
	}

	bufText := new(bytes.Buffer)
	textParams := map[string]interface{}{
		"displayName": userInfo.DisplayName,
	}

	err = templates.PlainTextEmailTemplateStep2.Execute(bufText, textParams)

	if err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}

	ctx.Logger.Debugf("Sending an email to user %s (%s) to inform that the password has changed.",
		username, userInfo.Emails[0])

	err = ctx.Providers.Notifier.Send(userInfo.Emails[0], "Password changed successfully", bufText.String(), bufHTML.String())

	if err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}
}

// validatePassword validates if the password met the password policy rules.
func validatePassword(ctx *middlewares.AutheliaCtx, password string) error {
	// password validation applies only to standard passwor policy.
	if !ctx.Configuration.PasswordPolicy.Standard.Enabled {
		return nil
	}

	requireLowercase := ctx.Configuration.PasswordPolicy.Standard.RequireLowercase
	requireUppercase := ctx.Configuration.PasswordPolicy.Standard.RequireUppercase
	requireNumber := ctx.Configuration.PasswordPolicy.Standard.RequireNumber
	requireSpecial := ctx.Configuration.PasswordPolicy.Standard.RequireSpecial
	minLength := ctx.Configuration.PasswordPolicy.Standard.MinLength
	maxlength := ctx.Configuration.PasswordPolicy.Standard.MaxLength

	var patterns []string

	if (minLength > 0 && len(password) < minLength) || (maxlength > 0 && len(password) > maxlength) {
		return errPasswordPolicyNoMet
	}

	if requireLowercase {
		patterns = append(patterns, "[a-z]+")
	}

	if requireUppercase {
		patterns = append(patterns, "[A-Z]+")
	}

	if requireNumber {
		patterns = append(patterns, "[0-9]+")
	}

	if requireSpecial {
		patterns = append(patterns, "[^a-zA-Z0-9]+")
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)

		if err != nil {
			return err
		}

		if found := re.MatchString(password); !found {
			return errPasswordPolicyNoMet
		}
	}

	return nil
}
