package middlewares

import (
	"bytes"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/utils"
)

// AllowedHosts checks the host header to ensure it's coming from an expected host.
func AllowedHosts(hosts []string) Middleware {
	hosts = utils.StringSliceRemoveDuplicates(hosts)

	length := len(hosts)

	if length == 0 {
		return nil
	}

	allowed := make([][]byte, length)

	for i := 0; i < length; i++ {
		allowed[i] = []byte(hosts[i])
	}

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			host := ctx.Host()

			for i := 0; i < length; i++ {
				if bytes.Equal(host, allowed[i]) {
					next(ctx)

					return
				}
			}

			SetStatusCodeResponse(ctx, fasthttp.StatusNotFound)
		}
	}
}
