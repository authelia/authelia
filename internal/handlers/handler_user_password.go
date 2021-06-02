package handlers

import (
	"fmt"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

type updatePasswordRequestBody struct {
	OldPassword string `json:"old_password"`
	Password    string `json:"password"`
}

// UserPasswordPost handler for resetting passwords.
func UserPasswordPost(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	var requestBody updatePasswordRequestBody
	err := ctx.ParseBody(&requestBody)

	if err != nil {
		ctx.Error(err, unableToUpdatePasswordMessage)
		return
	}

	ok, err := ctx.Providers.UserProvider.CheckUserPassword(userSession.Username, requestBody.OldPassword)
	if !ok || err != nil {
		ctx.Error(err, unableToUpdatePasswordMessage)
		return
	}

	err = ctx.Providers.UserProvider.UpdatePassword(userSession.Username, requestBody.Password)

	if err != nil {
		switch {
		case utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityCodes):
			ctx.Error(fmt.Errorf("%s", err), ldapPasswordComplexityCode)
		case utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityErrors):
			ctx.Error(fmt.Errorf("%s", err), ldapPasswordComplexityCode)
		default:
			ctx.Error(fmt.Errorf("%s", err), unableToUpdatePasswordMessage)
		}

		return
	}

	ctx.Logger.Debugf("Password of user %s has been modified", userSession.Username)

	ctx.ReplyOK()
}
