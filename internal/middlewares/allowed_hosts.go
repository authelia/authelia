// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/logging"
)

// AllowedHosts returns a middleware which ensures the host header of the request is one of the provided hosts, otherwise
// it logs the rejection at the debug level and responds with a 404 Not Found. It returns nil if no hosts are provided.
func AllowedHosts(hosts []string) (middleware Basic) {
	if len(hosts) == 0 {
		return nil
	}

	normalized := make([]string, len(hosts))

	for i, host := range hosts {
		normalized[i] = strings.ToLower(host)
	}

	normalized = slices.Compact(slices.Sorted(slices.Values(normalized)))

	allowed := make([][]byte, len(normalized))

	for i, host := range normalized {
		allowed[i] = []byte(host)
	}

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			host := ctx.Host()

			for _, a := range allowed {
				if bytes.EqualFold(host, a) {
					next(ctx)

					return
				}
			}

			NewRequestLogger(ctx).WithField(logging.FieldHost, string(host)).Debug("Request rejected as the host is not one of the allowed hosts")

			ctx.Response.Reset()

			SetContentTypeTextPlain(ctx)

			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetBodyString(fmt.Sprintf("%d %s", fasthttp.StatusNotFound, fasthttp.StatusMessage(fasthttp.StatusNotFound)))
		}
	}
}
