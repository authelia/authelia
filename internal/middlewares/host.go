package middlewares

import (
	"fmt"
)

func GetForwardedURI(ctx *AutheliaCtx) (string, error) {
	if ctx.XForwardedProto() == nil {
		return "", errMissingXForwardedProto
	}

	if ctx.XForwardedHost() == nil {
		return "", errMissingXForwardedHost
	}

	return fmt.Sprintf("%s://%s%s", ctx.XForwardedProto(),
		ctx.XForwardedHost(), ctx.Configuration.Server.Path), nil
}
