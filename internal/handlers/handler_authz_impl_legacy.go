package handlers

import (
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectLegacy(ctx AuthzContext) (object authorization.Object, err error) {
	var (
		method          []byte
		requestedObject *authorization.Object
	)

	if method = ctx.XForwardedMethod(); len(method) == 0 {
		method = ctx.Method()
	}

	if requestedObject, err = authorization.NewObjectMethodURLOrSchemeHostPath(method, ctx.XOriginalURL(), ctx.XForwardedProto(), ctx.GetXForwardedHost(), ctx.GetXForwardedURI()); err != nil {
		return object, err
	}

	return *requestedObject, nil
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
