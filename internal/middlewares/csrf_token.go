// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

// The RequireCSRFToken middleware rejects requests using an unsafe method which don't carry the CSRF token bound to the
// session cookie of the request in the X-CSRF-Token request header. The token is only known to the user agent which
// holds the session, so unlike the CrossSiteRequestForgery middleware this doesn't rely on the user agent reporting where
// the request originated.
//
// A request without a session cookie is permitted, as a forged request can only act on behalf of a user by way of the
// session cookie their user agent attaches to it. Such a request is left to the CrossSiteRequestForgery middleware, and
// to the endpoint itself which is expected to reject it if it requires a session.
func RequireCSRFToken(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if isSafeMethod(ctx.RequestCtx) {
			next(ctx)

			return
		}

		provider, err := ctx.GetSessionProvider()
		if err != nil {
			ctx.Logger.WithError(err).Error("Error occurred retrieving the session provider to verify the CSRF token")

			ctx.ReplyForbidden()

			return
		}

		if len(provider.CSRFToken(ctx)) == 0 {
			next(ctx)

			return
		}

		if !provider.VerifyCSRFToken(ctx, string(ctx.Request.Header.PeekBytes(headerXCSRFToken))) {
			ctx.Logger.Warn("Request rejected as it did not include a valid CSRF token")

			ctx.ReplyForbidden()

			return
		}

		next(ctx)
	}
}
