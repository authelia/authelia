package server

import (
	"net"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/logging"
)

func autheliaFastHTTPErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	if _, ok := err.(*fasthttp.ErrSmallBuffer); ok {
		logging.Logger().Tracef("Request was too large (header size: %d, body size: %d) to handle from client %s: %s", ctx.Request.Header.Len(), len(ctx.Request.Body()), ctx.RemoteIP().String(), ctx.RequestURI())
		ctx.Error("Too big request header", fasthttp.StatusRequestHeaderFieldsTooLarge)
	} else if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
		logging.Logger().Tracef("Request timeout occurred while handling from client %s: %s", ctx.RemoteIP().String(), ctx.RequestURI())
		ctx.Error("Request timeout", fasthttp.StatusRequestTimeout)
	} else {
		logging.Logger().Tracef("An unknown error for a bad request occurred while handling from client %s: %s", ctx.RemoteIP().String(), ctx.RequestURI())
		ctx.Error("Error when parsing request", fasthttp.StatusBadRequest)
	}
}
