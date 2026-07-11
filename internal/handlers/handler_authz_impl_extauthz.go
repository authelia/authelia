package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectExtAuthz(ctx AuthzContext) (object authorization.Object, err error) {
	var requestedObject *authorization.Object

	host := ctx.Host()

	if requestedObject, err = authorization.NewObjectMethodSchemeHostPath(ctx.Method(), ctx.XForwardedProto(), host, ctx.AuthzPath()); err != nil {
		return object, fmt.Errorf("failed to parse start line value 'Method': %w", err)
	}

	return *requestedObject, nil
}

func handleAuthzUnauthorizedExtAuthz(ctx AuthzContext, authn *Authn, redirectionURL *url.URL) {
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
