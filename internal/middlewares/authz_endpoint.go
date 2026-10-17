// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/valyala/fasthttp"
)

// AuthzEndpoint marks every request handled by next as a request to an authz endpoint, which decides which session
// cookie is honored for it: the authz cookie is honored only on requests marked this way, and the session cookie only
// on requests which aren't. It must wrap the handler where the authz endpoint routes are registered so that every
// middleware which reads the session observes the mark.
func AuthzEndpoint(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetUserValue(UserValueKeyAuthzEndpoint, true)

		next(ctx)
	}
}
