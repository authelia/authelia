package middlewares

import (
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

// CORSApplyAutomaticAllowAllPolicy applies a CORS policy that automatically grants all Origins as well
// as all Request Headers other than Cookie and *. It does not allow credentials, and has a max age of 100. Vary is applied
// to both Accept-Encoding and Origin. It grants the GET Request Method only.
func CORSApplyAutomaticAllowAllPolicy(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if origin := ctx.Request.Header.PeekBytes(headerOrigin); origin != nil {
			corsApplyAutomaticAllowAllPolicy(&ctx.Request, &ctx.Response, origin)
		}

		next(ctx)
	}
}

func corsApplyAutomaticAllowAllPolicy(req *fasthttp.Request, resp *fasthttp.Response, origin []byte) {
	originURL, err := url.Parse(string(origin))
	if err != nil || originURL.Scheme != "https" {
		return
	}

	resp.Header.SetBytesKV(headerVary, headerValueVary)
	resp.Header.SetBytesKV(headerAccessControlAllowOrigin, origin)
	resp.Header.SetBytesKV(headerAccessControlAllowCredentials, headerValueFalse)
	resp.Header.SetBytesKV(headerAccessControlMaxAge, headerValueMaxAge)

	if headers := req.Header.PeekBytes(headerAccessControlRequestHeaders); headers != nil {
		requestedHeaders := strings.Split(string(headers), ",")
		allowHeaders := make([]string, len(requestedHeaders))

		for i, header := range requestedHeaders {
			headerTrimmed := strings.Trim(header, " ")
			if !strings.EqualFold("*", headerTrimmed) && !strings.EqualFold("Cookie", headerTrimmed) {
				allowHeaders[i] = headerTrimmed
			}
		}

		if len(allowHeaders) != 0 {
			resp.Header.SetBytesKV(headerAccessControlAllowHeaders, []byte(strings.Join(allowHeaders, ", ")))
		}
	}

	if requestMethods := req.Header.PeekBytes(headerAccessControlRequestMethod); requestMethods != nil {
		resp.Header.SetBytesKV(headerAccessControlAllowMethods, requestMethods)
	}
}
