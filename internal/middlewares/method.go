// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

// NewMethodMaxLength returns a middleware which rejects requests with a method longer than the provided length with a
// 405 Method Not Allowed response.
func NewMethodMaxLength(length int) Basic {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if len(ctx.Method()) > length {
				SetContentTypeTextPlain(ctx)

				ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
				ctx.SetBodyString(fmt.Sprintf("%d %s", fasthttp.StatusMethodNotAllowed, fasthttp.StatusMessage(fasthttp.StatusMethodNotAllowed)))

				return
			}

			next(ctx)
		}
	}
}
