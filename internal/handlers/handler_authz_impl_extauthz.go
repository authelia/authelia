package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func authzGetObjectImplExtAuthz(ctx *middlewares.AutheliaCtx) (object authorization.Object, err error) {
	var targetURL *url.URL

	if targetURL, err = ctx.GetForwardedURL(); err != nil {
		return object, fmt.Errorf("failed to get target URL: %w", err)
	}

	return authorization.NewObjectRaw(targetURL, ctx.XForwardedMethod()), nil
}

func authzHandleUnauthorizedImplExtAuthz(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL) {
	var (
		statusCode int
	)

	switch {
	case ctx.IsXHR() || !ctx.AcceptsMIME("text/html"):
		statusCode = fasthttp.StatusUnauthorized
	default:
		switch authn.Object.Method {
		case fasthttp.MethodGet, fasthttp.MethodOptions, "":
			statusCode = fasthttp.StatusFound
		default:
			statusCode = fasthttp.StatusSeeOther
		}
	}

	ctx.Logger.Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d with location redirect to %s", authn.Object.String(), authn.Method, authn.Username, statusCode, redirectionURL)
	ctx.SpecialRedirect(redirectionURL.String(), statusCode)
}
