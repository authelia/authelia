package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func handleAuthzGetObjectAuthRequest(ctx *middlewares.AutheliaCtx) (object authorization.Object, err error) {
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

func handleAuthzUnauthorizedAuthRequest(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL) {
	handleAuthzSpecialRedirect(ctx, authn, redirectionURL, fasthttp.StatusUnauthorized)
}

func handleAuthzForbiddenAuthRequest(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL) {
	handleAuthzSpecialRedirect(ctx, authn, redirectionURL, fasthttp.StatusForbidden)
}
