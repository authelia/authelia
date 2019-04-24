package handlers

import (
	"github.com/clems4ever/authelia/authentication"
	"github.com/clems4ever/authelia/middlewares"
)

// SecondFactorAvailableMethodsGet retrieve available 2FA methods.
// The supported methods are: "totp", "u2f", "duo"
func SecondFactorAvailableMethodsGet(ctx *middlewares.AutheliaCtx) {
	availableMethods := MethodList{authentication.TOTP, authentication.U2F}

	if ctx.Configuration.DuoAPI != nil {
		availableMethods = append(availableMethods, authentication.DuoPush)
	}

	ctx.Logger.Debugf("Available methods are %s", availableMethods)
	ctx.SetJSONBody(availableMethods)
}
