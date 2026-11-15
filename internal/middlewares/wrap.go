// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"bytes"
	"net"

	"github.com/valyala/fasthttp"
)

// Wrap a handler with another middleware if it isn't nil.
func Wrap(middleware Basic, next fasthttp.RequestHandler) (handler fasthttp.RequestHandler) {
	if middleware == nil {
		return next
	}

	return middleware(next)
}

// MultiWrap allows wrapping a handler with additional middlewares if they are not nil.
func MultiWrap(next fasthttp.RequestHandler, middlewares ...Basic) (handler fasthttp.RequestHandler) {
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] == nil {
			continue
		}

		next = middlewares[i](next)
	}

	return next
}

// RequestCtxRemoteIP returns the remote IP of the given request context. This is the left most value of the first
// X-Forwarded-For header that is not empty if it is a valid IP, otherwise it is the TCP remote IP.
func RequestCtxRemoteIP(ctx *fasthttp.RequestCtx) net.IP {
	for _, header := range ctx.Request.Header.PeekAll(fasthttp.HeaderXForwardedFor) {
		if len(bytes.TrimSpace(header)) == 0 {
			continue
		}

		value, _, _ := bytes.Cut(header, []byte(","))

		if ip := net.ParseIP(string(bytes.TrimSpace(value))); ip != nil {
			return ip
		}

		break
	}

	return ctx.RemoteIP()
}
