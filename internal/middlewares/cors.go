package middlewares

import (
	"net/url"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// CORS generates a middleware for adding CORS headers.
func CORS(config *schema.CORSConfiguration) Middleware {
	if config == nil || !config.Enable {
		return func(next RequestHandler) RequestHandler {
			return func(ctx *AutheliaCtx) {
				next(ctx)
			}
		}
	}

	headerOriginBytes := []byte(headerOrigin)
	processCORS := generateProcessCORS(config)

	return func(next RequestHandler) RequestHandler {
		return func(ctx *AutheliaCtx) {
			corsOrigin := ctx.Request.Header.PeekBytes(headerOriginBytes)

			if corsOrigin != nil {
				corsOriginURL, err := url.Parse(string(corsOrigin))
				if err == nil && corsOriginURL != nil && corsOriginURL.Scheme == "https" {
					processCORS(ctx, config, corsOrigin, *corsOriginURL)
				}
			}

			next(ctx)
		}
	}
}

func generateProcessCORS(config *schema.CORSConfiguration) func(ctx *AutheliaCtx, config *schema.CORSConfiguration, corsOrigin []byte, corsOriginURL url.URL) {
	var (
		headerVaryBytes                          = []byte(headerVary)
		headerAccessControlRequestHeadersBytes   = []byte(headerAccessControlRequestHeaders)
		headerAccessControlRequestMethodBytes    = []byte(headerAccessControlRequestMethod)
		headerAccessControlAllowOriginBytes      = []byte(headerAccessControlAllowOrigin)
		headerAccessControlAllowCredentialsBytes = []byte(headerAccessControlAllowCredentials)
		headerAccessControlAllowHeadersBytes     = []byte(headerAccessControlAllowHeaders)
		headerAccessControlAllowMethodsBytes     = []byte(headerAccessControlAllowMethods)
		headerAccessControlMaxAgeBytes           = []byte(headerAccessControlMaxAge)

		falseBytes   = []byte("false")
		maxAgeBytes  = []byte(string(rune(config.MaxAge)))
		varyBytes    = []byte(strings.Join(config.Vary, ", "))
		methodsBytes []byte
		headersBytes []byte
	)

	if len(config.Methods) != 0 {
		methodsBytes = []byte(strings.Join(config.Methods, ", "))
	}

	if len(config.Headers) != 0 {
		headersBytes = []byte(strings.Join(config.Headers, ", "))
	}

	return func(ctx *AutheliaCtx, config *schema.CORSConfiguration, corsOrigin []byte, corsOriginURL url.URL) {
		if corsOriginURL.Path != "" || corsOriginURL.RawPath != "" {
			return
		}

		origins := len(config.Origins)

		switch {
		case origins == 0 && !utils.IsRedirectionSafe(corsOriginURL, ctx.Configuration.Session.Domain):
			return
		case origins != 0 && config.Origins[0] != "*" &&
			(!config.IncludeProtected || !utils.IsRedirectionSafe(corsOriginURL, ctx.Configuration.Session.Domain)) &&
			!utils.IsStringInSliceFold(corsOriginURL.String(), config.Origins):
			return
		}

		ctx.Response.Header.SetBytesKV(headerAccessControlAllowOriginBytes, corsOrigin)
		ctx.Response.Header.SetBytesKV(headerVaryBytes, varyBytes)
		ctx.Response.Header.SetBytesKV(headerAccessControlAllowCredentialsBytes, falseBytes)
		ctx.Response.Header.SetBytesKV(headerAccessControlMaxAgeBytes, maxAgeBytes)

		if len(methodsBytes) != 0 {
			ctx.Response.Header.SetBytesKV(headerAccessControlAllowMethodsBytes, methodsBytes)
		} else {
			ctx.Response.Header.SetBytesKV(headerAccessControlAllowMethodsBytes, ctx.Request.Header.PeekBytes(headerAccessControlRequestMethodBytes))
		}

		if len(headersBytes) != 0 {
			ctx.Response.Header.SetBytesKV(headerAccessControlAllowHeadersBytes, headersBytes)
		} else {
			ctx.Response.Header.SetBytesKV(headerAccessControlAllowHeadersBytes, ctx.Request.Header.PeekBytes(headerAccessControlRequestHeadersBytes))
		}
	}
}
