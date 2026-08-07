package handlers

import (
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectForwardAuth(ctx AuthzContext) (object authorization.Object, err error) {
	var requestedObject *authorization.Object

	if requestedObject, err = authorization.NewObjectMethodSchemeHostPath(ctx.XForwardedMethod(), ctx.XForwardedProto(), ctx.XForwardedHost(), ctx.XForwardedURI()); err != nil {
		return object, err
	}

	return *requestedObject, nil
}

func handleAuthzUnauthorizedForwardAuth(ctx AuthzContext, authn *Authn, redirectionURL *url.URL) {
	var (
		statusCode int
	)

	switch {
	case ctx.IsXHR() || !ctx.AcceptsMIME("text/html"):
		statusCode = fasthttp.StatusUnauthorized
	default:
		switch authn.Object.Method {
		case fasthttp.MethodGet, fasthttp.MethodOptions, fasthttp.MethodHead:
			statusCode = fasthttp.StatusFound
		default:
			statusCode = fasthttp.StatusSeeOther
		}
	}

	ctx.GetLogger().Infof(logFmtAuthzRedirect, authn.Object.String(), authn.Method, authn.Username, statusCode, redirectionURL)

	switch authn.Object.Method {
	case fasthttp.MethodHead:
		ctx.SpecialRedirectNoBody(redirectionURL.String(), statusCode)
	default:
		ctx.SpecialRedirect(redirectionURL.String(), statusCode)
	}
}
