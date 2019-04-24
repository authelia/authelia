package handlers

import (
	"fmt"

	"github.com/clems4ever/authelia/middlewares"
)

// LogoutPost is the handler logging out the user attached to the given cookie.
func LogoutPost(ctx *middlewares.AutheliaCtx) {
	ctx.Logger.Debug("Destroy session")
	err := ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to destroy session during logout: %s", err), operationFailedMessage)
	}

	ctx.ReplyOK()
}
