package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewCORSMiddleware(t *testing.T) {
	cors := NewCORSMiddleware()

	assert.Equal(t, headerValueMaxAge, cors.maxAge)
	assert.Equal(t, headerValueVary, cors.vary)
	assert.Equal(t, headerValueFalse, cors.credentials)

	assert.Nil(t, cors.methods)
	assert.Nil(t, cors.origins)
	assert.Nil(t, cors.headers)
	assert.False(t, cors.varyOnly)
}

func TestCORSMiddleware_WithVary(t *testing.T) {
	cors := NewCORSMiddleware()

	assert.Equal(t, headerValueVary, cors.vary)
	assert.False(t, cors.varyOnly)

	cors.WithVary()
	assert.Nil(t, cors.vary)
	assert.False(t, cors.varyOnly)

	cors.WithVary("Origin", "Example", "Test")

	assert.Equal(t, []byte("Origin, Example, Test"), cors.vary)
	assert.False(t, cors.varyOnly)
}

func TestCORSMiddleware_WithAllowedMethods(t *testing.T) {
	cors := NewCORSMiddleware()

	assert.Nil(t, cors.methods)

	cors.WithAllowedMethods("GET")

	assert.Equal(t, []byte("GET"), cors.methods)

	cors.WithAllowedMethods("POST", "PATCH")

	assert.Equal(t, []byte("POST, PATCH"), cors.methods)

	cors.WithAllowedMethods()

	assert.Nil(t, cors.methods)
}

func TestCORSMiddleware_WithAllowedOrigins(t *testing.T) {
	cors := NewCORSMiddleware()

	assert.Nil(t, cors.origins)

	cors.WithAllowedOrigins("https://google.com", "http://localhost")

	assert.Equal(t, [][]byte{[]byte("https://google.com"), []byte("http://localhost")}, cors.origins)

	cors.WithAllowedOrigins()

	assert.Nil(t, cors.origins)
}

func TestCORSMiddleware_WithAllowedHeaders(t *testing.T) {
	cors := NewCORSMiddleware()

	assert.Nil(t, cors.headers)

	cors.WithAllowedHeaders("Example", "Another")

	assert.Equal(t, []string{"Example", "Another"}, cors.headers)

	cors.WithAllowedHeaders()

	assert.Nil(t, cors.headers)
}

func TestCORSMiddleware_WithAllowCredentials(t *testing.T) {
	cors := NewCORSMiddleware()

	assert.Equal(t, headerValueFalse, cors.credentials)

	cors.WithAllowCredentials(false)

	assert.Equal(t, headerValueFalse, cors.credentials)

	cors.WithAllowCredentials(true)

	assert.Equal(t, []byte("true"), cors.credentials)
}

func TestCORSMiddleware_WithVaryOnly(t *testing.T) {
	cors := NewCORSMiddleware()

	assert.False(t, cors.varyOnly)

	cors.WithVaryOnly(false)

	assert.False(t, cors.varyOnly)

	cors.WithVaryOnly(true)

	cors.WithVaryOnly(true)
}

func TestCORSMiddleware_WithMaxAge(t *testing.T) {
	cors := NewCORSMiddleware()

	assert.Equal(t, []byte("100"), cors.maxAge)

	cors.WithMaxAge(20)

	assert.Equal(t, []byte("20"), cors.maxAge)

	cors.WithMaxAge(0)

	assert.Nil(t, cors.maxAge)
}

func TestCORSMiddleware_HandleOPTIONS(t *testing.T) {
	fctx := &fasthttp.RequestCtx{}

	ctx, _ := NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSMiddleware()
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	fctx = &fasthttp.RequestCtx{}

	ctx, _ = NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedMethods("GET", "OPTIONS")
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSMiddleware_HandleOPTIONS_WithoutOrigin(t *testing.T) {
	fctx := &fasthttp.RequestCtx{}

	ctx, _ := NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")

	cors := NewCORSMiddleware()
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	fctx = &fasthttp.RequestCtx{}

	ctx, _ = NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")

	cors.WithAllowedMethods("GET", "OPTIONS")
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSMiddleware_HandleOPTIONSWithAllowedOrigins(t *testing.T) {
	fctx := &fasthttp.RequestCtx{}

	ctx, _ := NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSMiddleware()
	cors.WithAllowedOrigins("https://myapp.example.com")
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	fctx = &fasthttp.RequestCtx{}

	ctx, _ = NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedOrigins("https://anotherapp.example.com")
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	fctx = &fasthttp.RequestCtx{}

	ctx, _ = NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedOrigins("*")
	cors.WithAllowedMethods("GET", "OPTIONS")
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, originValueWildcard, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSMiddleware_HandleOPTIONSWithVaryOnly(t *testing.T) {
	fctx := &fasthttp.RequestCtx{}

	ctx, _ := NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSMiddleware()

	cors.WithVaryOnly(true)

	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	fctx = &fasthttp.RequestCtx{}

	ctx, _ = NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedMethods("GET", "OPTIONS")
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSMiddleware_HandleOPTIONSWithAllowedHeaders(t *testing.T) {
	fctx := &fasthttp.RequestCtx{}

	ctx, _ := NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSMiddleware()

	cors.WithAllowedHeaders("Example", "Test")

	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("Example, Test"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	fctx = &fasthttp.RequestCtx{}

	ctx, _ = NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedMethods("GET", "OPTIONS")
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("Example, Test"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	fctx = &fasthttp.RequestCtx{}

	ctx, _ = NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowCredentials(true)
	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueTrue, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("Example, Test, Cookie, Authorization, Proxy-Authorization"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSMiddleware_HandleOPTIONS_ShouldNotAllowWildcardInRequestedHeaders(t *testing.T) {
	fctx := &fasthttp.RequestCtx{}

	ctx, _ := NewAutheliaCtx(fctx, schema.Configuration{}, Providers{})

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "*")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSMiddleware()

	cors.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAllow))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func Test_CORSApplyAutomaticAllowAllPolicy_WithoutRequestMethod(t *testing.T) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.Response{}

	origin := []byte("https://myapp.example.com")
	req.Header.SetBytesKV(headerOrigin, origin)
	req.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")

	cors := NewCORSMiddleware()
	cors.handle(req, &resp)

	assert.Equal(t, []byte("Accept-Encoding, Origin"), resp.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, resp.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, resp.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, resp.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), resp.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlAllowMethods))
}

func Test_CORSApplyAutomaticAllowAllPolicy_WithRequestMethod(t *testing.T) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.Response{}

	origin := []byte("https://myapp.example.com")

	req.Header.SetBytesKV(headerOrigin, origin)
	req.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	req.Header.SetBytesK(headerAccessControlRequestMethod, "GET")

	cors := NewCORSMiddleware()
	cors.handle(req, &resp)

	assert.Equal(t, []byte("Accept-Encoding, Origin"), resp.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, resp.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, resp.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, resp.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), resp.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET"), resp.Header.PeekBytes(headerAccessControlAllowMethods))
}

func Test_CORSApplyAutomaticAllowAllPolicy_ShouldNotModifyFotNonHTTPSRequests(t *testing.T) {
	req := fasthttp.AcquireRequest()

	resp := fasthttp.Response{}

	origin := []byte("http://myapp.example.com")

	req.Header.SetBytesKV(headerOrigin, origin)
	req.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	req.Header.SetBytesK(headerAccessControlRequestMethod, "GET")

	cors := NewCORSMiddleware().WithVary()
	cors.handle(req, &resp)

	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlAllowMethods))
}
