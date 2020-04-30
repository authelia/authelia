package server

import (
	"net"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/logging"
)

// Replacement for the default error handler in fasthttp.
func autheliaErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	if _, ok := err.(*fasthttp.ErrSmallBuffer); ok {
		// Note: Getting X-Forwarded-For or Request URI is impossible for ths error.
		logging.Logger().Tracef("Request was too large to handle from client %s. Response Code %d.", ctx.RemoteIP().String(), fasthttp.StatusRequestHeaderFieldsTooLarge)
		ctx.Error("Request header too large", fasthttp.StatusRequestHeaderFieldsTooLarge)
	} else if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
		// TODO: Add X-Forwarded-For Check here.
		logging.Logger().Tracef("Request timeout occurred while handling from client %s: %s. Response Code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusRequestTimeout)
		ctx.Error("Request timeout", fasthttp.StatusRequestTimeout)
	} else {
		// TODO: Add X-Forwarded-For Check here.
		logging.Logger().Tracef("An unknown error occurred while handling a request from client %s: %s. Response Code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusBadRequest)
		ctx.Error("Error when parsing request", fasthttp.StatusBadRequest)
	}
}
