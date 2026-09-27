// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/valyala/fasthttp"
)

// AllowedHosts returns a middleware which ensures the host header of the request is one of the provided hosts, otherwise
// it responds with a 404 Not Found. It returns nil if no hosts are provided.
func AllowedHosts(hosts []string) (middleware Basic) {
	if len(hosts) == 0 {
		return nil
	}

	hosts = slices.Compact(slices.Sorted(slices.Values(hosts)))

	allowed := make([][]byte, len(hosts))

	for i, host := range hosts {
		allowed[i] = []byte(host)
	}

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			host := ctx.Host()

			for _, a := range allowed {
				if bytes.Equal(host, a) {
					next(ctx)

					return
				}
			}

			ctx.Response.Reset()

			SetContentTypeTextPlain(ctx)

			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetBodyString(fmt.Sprintf("%d %s", fasthttp.StatusNotFound, fasthttp.StatusMessage(fasthttp.StatusNotFound)))
		}
	}
}
