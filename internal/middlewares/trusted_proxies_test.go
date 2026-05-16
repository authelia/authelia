// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewTrustedProxies(t *testing.T) {
	_, network, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	const (
		trusted   = "10.0.0.1"
		untrusted = "192.0.2.1"
	)

	testCases := []struct {
		name     string
		remote   string
		raw      string
		header   string
		expected string
	}{
		{
			"ShouldStripHeaderFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nX-Original-URL: https://example.com/\r\n\r\n",
			"X-Original-URL",
			"",
		},
		{
			"ShouldStripXAutheliaURLFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nX-Authelia-URL: https://example.com/\r\n\r\n",
			"X-Authelia-URL",
			"",
		},
		{
			"ShouldStripTrailerFromUntrusted",
			untrusted,
			"POST / HTTP/1.1\r\nHost: example.com\r\nTransfer-Encoding: chunked\r\n\r\n1\r\na\r\n0\r\nX-Original-URL: https://example.com/\r\n\r\n",
			"X-Original-URL",
			"",
		},
		{
			"ShouldStripXRealFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nX-Real-IP: 192.0.2.2\r\n\r\n",
			"X-Real-IP",
			"",
		},
		{
			"ShouldStripXForwardedFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nX-Forwarded-For: 192.0.2.2\r\n\r\n",
			"X-Forwarded-For",
			"",
		},
		{
			"ShouldStripForwardedFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nForwarded: for=192.0.2.2\r\n\r\n",
			"Forwarded",
			"",
		},
		{
			"ShouldStripViaFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nVia: 1.1 example\r\n\r\n",
			"Via",
			"",
		},
		{
			"ShouldStripLowerCaseHeaderFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nx-forwarded-host: example.com\r\n\r\n",
			"X-Forwarded-Host",
			"",
		},
		{
			"ShouldStripLowerCaseExactHeaderFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nx-authelia-url: https://example.com/\r\n\r\n",
			"X-Authelia-URL",
			"",
		},
		{
			"ShouldNotStripProxyAuthorizationFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nProxy-Authorization: Basic abc\r\n\r\n",
			"Proxy-Authorization",
			"Basic abc",
		},
		{
			"ShouldStripHeaderFromUntrustedIPv6",
			"2001:db8::1",
			"GET / HTTP/1.1\r\nHost: example.com\r\nX-Original-URL: https://example.com/\r\n\r\n",
			"X-Original-URL",
			"",
		},
		{
			"ShouldNotStripHeaderFromTrustedIPv4Mapped",
			"::ffff:10.0.0.1",
			"GET / HTTP/1.1\r\nHost: example.com\r\nX-Original-URL: https://example.com/\r\n\r\n",
			"X-Original-URL",
			"https://example.com/",
		},
		{
			"ShouldNotStripOtherHeaderFromUntrusted",
			untrusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nX-Other: abc\r\n\r\n",
			"X-Other",
			"abc",
		},
		{
			"ShouldNotStripHeaderFromTrusted",
			trusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nX-Original-URL: https://example.com/\r\n\r\n",
			"X-Original-URL",
			"https://example.com/",
		},
		{
			"ShouldNotStripXAutheliaURLFromTrusted",
			trusted,
			"GET / HTTP/1.1\r\nHost: example.com\r\nX-Authelia-URL: https://example.com/\r\n\r\n",
			"X-Authelia-URL",
			"https://example.com/",
		},
		{
			"ShouldNotStripTrailerFromTrusted",
			trusted,
			"POST / HTTP/1.1\r\nHost: example.com\r\nTransfer-Encoding: chunked\r\n\r\n1\r\na\r\n0\r\nX-Original-URL: https://example.com/\r\n\r\n",
			"X-Original-URL",
			"https://example.com/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}

			ctx.Init(&fasthttp.Request{}, &net.TCPAddr{IP: net.ParseIP(tc.remote), Port: 1234}, nil)

			require.NoError(t, ctx.Request.Read(bufio.NewReader(strings.NewReader(tc.raw))))

			var called bool

			NewTrustedProxies(schema.ServerHeadersTrustedProxies{Enable: true, Networks: []*net.IPNet{network}})(func(ctx *fasthttp.RequestCtx) {
				called = true

				assert.Equal(t, tc.expected, string(ctx.Request.Header.Peek(tc.header)))
			})(ctx)

			assert.True(t, called)
		})
	}
}

func TestNewTrustedProxiesShouldNotStripWhenDisabled(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	ctx.Init(&fasthttp.Request{}, &net.TCPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1234}, nil)
	ctx.Request.Header.Set("X-Original-URL", "https://example.com/")

	NewTrustedProxies(schema.ServerHeadersTrustedProxies{})(func(ctx *fasthttp.RequestCtx) {
		assert.Equal(t, "https://example.com/", string(ctx.Request.Header.Peek("X-Original-URL")))
	})(ctx)
}

func TestNewTrustedProxiesShouldStripWhenEnabledWithoutNetworks(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	ctx.Init(&fasthttp.Request{}, &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1234}, nil)
	ctx.Request.Header.Set("X-Original-URL", "https://example.com/")

	var called bool

	NewTrustedProxies(schema.ServerHeadersTrustedProxies{Enable: true})(func(ctx *fasthttp.RequestCtx) {
		called = true

		assert.Empty(t, ctx.Request.Header.Peek("X-Original-URL"))
	})(ctx)

	assert.True(t, called)
}

func TestTrustedProxiesRemoteIPShouldUseHeaderWhenDisabled(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	ctx.Init(&fasthttp.Request{}, &net.TCPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1234}, nil)
	ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "1.1.1.1, 203.0.113.9")

	assert.Equal(t, "1.1.1.1", TrustedProxiesRemoteIP(ctx, schema.ServerHeadersTrustedProxies{}).String())
}

func TestTrustedProxiesShouldTrustNonTCPPeer(t *testing.T) {
	config := schema.ServerHeadersTrustedProxies{Enable: true}

	ctx := &fasthttp.RequestCtx{}

	ctx.Init(&fasthttp.Request{}, &net.UnixAddr{Name: "@", Net: "unix"}, nil)
	ctx.Request.Header.Set("X-Original-URL", "https://example.com/")
	ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "1.1.1.1, 203.0.113.9")

	assert.Equal(t, "1.1.1.1", TrustedProxiesRemoteIP(ctx, config).String())

	var called bool

	NewTrustedProxies(config)(func(ctx *fasthttp.RequestCtx) {
		called = true

		assert.Equal(t, "https://example.com/", string(ctx.Request.Header.Peek("X-Original-URL")))
		assert.Equal(t, "1.1.1.1", RequestCtxRemoteIP(ctx).String())
	})(ctx)

	assert.True(t, called)
}

func TestTrustedProxiesShouldNotTrustUnknownPeer(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	ctx.Request.Header.Set("X-Original-URL", "https://example.com/")

	var called bool

	NewTrustedProxies(schema.ServerHeadersTrustedProxies{Enable: true})(func(ctx *fasthttp.RequestCtx) {
		called = true

		assert.Empty(t, ctx.Request.Header.Peek("X-Original-URL"))
	})(ctx)

	assert.True(t, called)
}

func TestTrustedProxiesShouldTrustUnixSocketListener(t *testing.T) {
	dir, err := os.MkdirTemp("", "authelia")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	listener, err := net.Listen("unix", filepath.Join(dir, "s.sock"))
	require.NoError(t, err)

	type result struct {
		url, ip string
	}

	results := make(chan result, 1)

	config := schema.ServerHeadersTrustedProxies{Enable: true}

	server := &fasthttp.Server{Handler: NewTrustedProxies(config)(func(ctx *fasthttp.RequestCtx) {
		results <- result{string(ctx.Request.Header.Peek("X-Original-URL")), RequestCtxRemoteIP(ctx).String()}
	})}

	go func() {
		_ = server.Serve(listener)
	}()

	defer func() {
		_ = server.Shutdown()
	}()

	conn, err := net.Dial("unix", listener.Addr().String())
	require.NoError(t, err)

	defer conn.Close()

	_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: example.com\r\nX-Original-URL: https://example.com/\r\nX-Forwarded-For: 1.1.1.1, 203.0.113.9\r\nConnection: close\r\n\r\n"))
	require.NoError(t, err)

	select {
	case actual := <-results:
		assert.Equal(t, result{"https://example.com/", "1.1.1.1"}, actual)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the request")
	}
}

func TestNewTrustedProxiesAuthz(t *testing.T) {
	_, network, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	const (
		trusted   = "10.0.0.1"
		untrusted = "192.0.2.1"
		uri       = "/api/authz/x?authelia_url=https://auth.example.com/&rd=https://auth.example.com/&other=abc"
	)

	testCases := []struct {
		name              string
		implementation    string
		remote            string
		expectAutheliaURL bool
		expectRD          bool
	}{
		{"ShouldRemoveAutheliaURLForwardAuth", schema.AuthzImplementationForwardAuth, untrusted, false, true},
		{"ShouldRemoveAutheliaURLAuthRequest", schema.AuthzImplementationAuthRequest, untrusted, false, true},
		{"ShouldRemoveRDLegacy", schema.AuthzImplementationLegacy, untrusted, true, false},
		{"ShouldRemoveRDDefault", "", untrusted, true, false},
		{"ShouldNotRemoveForwardAuthTrusted", schema.AuthzImplementationForwardAuth, trusted, true, true},
		{"ShouldNotRemoveAuthRequestTrusted", schema.AuthzImplementationAuthRequest, trusted, true, true},
		{"ShouldNotRemoveLegacyTrusted", schema.AuthzImplementationLegacy, trusted, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}

			ctx.Init(&fasthttp.Request{}, &net.TCPAddr{IP: net.ParseIP(tc.remote), Port: 1234}, nil)
			ctx.Request.SetRequestURI(uri)

			middleware := NewTrustedProxiesAuthz(schema.ServerHeadersTrustedProxies{Enable: true, Networks: []*net.IPNet{network}}, tc.implementation)
			require.NotNil(t, middleware)

			var called bool

			middleware(func(ctx *fasthttp.RequestCtx) {
				called = true

				assert.Equal(t, tc.expectAutheliaURL, ctx.QueryArgs().Has(queryArgAutheliaURL))
				assert.Equal(t, tc.expectRD, ctx.QueryArgs().Has(queryArgRedirect))
				assert.Equal(t, "abc", string(ctx.QueryArgs().Peek("other")))
				assert.Equal(t, uri, string(ctx.RequestURI()))
			})(ctx)

			assert.True(t, called)
		})
	}

	assert.Nil(t, NewTrustedProxiesAuthz(schema.ServerHeadersTrustedProxies{Enable: true}, schema.AuthzImplementationExtAuthz))
	assert.Nil(t, NewTrustedProxiesAuthz(schema.ServerHeadersTrustedProxies{}, schema.AuthzImplementationForwardAuth))
}
