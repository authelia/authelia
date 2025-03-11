package middlewares

import (
	"bytes"
	"net/url"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/utils"
)

// NewCORSPolicyBuilder returns a new CORSPolicyBuilder which is used to build a CORSPolicy which adds the Vary header
// with a value reflecting that the Origin header will Vary this response, then if the Origin header has a https scheme
// it makes the following additional adjustments: copies the Origin header to the Access-Control-Allow-Origin header
// effectively allowing all origins, sets the Access-Control-Allow-Credentials header to false which disallows CORS
// requests from sending cookies etc, sets the Access-Control-Allow-Headers header to the value specified by
// Access-Control-Request-Headers in the request excluding the Cookie/Authorization/Proxy-Authorization and special *
// values, sets Access-Control-Allow-Methods to the value specified by the Access-Control-Request-Method header, sets
// the Access-Control-Max-Age header to 100.
//
// These behaviours can be overridden by the With methods on the returned policy.
func NewCORSPolicyBuilder() (policy *CORSPolicyBuilder) {
	return &CORSPolicyBuilder{
		enabled: true,
		maxAge:  100,
	}
}

// CORSPolicyBuilder is a special middleware which provides CORS headers via handlers and middleware methods which can be
// configured. It aims to simplify CORS configurations.
type CORSPolicyBuilder struct {
	enabled     bool
	varyOnly    bool
	varySet     bool
	methods     []string
	headers     []string
	origins     []string
	credentials bool
	vary        []string
	maxAge      int
}

// Build reads the CORSPolicyBuilder configuration and generates a CORSPolicy.
func (b *CORSPolicyBuilder) Build() (policy *CORSPolicy) {
	policy = &CORSPolicy{
		enabled:     b.enabled,
		varyOnly:    b.varyOnly,
		credentials: []byte(strconv.FormatBool(b.credentials)),
		origins:     b.buildOrigins(),
		headers:     b.buildHeaders(),
		vary:        b.buildVary(),
	}

	if len(b.methods) != 0 {
		policy.methods = []byte(strings.Join(b.methods, ", "))
	}

	if b.maxAge <= 0 {
		policy.maxAge = headerValueMaxAge
	} else {
		policy.maxAge = []byte(strconv.Itoa(b.maxAge))
	}

	return policy
}

func (b CORSPolicyBuilder) buildOrigins() (origins [][]byte) {
	if len(b.origins) != 0 {
		if len(b.origins) == 1 && b.origins[0] == "*" {
			origins = append(origins, []byte(b.origins[0]))
		} else {
			for _, origin := range b.origins {
				origins = append(origins, []byte(origin))
			}
		}
	}

	return origins
}

func (b CORSPolicyBuilder) buildHeaders() (headers []byte) {
	if len(b.headers) != 0 {
		h := b.headers

		if b.credentials {
			if !utils.IsStringInSliceFold(fasthttp.HeaderCookie, h) {
				h = append(h, fasthttp.HeaderCookie)
			}

			if !utils.IsStringInSliceFold(fasthttp.HeaderAuthorization, h) {
				h = append(h, fasthttp.HeaderAuthorization)
			}

			if !utils.IsStringInSliceFold(fasthttp.HeaderProxyAuthorization, h) {
				h = append(h, fasthttp.HeaderProxyAuthorization)
			}
		}

		headers = utils.JoinAndCanonicalizeHeaders(headerSeparator, h...)
	}

	return headers
}

func (b CORSPolicyBuilder) buildVary() (vary []byte) {
	if b.varySet {
		if len(b.vary) != 0 {
			vary = utils.JoinAndCanonicalizeHeaders(headerSeparator, b.vary...)
		}
	} else {
		if len(b.origins) == 1 && b.origins[0] == "*" {
			vary = headerValueVaryWildcard
		} else {
			vary = headerValueVary
		}
	}

	return vary
}

// WithEnabled changes the enabled state of the middleware. If the middleware is initialized with NewCORSPolicyBuilder this
// value will be true but this function can override the value. Setting it to false prevents the middleware from adding
// any CORS headers. The only effect this middleware has after disabling this is the HandleOPTIONS and HandleOnlyOPTIONS
// handlers still function to return a HTTP 204 No Content, with the Allow header communicating the available HTTP
// method verbs. The main benefit of this option is that you don't have to implement complex logic to add/remove the
// middleware, you can just add it with the Middleware method, and adjust it using the WithEnabled method.
func (b *CORSPolicyBuilder) WithEnabled(enabled bool) (policy *CORSPolicyBuilder) {
	b.enabled = enabled

	return b
}

// WithAllowedMethods takes a list or HTTP methods and adjusts the Access-Control-Allow-Methods header to respond with
// that value.
func (b *CORSPolicyBuilder) WithAllowedMethods(methods ...string) (policy *CORSPolicyBuilder) {
	b.methods = methods

	return b
}

// WithAllowedOrigins takes a list of origin strings and only applies the CORS policy if the origin matches one of these.
func (b *CORSPolicyBuilder) WithAllowedOrigins(origins ...string) (policy *CORSPolicyBuilder) {
	b.origins = origins

	return b
}

// WithAllowedHeaders takes a list of header strings and alters the default Access-Control-Allow-Headers header.
func (b *CORSPolicyBuilder) WithAllowedHeaders(headers ...string) (policy *CORSPolicyBuilder) {
	b.headers = headers

	return b
}

// WithAllowCredentials takes bool and alters the default Access-Control-Allow-Credentials header.
func (b *CORSPolicyBuilder) WithAllowCredentials(allow bool) (policy *CORSPolicyBuilder) {
	b.credentials = allow

	return b
}

// WithVary takes a list of header strings and alters the default Vary header.
func (b *CORSPolicyBuilder) WithVary(headers ...string) (policy *CORSPolicyBuilder) {
	b.vary = headers
	b.varySet = true

	return b
}

// WithVaryOnly just adds the Vary header.
func (b *CORSPolicyBuilder) WithVaryOnly(varyOnly bool) (policy *CORSPolicyBuilder) {
	b.varyOnly = varyOnly

	return b
}

// WithMaxAge takes an integer and alters the default Access-Control-Max-Age header.
func (b *CORSPolicyBuilder) WithMaxAge(age int) (policy *CORSPolicyBuilder) {
	b.maxAge = age

	return b
}

// CORSPolicy is a middleware that handles adding CORS headers.
type CORSPolicy struct {
	enabled     bool
	varyOnly    bool
	methods     []byte
	headers     []byte
	origins     [][]byte
	credentials []byte
	vary        []byte
	maxAge      []byte
}

// HandleOPTIONS is an OPTIONS handler that just adds CORS headers, the Allow header, and sets the status code to 200
// without a body. This handler should generally not be used without using WithAllowedMethods.
func (p *CORSPolicy) HandleOPTIONS(ctx *fasthttp.RequestCtx) {
	p.handleOPTIONS(ctx)
	p.handle(ctx)
}

// HandleOnlyOPTIONS is an OPTIONS handler that just handles the Allow header, and sets the status code to 200
// without a body. This handler should generally not be used without using WithAllowedMethods.
func (p *CORSPolicy) HandleOnlyOPTIONS(ctx *fasthttp.RequestCtx) {
	p.handleOPTIONS(ctx)
}

// Middleware provides a middleware that adds the appropriate CORS headers for this CORSPolicyBuilder.
func (p *CORSPolicy) Middleware(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler) {
	return func(ctx *fasthttp.RequestCtx) {
		p.handle(ctx)

		next(ctx)
	}
}

func (p *CORSPolicy) handle(ctx *fasthttp.RequestCtx) {
	if !p.enabled {
		return
	}

	p.handleVary(ctx)

	if !p.varyOnly {
		p.handleCORS(ctx)
	}
}

func (p *CORSPolicy) handleOPTIONS(ctx *fasthttp.RequestCtx) {
	ctx.Response.ResetBody()

	/* The OPTIONS method should not return a 204 as per the following specifications when read together:

	RFC7231 (https://datatracker.ietf.org/doc/html/rfc7231#section-4.3.7):
		A server MUST generate a Content-Length field with a value of "0" if no payload body is to be sent in
		the response.

	RFC7230 (https://datatracker.ietf.org/doc/html/rfc7230#section-3.3.2):
		A server MUST NOT send a Content-Length header field in any response with a status code of 1xx (Informational)
		or 204 (No Content).
	*/
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetBytesKV(headerContentLength, headerValueZero)

	if len(p.methods) == 0 || len(ctx.Request.Header.PeekBytes(headerOrigin)) == 0 || (len(ctx.Request.Header.PeekBytes(headerAccessControlRequestMethod)) == 0 && ctx.Request.Header.IsOptions()) {
		return
	}

	ctx.Response.Header.SetBytesKV(headerAccessControlAllowMethods, p.methods)
}

func (p *CORSPolicy) handleVary(ctx *fasthttp.RequestCtx) {
	if len(p.vary) != 0 {
		ctx.Response.Header.SetBytesKV(headerVary, p.vary)
	}
}

//nolint:gocyclo
func (p *CORSPolicy) handleCORS(ctx *fasthttp.RequestCtx) {
	var (
		originURL *url.URL
		err       error
	)

	origin := ctx.Request.Header.PeekBytes(headerOrigin)
	acrm := ctx.Request.Header.PeekBytes(headerAccessControlRequestMethod)

	// This request is NOT a CORS Preflight request.
	if len(origin) == 0 || (len(acrm) == 0 && ctx.Request.Header.IsOptions()) {
		return
	}

	if len(origin) != 0 {
		// Skip processing of any `https` scheme URL that has not expressly been configured.
		if originURL, err = url.ParseRequestURI(string(origin)); err != nil || (originURL.Scheme != strProtoHTTPS && p.origins == nil) {
			return
		}
	}

	var allowedOrigin []byte

	switch len(p.origins) {
	case 0:
		allowedOrigin = origin
	case 1:
		if bytes.Equal(p.origins[0], headerValueOriginWildcard) {
			allowedOrigin = headerValueOriginWildcard
		} else if bytes.Equal(p.origins[0], origin) {
			allowedOrigin = origin
		}
	default:
		for i := 0; i < len(p.origins); i++ {
			if bytes.Equal(p.origins[i], headerValueOriginWildcard) {
				allowedOrigin = headerValueOriginWildcard

				break
			} else if bytes.Equal(p.origins[i], origin) {
				allowedOrigin = origin

				break
			}
		}

		if len(allowedOrigin) == 0 {
			return
		}
	}

	if len(allowedOrigin) != 0 {
		ctx.Response.Header.SetBytesKV(headerAccessControlAllowOrigin, allowedOrigin)
	}

	if len(p.credentials) != 0 {
		ctx.Response.Header.SetBytesKV(headerAccessControlAllowCredentials, p.credentials)
	}

	if len(p.maxAge) != 0 {
		ctx.Response.Header.SetBytesKV(headerAccessControlMaxAge, p.maxAge)
	}

	p.handleAllowedHeaders(ctx)
	p.handleAllowedMethods(ctx)
}

func (p *CORSPolicy) handleAllowedMethods(ctx *fasthttp.RequestCtx) {
	switch len(p.methods) {
	case 0:
		// TODO: It may be beneficial to be able to control this automatic behaviour.
		if requestMethods := ctx.Request.Header.PeekBytes(headerAccessControlRequestMethod); requestMethods != nil {
			ctx.Response.Header.SetBytesKV(headerAccessControlAllowMethods, requestMethods)
		}
	default:
		ctx.Response.Header.SetBytesKV(headerAccessControlAllowMethods, p.methods)
	}
}

func (p *CORSPolicy) handleAllowedHeaders(ctx *fasthttp.RequestCtx) {
	switch len(p.headers) {
	case 0:
		// TODO: It may be beneficial to be able to control this automatic behaviour.
		if headers := ctx.Request.Header.PeekBytes(headerAccessControlRequestHeaders); headers != nil {
			requestedHeaders := strings.Split(string(headers), ",")
			allowHeaders := make([]string, 0, len(requestedHeaders))

			for i := 0; i < len(requestedHeaders); i++ {
				headerTrimmed := strings.Trim(requestedHeaders[i], " ")

				if headerTrimmed == "*" {
					continue
				}

				if bytes.Equal(p.credentials, headerValueTrue) ||
					(!strings.EqualFold(fasthttp.HeaderCookie, headerTrimmed) &&
						!strings.EqualFold(fasthttp.HeaderAuthorization, headerTrimmed) &&
						!strings.EqualFold(fasthttp.HeaderProxyAuthorization, headerTrimmed)) {
					allowHeaders = append(allowHeaders, headerTrimmed)
				}
			}

			if len(allowHeaders) != 0 {
				ctx.Response.Header.SetBytesKV(headerAccessControlAllowHeaders, []byte(strings.Join(allowHeaders, ", ")))
			}
		}
	default:
		ctx.Response.Header.SetBytesKV(headerAccessControlAllowHeaders, p.headers)
	}
}
