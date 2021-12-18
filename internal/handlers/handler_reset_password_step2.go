package handlers

import (
	"fmt"
	"regexp"

	"github.com/authelia/authelia/v4/internal/middlewares"
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

	err = ctx.Providers.UserProvider.UpdatePassword(*userSession.PasswordResetUsername, requestBody.Password)

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

	ctx.Logger.Debugf("Password of user %s has been reset", *userSession.PasswordResetUsername)

	// Reset the request.
	userSession.PasswordResetUsername = nil
	err = ctx.SaveSession(userSession)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to update password reset state: %s", err), messageOperationFailed)
		return
	}

	ctx.ReplyOK()
}

// validatePassword validates if the password met the password policy rules.
func validatePassword(ctx *middlewares.AutheliaCtx, password string) error {
	requireLowercase := ctx.Configuration.PasswordPolicy.RequireLowercase
	requireUppercase := ctx.Configuration.PasswordPolicy.RequireUppercase
	requireNumber := ctx.Configuration.PasswordPolicy.RequireNumber
	requireSpecial := ctx.Configuration.PasswordPolicy.RequireSpecial

	if requireLowercase {
		regexStr := "[a-z]+"
		re, err := regexp.Compile(regexStr)

		if err != nil {
			return err
		}

		if found := re.MatchString(password); !found {
			return errPasswordPolicyNoMet
		}
	}

	if requireUppercase {
		regexStr := "[A-Z]+"
		re, err := regexp.Compile(regexStr)

		if err != nil {
			return err
		}

		if found := re.MatchString(password); !found {
			return errPasswordPolicyNoMet
		}
	}

	if requireNumber {
		regexStr := "[0-9]+"
		re, err := regexp.Compile(regexStr)

		if err != nil {
			return err
		}

		if found := re.MatchString(password); !found {
			return errPasswordPolicyNoMet
		}
	}

	if requireSpecial {
		regexStr := "[^a-zA-Z0-9]+"
		re, err := regexp.Compile(regexStr)

		if err != nil {
			return err
		}

		if found := re.MatchString(password); !found {
			return errPasswordPolicyNoMet
		}
	}

	return nil
}