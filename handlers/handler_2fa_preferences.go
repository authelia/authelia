package handlers

import (
	"fmt"

	"github.com/clems4ever/authelia/authentication"
	"github.com/clems4ever/authelia/middlewares"
)

// SecondFactorPreferencesGet get the user preferences regarding 2FA.
func SecondFactorPreferencesGet(ctx *middlewares.AutheliaCtx) {
	preferences := preferences{
		Method: "totp",
	}

	userSession := ctx.GetSession()
	method, err := ctx.Providers.StorageProvider.LoadPrefered2FAMethod(userSession.Username)
	ctx.Logger.Debugf("Loaded prefered 2FA method of user %s is %s", userSession.Username, method)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to load prefered 2FA method: %s", err), operationFailedMessage)
		return
	}

	if method != "" {
		// Set the retrieved method.
		preferences.Method = method
	}

	ctx.SetJSONBody(preferences)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// SecondFactorPreferencesPost update the user preferences regarding 2FA.
func SecondFactorPreferencesPost(ctx *middlewares.AutheliaCtx) {
	bodyJSON := preferences{}

	err := ctx.ParseBody(&bodyJSON)
	if err != nil {
		ctx.Error(err, operationFailedMessage)
		return
	}

	if !stringInSlice(bodyJSON.Method, authentication.PossibleMethods) {
		ctx.Error(fmt.Errorf("Unknown method %s, it should be either u2f, totp or duo_push", bodyJSON.Method), operationFailedMessage)
		return
	}

	userSession := ctx.GetSession()

	ctx.Logger.Debugf("Save new prefered 2FA method of user %s to %s", userSession.Username, bodyJSON.Method)
	err = ctx.Providers.StorageProvider.SavePrefered2FAMethod(userSession.Username, bodyJSON.Method)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to save new prefered 2FA method: %s", err), operationFailedMessage)
		return
	}

	ctx.ReplyOK()
}
