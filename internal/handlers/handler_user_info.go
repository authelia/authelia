package handlers

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

// UserInfoGet get the info related to the user identified by the session.
func UserInfoGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	userInfo, err := ctx.Providers.StorageProvider.LoadUserInfo(ctx, userSession.Username)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to load user information: %v", err), messageOperationFailed)
		return
	}

	userInfo.DisplayName = userSession.DisplayName

	err = ctx.SetJSONBody(userInfo)
	if err != nil {
		ctx.Logger.Errorf("Unable to set user info response in body: %s", err)
	}
}

// MethodPreferencePost update the user preferences regarding 2FA method.
func MethodPreferencePost(ctx *middlewares.AutheliaCtx) {
	bodyJSON := preferred2FAMethodBody{}

	err := ctx.ParseBody(&bodyJSON)
	if err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}

	if !utils.IsStringInSlice(bodyJSON.Method, authentication.PossibleMethods) {
		ctx.Error(fmt.Errorf("unknown method '%s', it should be one of %s", bodyJSON.Method, strings.Join(authentication.PossibleMethods, ", ")), messageOperationFailed)
		return
	}

	userSession := ctx.GetSession()
	ctx.Logger.Debugf("Save new preferred 2FA method of user %s to %s", userSession.Username, bodyJSON.Method)
	err = ctx.Providers.StorageProvider.SavePreferred2FAMethod(ctx, userSession.Username, bodyJSON.Method)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to save new preferred 2FA method: %s", err), messageOperationFailed)
		return
	}

	ctx.ReplyOK()
}
