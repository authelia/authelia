package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func authzGetObjectImplAuthRequest(ctx *middlewares.AutheliaCtx) (object authorization.Object, err error) {
	var targetURL *url.URL

	if targetURL, err = url.ParseRequestURI(string(ctx.XOriginalURL())); err != nil {
		return object, fmt.Errorf("failed to parse header X-Original-URL: %w", err)
	}

	return authorization.NewObjectRaw(targetURL, ctx.XOriginalMethod()), nil
}

func authzHandleUnauthorizedImplAuthRequest(ctx *middlewares.AutheliaCtx, authn *Authn, _ *url.URL) {
	ctx.Logger.Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d", authn.Object.URL.String(), authn.Method, authn.Username, fasthttp.StatusUnauthorized)
	ctx.ReplyUnauthorized()
}
