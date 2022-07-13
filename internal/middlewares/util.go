package middlewares

import (
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
