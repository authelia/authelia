package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func handleAuthzGetObjectLegacy(ctx *middlewares.AutheliaCtx) (object authorization.Object, err error) {
	var (
		targetURL *url.URL
		method    []byte
	)

	if targetURL, err = ctx.GetXOriginalURLOrXForwardedURL(); err != nil {
		return object, fmt.Errorf("failed to get target URL: %w", err)
	}

	if method = ctx.XForwardedMethod(); len(method) == 0 {
		method = ctx.Method()
	}

	if hasInvalidMethodCharacters(method) {
		return object, fmt.Errorf("header 'X-Forwarded-Method' with value '%s' has invalid characters", method)
	}

	return authorization.NewObjectRaw(targetURL, method), nil
}

func handleAuthzUnauthorizedLegacy(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL) {
	var (
		statusCode int
	)

	if authn.Type == AuthnTypeAuthorization {
		handleAuthzUnauthorizedAuthorizationBasic(ctx, authn)
		return
	}

	switch {
	case isNotRequestingWebpage(ctx) || redirectionURL == nil:
		statusCode = fasthttp.StatusUnauthorized
	default:
		if authn.Object.Method == "" {
			statusCode = fasthttp.StatusFound
		} else {
			statusCode = deriveStatusCodeFromAuthnMethod(authn)
		}
	}

	if redirectionURL != nil {
		handleAuthzSpecialRedirect(ctx, authn, redirectionURL, statusCode)
	} else {
		ctx.Logger.Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d", authn.Object.URL.String(), authn.Method, authn.Username, statusCode)
		ctx.ReplyUnauthorized()
	}
}

func handleAuthzForbiddenLegacy(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL) {
	var (
		statusCode int
	)

	if authn.Type == AuthnTypeAuthorization {
		ctx.Logger.Infof("Access to %s (method %s) is forbidden for user %s, responding with status code %d", authn.Object.URL.String(), authn.Method, authn.Username, fasthttp.StatusForbidden)
		ctx.ReplyForbidden()
		return
	}

	switch {
	case isNotRequestingWebpage(ctx) || redirectionURL == nil:
		statusCode = fasthttp.StatusForbidden
	default:
		if authn.Object.Method == "" {
			statusCode = fasthttp.StatusFound
		} else {
			statusCode = deriveStatusCodeFromAuthnMethod(authn)
		}
	}

	if redirectionURL != nil {
		handleAuthzSpecialRedirect(ctx, authn, redirectionURL, statusCode)
	} else {
		ctx.Logger.Infof("Access to %s (method %s) is forbidden for user %s, responding with status code %d", authn.Object.URL.String(), authn.Method, authn.Username, statusCode)
		ctx.ReplyForbidden()
	}
}
