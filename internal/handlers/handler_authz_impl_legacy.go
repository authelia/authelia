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

	descriptor := "header 'X-Forwarded-Method'"

	if method = ctx.XForwardedMethod(); len(method) == 0 {
		method, descriptor = ctx.Method(), "start line value 'Method'"
	}

	if hasInvalidMethodCharacters(method) {
		return object, fmt.Errorf("%s with value '%s' has invalid characters", descriptor, method)
	}

	return authorization.NewObjectRaw(targetURL, method), nil
}

func handleAuthzUnauthorizedLegacy(ctx AuthzContext, authn *Authn, redirectionURL *url.URL) {
	if authn.Type == AuthnTypeAuthorization {
		handleAuthzUnauthorizedAuthorizationBasic(ctx, authn)

		return
	}

	if redirectionURL == nil {
		ctx.GetLogger().Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d", authn.Object.String(), authn.Method, authn.Username, fasthttp.StatusUnauthorized)
		ctx.ReplyUnauthorized()

		return
	}

	doAuthzRedirect(ctx, authn, redirectionURL, getAuthzRedirectStatusCode(ctx, authn.Object.Method))
}
