// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/valyala/fasthttp"
)

// Status handles basic status responses.
func Status(statusCode int) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		SetStatusCodeResponse(ctx, statusCode)
	}
}
