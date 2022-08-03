package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func authzGetObjectImplLegacy(ctx *middlewares.AutheliaCtx) (object authorization.Object, err error) {
	var targetURL *url.URL

	if targetURL, err = ctx.GetOriginalURL(); err != nil {
		return object, fmt.Errorf("failed to get target URL: %w", err)
	}

	return authorization.NewObjectRaw(targetURL, ctx.XForwardedMethod()), nil
}

func authzHandleUnauthorizedImplLegacy(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL) {
	var (
		statusCode int
	)

	if authn.Type == AuthnTypeAuthorization {
		authzHandleUnauthorizedAuthorizationBasic(ctx, authn, nil)

		return
	}

	switch {
	case ctx.IsXHR() || !ctx.AcceptsMIME("text/html") || redirectionURL == nil:
		statusCode = fasthttp.StatusUnauthorized
	default:
		switch authn.Object.Method {
		case fasthttp.MethodGet, fasthttp.MethodOptions, "":
			statusCode = fasthttp.StatusFound
		default:
			statusCode = fasthttp.StatusSeeOther
		}
	}

	if redirectionURL != nil {
		ctx.Logger.Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d with location redirect to %s", authn.Object.URL.String(), authn.Method, authn.Username, statusCode, redirectionURL.String())
		ctx.SpecialRedirect(redirectionURL.String(), statusCode)
	} else {
		ctx.Logger.Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d", authn.Object.URL.String(), authn.Method, authn.Username, statusCode)
		ctx.ReplyUnauthorized()
	}
}
