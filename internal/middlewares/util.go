package middlewares

import (
	"github.com/valyala/fasthttp"
)

// SetContentTypeApplicationJSON sets the Content-Type header to `application/json; charset=utf-8`.
func SetContentTypeApplicationJSON(ctx *fasthttp.RequestCtx) {
	ctx.SetContentTypeBytes(contentTypeApplicationJSON)
}

// SetContentTypeTextPlain sets the Content-Type header to `text/plain; charset=utf-8`.
func SetContentTypeTextPlain(ctx *fasthttp.RequestCtx) {
	ctx.SetContentTypeBytes(contentTypeTextPlain)
}
