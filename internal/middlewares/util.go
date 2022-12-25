package middlewares

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

// SetContentTypeApplicationJSON sets the Content-Type header to `application/json; charset=utf8`.
func SetContentTypeApplicationJSON(ctx *fasthttp.RequestCtx) {
	ctx.SetContentTypeBytes(contentTypeApplicationJSON)
}

// SetContentTypeTextPlain sets the Content-Type header to `text/plain; charset=utf8`.
func SetContentTypeTextPlain(ctx *fasthttp.RequestCtx) {
	ctx.SetContentTypeBytes(contentTypeTextPlain)
}

// SetStatusCodeResponse writes a response status code and an appropriate body on either a
// *fasthttp.RequestCtx or *middlewares.AutheliaCtx.
func SetStatusCodeResponse(ctx *fasthttp.RequestCtx, statusCode int) {
	ctx.Response.Reset()
	ctx.SetContentTypeBytes(contentTypeTextPlain)
	ctx.SetStatusCode(statusCode)
	ctx.SetBodyString(fmt.Sprintf("%d %s", statusCode, fasthttp.StatusMessage(statusCode)))
}
