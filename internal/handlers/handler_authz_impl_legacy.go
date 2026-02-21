package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectLegacy(ctx AuthzContext) (object authorization.Object, err error) {
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

func handleAuthzUnauthorizedLegacy(ctx AuthzContext, authn *Authn, redirectionURL *url.URL) {
	var (
		statusCode int
	)

	if authn.Type == AuthnTypeAuthorization {
		handleAuthzUnauthorizedAuthorizationBasic(ctx, authn)

		return
	}

	switch {
	case ctx.IsXHR() || !ctx.AcceptsMIME("text/html") || redirectionURL == nil:
		statusCode = fasthttp.StatusUnauthorized
	default:
		switch authn.Object.Method {
		case fasthttp.MethodGet, fasthttp.MethodOptions, fasthttp.MethodHead, "":
			statusCode = fasthttp.StatusFound
		default:
			statusCode = fasthttp.StatusSeeOther
		}
	}

	if redirectionURL != nil {
		ctx.GetLogger().Infof(logFmtAuthzRedirect, authn.Object.URL.String(), authn.Method, authn.Username, statusCode, redirectionURL)

		switch authn.Object.Method {
		case fasthttp.MethodHead:
			ctx.SpecialRedirectNoBody(redirectionURL.String(), statusCode)
		default:
			ctx.SpecialRedirect(redirectionURL.String(), statusCode)
		}
	} else {
		ctx.GetLogger().Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d", authn.Object.URL.String(), authn.Method, authn.Username, statusCode)
		ctx.ReplyUnauthorized()
	}
}
