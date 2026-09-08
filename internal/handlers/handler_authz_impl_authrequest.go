package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectAuthRequest(ctx AuthzContext) (object authorization.Object, err error) {
	var (
		method          []byte
		originalURL     []byte
		requestedObject *authorization.Object
	)

	if method = ctx.XOriginalMethod(); len(method) == 0 {
		return object, fmt.Errorf("header 'X-Original-Method' is empty")
	}

	if originalURL = ctx.XOriginalURL(); len(originalURL) == 0 {
		return object, fmt.Errorf("header 'X-Original-URL' is empty")
	}

	if requestedObject, err = authorization.NewObjectMethodURL(method, originalURL); err != nil {
		return object, err
	}

	return *requestedObject, nil
}

func handleAuthzUnauthorizedAuthRequest(ctx AuthzContext, authn *Authn, redirectionURL *url.URL) {
	doAuthzRedirect(ctx, authn, redirectionURL, fasthttp.StatusUnauthorized)
}
