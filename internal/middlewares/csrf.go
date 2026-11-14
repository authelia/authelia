// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"bytes"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/utils"
)

// The CrossSiteRequestForgery middleware rejects requests using an unsafe method which the browser indicates did not
// originate from the same origin. The Sec-Fetch-Site request header is used when it's available, falling back to
// comparing the Origin request header to the origin of the request. Requests which include neither header were not
// made by a browser capable of making cross-origin requests on behalf of a user and are permitted.
//
// Unlike the SameSite cookie attribute this rejects requests from other origins which are the same site, such as other
// subdomains of the session cookie domain.
func CrossSiteRequestForgery(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if isSafeMethod(ctx) || isSameOriginRequest(ctx) {
			next(ctx)

			return
		}

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetContentTypeBytes(contentTypeTextPlain)
		ctx.SetBodyString(fmt.Sprintf("%d %s", fasthttp.StatusForbidden, fasthttp.StatusMessage(fasthttp.StatusForbidden)))
	}
}

func isSafeMethod(ctx *fasthttp.RequestCtx) bool {
	return ctx.IsGet() || ctx.IsHead() || ctx.IsOptions() || ctx.IsTrace()
}

func isSameOriginRequest(ctx *fasthttp.RequestCtx) bool {
	if site := ctx.Request.Header.PeekBytes(headerSecFetchSite); len(site) != 0 {
		return bytes.Equal(site, headerValueSameOrigin) || bytes.Equal(site, headerValueNone)
	}

	origin := ctx.Request.Header.PeekBytes(headerOrigin)

	if len(origin) == 0 {
		return true
	}

	proto := ctx.Request.Header.PeekBytes(headerXForwardedProto)

	if len(proto) == 0 {
		if ctx.IsTLS() {
			proto = protoHTTPS
		} else {
			proto = protoHTTP
		}
	}

	host := ctx.Request.Header.PeekBytes(headerXForwardedHost)

	if len(host) == 0 {
		host = ctx.Host()
	}

	return bytes.EqualFold(origin, utils.BytesJoin(proto, protoHostSeparator, host))
}
