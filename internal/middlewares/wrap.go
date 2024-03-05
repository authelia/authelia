package middlewares

import (
	"net"
	"strings"

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

func RequestCtxRemoteIP(ctx *fasthttp.RequestCtx) net.IP {
	if header := ctx.Request.Header.PeekBytes(headerXForwardedFor); len(header) != 0 {
		ips := strings.SplitN(string(header), ",", 2)

		if len(ips) != 0 {
			if ip := net.ParseIP(strings.Trim(ips[0], " ")); ip != nil {
				return ip
			}
		}
	}

	return ctx.RemoteIP()
}
