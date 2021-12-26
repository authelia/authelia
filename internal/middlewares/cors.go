package middlewares

import (
	"net/url"
	"strings"
)

// CORSApplyAutomaticBasicPolicy applies a CORS policy that automatically grants all Origins as well
// as all Request Headers other than Cookie and *. It does not allow credentials, and has a max age of 100. Vary is applied
// to both Accept-Encoding and Origin. It grants the GET Request Method only.
func CORSApplyAutomaticBasicPolicy(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if origin := ctx.Request.Header.PeekBytes(headerOrigin); origin != nil {
			corsApplyAutomaticBasicPolicy(ctx, origin)
		}

		next(ctx)
	}
}

func corsApplyAutomaticBasicPolicy(ctx *AutheliaCtx, origin []byte) {
	originURL, err := url.Parse(string(origin))
	if err != nil || originURL.Scheme != "https" {
		return
	}

	ctx.Response.Header.SetBytesKV(headerVary, headerValueVary)
	ctx.Response.Header.SetBytesKV(headerAccessControlAllowOrigin, origin)
	ctx.Response.Header.SetBytesKV(headerAccessControlAllowCredentials, headerValueFalse)
	ctx.Response.Header.SetBytesKV(headerAccessControlMaxAge, headerValueMaxAge)

	if headers := ctx.Request.Header.PeekBytes(headerAccessControlRequestHeaders); headers != nil {
		requestedHeaders := strings.Split(string(headers), ",")
		finalHeaders := make([]string, len(requestedHeaders))

		for _, header := range requestedHeaders {
			headerTrimmed := strings.Trim(header, " ")
			if !strings.EqualFold("*", headerTrimmed) && !strings.EqualFold("Cookie", headerTrimmed) {
				finalHeaders = append(finalHeaders, headerTrimmed)
			}
		}

		if len(finalHeaders) != 0 {
			ctx.Response.Header.SetBytesKV(headerAccessControlAllowHeaders, []byte(strings.Join(finalHeaders, ", ")))
		}
	}

	ctx.Response.Header.SetBytesKV(headerAccessControlAllowMethods, headerValueMethodGET)
}
