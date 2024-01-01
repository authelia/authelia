package middlewares

import (
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"
)

var (
	// ErrMissingXForwardedProto is returned on methods which require an X-Forwarded-Proto header.
	ErrMissingXForwardedProto = errors.New("missing required X-Forwarded-Proto header")

	// ErrMissingXForwardedHost is returned on methods which require an X-Forwarded-Host header.
	ErrMissingXForwardedHost = errors.New("missing required X-Forwarded-Host header")

	// ErrMissingHeaderHost is returned on methods which require an Host header.
	ErrMissingHeaderHost = errors.New("missing required Host header")

	// ErrMissingXOriginalURL is returned on methods which require an X-Original-URL header.
	ErrMissingXOriginalURL = errors.New("missing required X-Original-URL header")
)

// RecoverPanic recovers from panics and logs the error.
func RecoverPanic(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				NewRequestLogger(ctx).WithError(recoverErr(r)).Error("Panic (recovered) occurred while handling requests, please report this error")

				ctx.Response.Reset()
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetContentTypeBytes(contentTypeTextPlain)
				ctx.SetBodyString(fmt.Sprintf("%d %s", fasthttp.StatusInternalServerError, fasthttp.StatusMessage(fasthttp.StatusInternalServerError)))
			}
		}()

		next(ctx)
	}
}

func recoverErr(i any) error {
	switch v := i.(type) {
	case nil:
		return nil
	case string:
		return fmt.Errorf("recovered panic: %s", v)
	case error:
		return fmt.Errorf("recovered panic: %w", v)
	default:
		return fmt.Errorf("recovered panic with unknown type: %v", v)
	}
}
