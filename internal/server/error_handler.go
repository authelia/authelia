package server

import (
	"net"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/logging"
)

// Replacement for the default error handler in fasthttp.
func autheliaErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	if _, ok := err.(*fasthttp.ErrSmallBuffer); ok {
		logging.Logger().Tracef("Request was too large to handle from client %s: %s. Response Code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusRequestHeaderFieldsTooLarge)
		ctx.Error("Request header too large", fasthttp.StatusRequestHeaderFieldsTooLarge)
	} else if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
		logging.Logger().Tracef("Request timeout occurred while handling from client %s: %s. Response Code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusRequestTimeout)
		ctx.Error("Request timeout", fasthttp.StatusRequestTimeout)
	} else {
		logging.Logger().Tracef("An unknown error occurred while handling from client %s: %s. Response Code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusBadRequest)
		ctx.Error("Error when parsing request", fasthttp.StatusBadRequest)
	}
}
