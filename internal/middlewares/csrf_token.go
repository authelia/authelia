// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"errors"

	"github.com/authelia/authelia/v4/internal/session"
)

// The RequireCSRFToken middleware rejects requests using an unsafe method which don't carry the CSRF token of the
// established session of the request in the X-CSRF-Token request header. The token is only known to the user agent which
// holds the session, so unlike the CrossSiteRequestForgery middleware this doesn't rely on the user agent reporting where
// the request originated.
//
// A request without an established session is permitted, as a forged request can only act on behalf of a user by way of
// the session cookie their user agent attaches to it. Such a request is left to the CrossSiteRequestForgery middleware,
// and to the endpoint itself which is expected to reject it if it requires a session.
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

		switch err = provider.VerifyCSRFToken(ctx, string(ctx.Request.Header.PeekBytes(headerXCSRFToken))); {
		case err == nil, errors.Is(err, session.ErrCSRFTokenNoSession):
			next(ctx)
		case errors.Is(err, session.ErrCSRFTokenInvalid):
			ctx.Logger.Warn("Request rejected as it did not include a valid CSRF token")

			ctx.ReplyForbidden()
		default:
			ctx.Logger.WithError(err).Error("Error occurred verifying the CSRF token")

			ctx.ReplyForbidden()
		}
	}
}

// The RequireCSRFTokenBrowser middleware is the RequireCSRFToken middleware except it only requires the CSRF token from
// requests made by a browser, which are those carrying either the Sec-Fetch-Site or the Origin request header. A forged
// request can only be made by a browser and every browser this protects sends at least one of these headers on requests
// using an unsafe method, which a cross-site attacker can't prevent. This permits clients other than the portal which
// interact with the API on behalf of a user, such as the PAM module, without the CSRF token. It's only appropriate for
// endpoints where a forged request can't benefit the attacker, as the token no longer protects them should a reverse
// proxy strip both headers.
func RequireCSRFTokenBrowser(next RequestHandler) RequestHandler {
	required := RequireCSRFToken(next)

	return func(ctx *AutheliaCtx) {
		if len(ctx.Request.Header.PeekBytes(headerSecFetchSite)) == 0 && len(ctx.Request.Header.PeekBytes(headerOrigin)) == 0 {
			next(ctx)

			return
		}

		required(ctx)
	}
}
