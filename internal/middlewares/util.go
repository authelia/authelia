package middlewares

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

// SetStatusCodeResponse writes a response status code and an appropriate body on either a
// *fasthttp.RequestCtx or *middlewares.AutheliaCtx.
func SetStatusCodeResponse(ctx *fasthttp.RequestCtx, statusCode int) {
	ctx.Response.Reset()
	ctx.SetContentTypeBytes(contentTypeTextPlain)
	ctx.SetStatusCode(statusCode)
	ctx.SetBodyString(fmt.Sprintf("%d %s", statusCode, fasthttp.StatusMessage(statusCode)))
}
