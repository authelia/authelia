// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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

func handleAuthzAuthorizedStandard(ctx AuthzContext, headers []AuthzHeader, authn *Authn) {
	ctx.ReplyStatusCode(fasthttp.StatusOK)

	if authn.Details.GetUsername() == "" {
		return
	}

	resolver, updated := ctx.GetProviderUserAttributeResolver(), ctx.GetClock().Now()

	for _, header := range headers {
		object, ok := resolver.Resolve(header.Attribute, authn.Details, updated)
		if !ok {
			ctx.SetResponseHeaderValue(header.Key, "")

			continue
		}

		ctx.SetResponseHeaderValue(header.Key, authzHeaderValue(object))
	}
}

func handleAuthzAuthorizedLegacy(ctx AuthzContext, _ []AuthzHeader, authn *Authn) {
	ctx.ReplyStatusCode(fasthttp.StatusOK)

	if authn.Details.GetUsername() != "" {
		ctx.SetResponseHeaderValue(headerRemoteUser, authn.Details.GetUsername())
		ctx.SetResponseHeaderValue(headerRemoteGroups, strings.Join(authn.Details.GetGroups(), ","))
		ctx.SetResponseHeaderValue(headerRemoteName, authn.Details.GetDisplayName())

		switch emails := authn.Details.GetEmails(); len(emails) {
		case 0:
			ctx.SetResponseHeaderValue(headerRemoteEmail, "")
		default:
			ctx.SetResponseHeaderValue(headerRemoteEmail, emails[0])
		}
	}
}

func handleAuthzUnauthorizedAuthorizationBasic(ctx AuthzContext, authn *Authn) {
	ctx.GetLogger().Infof("Access to '%s' is not authorized to user '%s', sending 401 response with WWW-Authenticate header requesting Basic scheme", authn.Object.URL.String(), authn.Username)

	ctx.ReplyUnauthorized()

	ctx.SetResponseHeaderValueBytes(headerWWWAuthenticate, headerValueAuthenticateBasic)
}
