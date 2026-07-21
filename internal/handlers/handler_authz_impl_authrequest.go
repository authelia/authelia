package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func handleAuthzGetObjectAuthRequest(ctx AuthzContext) (object authorization.Object, err error) {
	var (
		targetURL *url.URL

		rawURL, method []byte
	)

	if rawURL = ctx.XOriginalURL(); len(rawURL) == 0 {
		return object, middlewares.ErrMissingXOriginalURL
	}

	if targetURL, err = url.ParseRequestURI(string(rawURL)); err != nil {
		return object, fmt.Errorf("failed to parse X-Original-URL header: %w", err)
	}

	if method = ctx.XOriginalMethod(); len(method) == 0 {
		return object, fmt.Errorf("header 'X-Original-Method' is empty")
	}

	if hasInvalidMethodCharacters(method) {
		return object, fmt.Errorf("header 'X-Original-Method' with value '%s' has invalid characters", method)
	}

	return authorization.NewObjectRaw(targetURL, method), nil
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
