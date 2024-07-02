package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

func handleAuthzPortalURLLegacy(ctx *middlewares.AutheliaCtx) (portalURL *url.URL, err error) {
	if portalURL, err = handleAuthzPortalURLFromQueryLegacy(ctx); err != nil || portalURL != nil {
		return portalURL, err
	}

	return handleAuthzPortalURLFromHeader(ctx)
}

func handleAuthzPortalURLFromHeader(ctx *middlewares.AutheliaCtx) (portalURL *url.URL, err error) {
	rawURL := ctx.XAutheliaURL()
	if rawURL == nil {
		return nil, nil
	}

	if portalURL, err = url.ParseRequestURI(string(rawURL)); err != nil {
		return nil, err
	}

	return portalURL, nil
}

func handleAuthzPortalURLFromQuery(ctx *middlewares.AutheliaCtx) (portalURL *url.URL, err error) {
	rawURL := ctx.QueryArgAutheliaURL()
	if rawURL == nil {
		return nil, nil
	}

	if portalURL, err = url.ParseRequestURI(string(rawURL)); err != nil {
		return nil, err
	}

	return portalURL, nil
}

func handleAuthzPortalURLFromQueryLegacy(ctx *middlewares.AutheliaCtx) (portalURL *url.URL, err error) {
	rawURL := ctx.QueryArgs().PeekBytes(qryArgRD)
	if rawURL == nil {
		return nil, nil
	}

	if portalURL, err = url.ParseRequestURI(string(rawURL)); err != nil {
		return nil, err
	}

	return portalURL, nil
}

func handleAuthzAuthorizedStandard(ctx *middlewares.AutheliaCtx, authn *Authn) {
	ctx.ReplyStatusCode(fasthttp.StatusOK)

	if authn.Details.Username != "" {
		ctx.Response.Header.SetBytesK(headerRemoteUser, authn.Details.Username)
		ctx.Response.Header.SetBytesK(headerRemoteGroups, strings.Join(authn.Details.Groups, ","))
		ctx.Response.Header.SetBytesK(headerRemoteName, authn.Details.DisplayName)

		switch len(authn.Details.Emails) {
		case 0:
			ctx.Response.Header.SetBytesK(headerRemoteEmail, "")
		default:
			ctx.Response.Header.SetBytesK(headerRemoteEmail, authn.Details.Emails[0])
		}
	}
}

func handleAuthzUnauthorizedAuthorizationBasic(ctx *middlewares.AutheliaCtx, authn *Authn) {
	ctx.Logger.Infof("Access to '%s' is not authorized to user '%s', sending 401 response with WWW-Authenticate header requesting Basic scheme", authn.Object.URL.String(), authn.Username)

	ctx.ReplyUnauthorized()

	ctx.Response.Header.SetBytesKV(headerWWWAuthenticate, headerValueAuthenticateBasic)
}

var protoHostSeparator = []byte("://")

func getRequestURIFromForwardedHeaders(protocol, host, uri []byte) (requestURI *url.URL, err error) {
	if len(protocol) == 0 {
		return nil, fmt.Errorf("missing protocol value")
	}

	if len(host) == 0 {
		return nil, fmt.Errorf("missing host value")
	}

	value := utils.BytesJoin(protocol, protoHostSeparator, host, uri)

	if requestURI, err = url.ParseRequestURI(string(value)); err != nil {
		return nil, fmt.Errorf("failed to parse forwarded headers: %w", err)
	}

	return requestURI, nil
}

func hasInvalidMethodCharacters(v []byte) bool {
	for _, c := range v {
		if c < 0x41 || c > 0x5A {
			return true
		}
	}

	return false
}

func isNotRequestingWebpage(ctx *middlewares.AutheliaCtx) bool {
	return ctx.IsXHR() || !ctx.AcceptsMIME("text/html")
}

func deriveStatusCodeFromAuthnMethod(authn *Authn) (statusCode int) {
	switch authn.Object.Method {
	case fasthttp.MethodGet, fasthttp.MethodOptions, fasthttp.MethodHead:
		return fasthttp.StatusFound
	default:
		return fasthttp.StatusSeeOther
	}
}

func handleAuthzSpecialRedirect(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL, statusCode int) {
	ctx.Logger.Infof(logFmtAuthzRedirect, authn.Object.String(), authn.Method, authn.Username, statusCode, redirectionURL)

	switch authn.Object.Method {
	case fasthttp.MethodHead:
		ctx.SpecialRedirectNoBody(redirectionURL.String(), statusCode)
	default:
		ctx.SpecialRedirect(redirectionURL.String(), statusCode)
	}
}
