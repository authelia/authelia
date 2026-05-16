// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"bytes"
	"net"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewTrustedProxies returns a middleware that removes headers from requests where the client is not trusted as a
// proxy.
func NewTrustedProxies(config schema.ServerHeadersTrustedProxies) Middleware {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		if !config.Enable {
			return next
		}

		return func(ctx *fasthttp.RequestCtx) {
			if !isTrustedPeer(ctx, config.Networks) {
				deleteProxyHeaders(&ctx.Request.Header, ctx.Request.Header.PeekKeys())
				deleteProxyHeaders(&ctx.Request.Header, ctx.Request.Header.PeekTrailerKeys())
			}

			next(ctx)
		}
	}
}

// NewTrustedProxiesAuthz returns a middleware for an authz endpoint that removes the query argument the implementation
// uses to determine the portal URL from requests where the client is not trusted as a proxy. These arguments have other
// meanings outside of the authz endpoints so this must only be used with them. It returns nil if there is nothing to
// remove as the middleware is not enabled or the implementation does not use a query argument.
func NewTrustedProxiesAuthz(config schema.ServerHeadersTrustedProxies, implementation string) Basic {
	if !config.Enable {
		return nil
	}

	var arg []byte

	switch implementation {
	case schema.AuthzImplementationExtAuthz:
		return nil
	case schema.AuthzImplementationForwardAuth, schema.AuthzImplementationAuthRequest:
		arg = qryArgAutheliaURL
	default:
		arg = qryArgRedirect
	}

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if !isTrustedPeer(ctx, config.Networks) {
				ctx.QueryArgs().DelBytes(arg)
			}

			next(ctx)
		}
	}
}

// TrustedProxiesRemoteIP returns the remote IP of a request which has not passed through the [NewTrustedProxies]
// middleware, such as one which failed to be read by the server.
func TrustedProxiesRemoteIP(ctx *fasthttp.RequestCtx, config schema.ServerHeadersTrustedProxies) net.IP {
	switch {
	case !config.Enable:
		return RequestCtxRemoteIP(ctx)
	case !isTrustedPeer(ctx, config.Networks):
		return ctx.RemoteIP()
	default:
		return RequestCtxRemoteIP(ctx)
	}
}

// isTrustedPeer returns true if the peer of the connection is trusted as a proxy. A peer which isn't connected via TCP,
// such as one connected via a unix socket, doesn't have an IP to check and access is instead controlled by the socket.
func isTrustedPeer(ctx *fasthttp.RequestCtx, networks []*net.IPNet) bool {
	if _, ok := ctx.RemoteAddr().(*net.TCPAddr); !ok {
		return true
	}

	return containsIP(ctx.RemoteIP(), networks)
}

func deleteProxyHeaders(header *fasthttp.RequestHeader, keys [][]byte) {
	for _, key := range keys {
		if !proxyHeader(key) {
			continue
		}

		header.DelBytes(key)
	}
}

func proxyHeader(key []byte) bool {
	if bytes.HasPrefix(key, headerPrefixXForwarded) {
		return true
	}

	if bytes.HasPrefix(key, headerPrefixXOriginal) {
		return true
	}

	if bytes.HasPrefix(key, headerPrefixXReal) {
		return true
	}

	if bytes.EqualFold(key, headerForwarded) {
		return true
	}

	if bytes.EqualFold(key, headerVia) {
		return true
	}

	if bytes.EqualFold(key, headerXAutheliaURL) {
		return true
	}

	return false
}

func containsIP(ip net.IP, networks []*net.IPNet) bool {
	for _, network := range networks {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
