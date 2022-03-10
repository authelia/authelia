package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func Test_CORSApplyAutomaticAllowAllPolicy_WithoutRequestMethod(t *testing.T) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.Response{}

	origin := []byte("https://myapp.example.com")

	req.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")

	corsApplyAutomaticAllowAllPolicy(req, &resp, origin)

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

	req.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	req.Header.SetBytesK(headerAccessControlRequestMethod, "GET")

	corsApplyAutomaticAllowAllPolicy(req, &resp, origin)

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

	req.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	req.Header.SetBytesK(headerAccessControlRequestMethod, "GET")

	corsApplyAutomaticAllowAllPolicy(req, &resp, origin)

	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), resp.Header.PeekBytes(headerAccessControlAllowMethods))
}
