package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectAuthRequest(ctx AuthzContext) (object authorization.Object, err error) {
	var requestedObject *authorization.Object

	if requestedObject, err = authorization.NewObjectMethodURL(ctx.XOriginalMethod(), ctx.XOriginalURL()); err != nil {
		return object, fmt.Errorf("failed to parse X-Original-URL header: %w", err)
	}

	return *requestedObject, nil
}

func handleAuthzUnauthorizedAuthRequest(ctx AuthzContext, authn *Authn, redirectionURL *url.URL) {
	ctx.GetLogger().Infof(logFmtAuthzRedirect, authn.Object.URL.String(), authn.Method, authn.Username, fasthttp.StatusUnauthorized, redirectionURL)

	switch authn.Object.Method {
	case fasthttp.MethodHead:
		ctx.SpecialRedirectNoBody(redirectionURL.String(), fasthttp.StatusUnauthorized)
	default:
		ctx.SpecialRedirect(redirectionURL.String(), fasthttp.StatusUnauthorized)
	}
}
