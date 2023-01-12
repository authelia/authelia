package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func handleAuthzGetObjectExtAuthz(ctx *middlewares.AutheliaCtx) (object authorization.Object, err error) {
	protocol, host, uri := ctx.XForwardedProto(), ctx.GetXForwardedHost(), ctx.XForwardedURI()

	if uri == nil {
		uri = ctx.AuthzPath()
	}

	var targetURL *url.URL

	if targetURL, err = getRequestURIFromForwardedHeaders(protocol, host, uri); err != nil {
		return object, fmt.Errorf("failed to get target URL: %w", err)
	}

	method := ctx.XForwardedMethod()

	if len(method) == 0 {
		method = ctx.Method()
	}

	if len(method) == 0 {
		return object, fmt.Errorf("start line value 'Method' is empty")
	}

	if hasInvalidMethodCharacters(method) {
		return object, fmt.Errorf("start line value 'Method' with value '%s' has invalid characters", method)
	}

	return authorization.NewObjectRaw(targetURL, method), nil
}

func handleAuthzUnauthorizedExtAuthz(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL) {
	var (
		statusCode int
	)

	switch {
	case ctx.IsXHR() || !ctx.AcceptsMIME("text/html"):
		statusCode = fasthttp.StatusUnauthorized
	default:
		switch authn.Object.Method {
		case fasthttp.MethodGet, fasthttp.MethodOptions, "":
			statusCode = fasthttp.StatusFound
		default:
			statusCode = fasthttp.StatusSeeOther
		}
	}

	ctx.Logger.Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d with location redirect to %s", authn.Object.String(), authn.Method, authn.Username, statusCode, redirectionURL)
	ctx.SpecialRedirect(redirectionURL.String(), statusCode)
}
