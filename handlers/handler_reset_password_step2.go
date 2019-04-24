package handlers

import (
	"fmt"

	"github.com/clems4ever/authelia/middlewares"
)

// ResetPasswordPost handler for resetting passwords
func ResetPasswordPost(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	// Those checks unsure that the identity verification process has been initiated and completed successfully
	// otherwise PasswordReset would not be set to true. We can improve the security of this check by making the
	// request expire at some point because here it only expires when the cookie expires...
	if userSession.PasswordResetUsername == nil {
		ctx.Error(fmt.Errorf("No identity verification process has been initiated"), unableToResetPasswordMessage)
		return
	}

	var requestBody resetPasswordStep2RequestBody
	err := ctx.ParseBody(&requestBody)

	if err != nil {
		ctx.Error(err, unableToResetPasswordMessage)
		return
	}

	err = ctx.Providers.UserProvider.UpdatePassword(*userSession.PasswordResetUsername, requestBody.Password)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to update password: %s", err), unableToResetPasswordMessage)
		return
	}

	ctx.Logger.Debugf("Password of user %s has been reset", *userSession.PasswordResetUsername)

	// Reset the request.
	userSession.PasswordResetUsername = nil
	err = ctx.SaveSession(userSession)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to update password reset state: %s", err), operationFailedMessage)
		return
	}

	ctx.ReplyOK()
}
