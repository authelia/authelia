package handlers

import (
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

func handleAuthzUnauthorizedCommon(ctx AuthzContext, authn *Authn, redirectionURL *url.URL) {
	doAuthzRedirect(ctx, authn, redirectionURL, getAuthzRedirectStatusCode(ctx, authn.Object.Method))
}

func handleAuthzPortalURLLegacy(ctx AuthzContext) (portalURL *url.URL, err error) {
	if portalURL, err = handleAuthzPortalURLFromQueryLegacy(ctx); err != nil || portalURL != nil {
		return portalURL, err
	}

	return handleAuthzPortalURLFromHeader(ctx)
}

func handleAuthzPortalURLFromHeader(ctx AuthzContext) (portalURL *url.URL, err error) {
	return parseAuthzPortalURL(ctx.XAutheliaURL())
}

func handleAuthzPortalURLFromQuery(ctx AuthzContext) (portalURL *url.URL, err error) {
	return parseAuthzPortalURL(ctx.QueryArgAutheliaURL())
}

func handleAuthzPortalURLFromQueryLegacy(ctx AuthzContext) (portalURL *url.URL, err error) {
	return parseAuthzPortalURL(ctx.GetRequestQueryArgValue(qryArgRD))
}

func handleAuthzAuthorizedStandard(ctx AuthzContext, authn *Authn) {
	ctx.ReplyStatusCode(fasthttp.StatusOK)

	if authn.Details.Username != "" {
		ctx.SetResponseHeaderValue(headerRemoteUser, authn.Details.Username)
		ctx.SetResponseHeaderValue(headerRemoteGroups, strings.Join(authn.Details.Groups, ","))
		ctx.SetResponseHeaderValue(headerRemoteName, authn.Details.DisplayName)

		switch len(authn.Details.Emails) {
		case 0:
			ctx.SetResponseHeaderValue(headerRemoteEmail, "")
		default:
			ctx.SetResponseHeaderValue(headerRemoteEmail, authn.Details.Emails[0])
		}
	}
}

func handleAuthzUnauthorizedAuthorizationBasic(ctx AuthzContext, authn *Authn) {
	ctx.GetLogger().Infof("Access to '%s' is not authorized to user '%s', sending 401 response with WWW-Authenticate header requesting Basic scheme", authn.Object.URL.String(), authn.Username)

	ctx.ReplyUnauthorized()

	ctx.SetResponseHeaderValueBytes(headerWWWAuthenticate, headerValueAuthenticateBasic)
}
