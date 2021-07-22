package middlewares

import (
	"net/url"

	"github.com/authelia/authelia/internal/utils"
)

// AutomaticCORSMiddleware automatically adds all relevant CORS headers to a request.
func AutomaticCORSMiddleware(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		corsOrigin := ctx.Request.Header.Peek(headerOrigin)

		if corsOrigin != nil {
			corsOriginURL, err := url.Parse(string(corsOrigin))

			if err == nil && corsOriginURL != nil && utils.IsRedirectionSafe(*corsOriginURL, ctx.Configuration.Session.Domain) {
				ctx.Response.Header.SetBytesV(headerAccessControlAllowOrigin, corsOrigin)
				ctx.Response.Header.Set(headerVary, "Accept-Encoding, Origin")
				ctx.Response.Header.Set(headerAccessControlAllowCredentials, "false")

				corsHeaders := ctx.Request.Header.Peek(headerAccessControlRequestHeaders)
				if corsHeaders != nil {
					ctx.Response.Header.SetBytesV(headerAccessControlAllowHeaders, corsHeaders)
				}

				corsMethod := ctx.Request.Header.Peek(headerAccessControlRequestMethod)
				if corsHeaders != nil {
					ctx.Response.Header.SetBytesV(headerAccessControlAllowMethods, corsMethod)
				} else {
					ctx.Response.Header.Set(headerAccessControlAllowMethods, "GET")
				}
			}
		}

		next(ctx)
	}
}
