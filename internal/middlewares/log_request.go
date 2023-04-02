// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/valyala/fasthttp"
)

// LogRequest provides trace logging for all requests.
func LogRequest(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		autheliaCtx := &AutheliaCtx{RequestCtx: ctx}
		logger := NewRequestLogger(autheliaCtx)

		logger.Trace("Request hit")
		next(ctx)
		logger.Tracef("Replied (status=%d)", ctx.Response.StatusCode())
	}
}
