package server

import (
	"net"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/logging"
)

// Replacement for the default error handler in fasthttp.
func handlerErrors(ctx *fasthttp.RequestCtx, err error) {
	logger := logging.Logger()

	switch e := err.(type) {
	case *fasthttp.ErrSmallBuffer:
		logger.Tracef("Request was too large to handle from client %s. Response Code %d.", ctx.RemoteIP().String(), fasthttp.StatusRequestHeaderFieldsTooLarge)
		ctx.Error("request header too large", fasthttp.StatusRequestHeaderFieldsTooLarge)
	case *net.OpError:
		if e.Timeout() {
			// TODO: Add X-Forwarded-For Check here.
			logger.Tracef("Request timeout occurred while handling from client %s: %s. Response Code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusRequestTimeout)
			ctx.Error("request timeout", fasthttp.StatusRequestTimeout)
		} else {
			// TODO: Add X-Forwarded-For Check here.
			logger.Tracef("An unknown error occurred while handling a request from client %s: %s. Response Code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusBadRequest)
			ctx.Error("error when parsing request", fasthttp.StatusBadRequest)
		}
	default:
		// TODO: Add X-Forwarded-For Check here.
		logger.Tracef("An unknown error occurred while handling a request from client %s: %s. Response Code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusBadRequest)
		ctx.Error("error when parsing request", fasthttp.StatusBadRequest)
	}
}

func handlerNotFound(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := strings.ToLower(string(ctx.Path()))

		for i := 0; i < len(httpServerDirs); i++ {
			if path == httpServerDirs[i].name || strings.HasPrefix(path, httpServerDirs[i].prefix) {
				handlers.SetStatusCodeResponse(ctx, fasthttp.StatusNotFound)

				return
			}
		}

		next(ctx)
	}
}

func handlerMethodNotAllowed(ctx *fasthttp.RequestCtx) {
	handlers.SetStatusCodeResponse(ctx, fasthttp.StatusMethodNotAllowed)
}
