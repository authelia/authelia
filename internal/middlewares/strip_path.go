// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"strings"

	"github.com/valyala/fasthttp"
)

// StripPath strips the first level of a path.
func StripPath(path string) (middleware Middleware) {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			uri := ctx.RequestURI()

			if strings.HasPrefix(string(uri), path) {
				ctx.SetUserValueBytes(keyUserValueBaseURL, path)

				newURI := strings.TrimPrefix(string(uri), path)
				ctx.Request.SetRequestURI(newURI)
			}

			next(ctx)
		}
	}
}
