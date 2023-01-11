package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
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
	rawURL := ctx.QueryArgs().PeekBytes(qryArgAutheliaURL)
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

func handleAuthzObjectVerifyStandard(ctx *middlewares.AutheliaCtx, object authorization.Object) (err error) {
	if !utils.IsURISecure(object.URL) {
		return fmt.Errorf("target URL '%s' has an insecure scheme '%s', only the 'https' and 'wss' schemes are supported so session cookies can be transmitted securely", object.URL.String(), object.URL.Scheme)
	}

	if !isURLUnderProtectedDomain(object.URL, ctx.Configuration.Session.Domain) {
		return fmt.Errorf("target URL '%s' is not on a domain which is a direct subdomain of the configured session domain '%s'", object.URL.String(), ctx.Configuration.Session.Domain)
	}

	return nil
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

func handleAuthzUnauthorizedAuthorizationBasic(ctx *middlewares.AutheliaCtx, authn *Authn, _ *url.URL) {
	ctx.Logger.Infof("Access to '%s' is not authorized to user '%s', sending 401 response with WWW-Authenticate header requesting Basic scheme", authn.Object.URL.String(), authn.Username)
	ctx.Response.Header.SetBytesKV(headerWWWAuthenticate, headerValueAuthenticateBasic)

	ctx.ReplyUnauthorized()
}
