package middleware

import (
	"github.com/valyala/fasthttp"
)

// Wrap a handler with another middleware if it isn't nil.
func Wrap(middleware Basic, next fasthttp.RequestHandler) (handler fasthttp.RequestHandler) {
	if middleware == nil {
		return next
	}

	return middleware(next)
}
