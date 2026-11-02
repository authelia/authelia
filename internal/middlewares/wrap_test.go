// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"bufio"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestWrap(t *testing.T) {
	mw := func(s string) Basic {
		return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
			return func(ctx *fasthttp.RequestCtx) {
				ctx.Response.AppendBodyString(s)
				next(ctx)
			}
		}
	}
	next := func(ctx *fasthttp.RequestCtx) { ctx.Response.AppendBodyString("N") }

	testCases := []struct {
		name       string
		middleware Basic
		expected   string
	}{
		{name: "ShouldReturnNextWhenMiddlewareNil", middleware: nil, expected: "N"},
		{name: "ShouldApplyMiddlewareWhenNotNil", middleware: mw("M"), expected: "MN"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := Wrap(tc.middleware, next)

			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			handler(&ctx)

			require.Equal(t, tc.expected, string(ctx.Response.Body()))
		})
	}
}

func TestMultiWrap(t *testing.T) {
	mw := func(s string) Basic {
		return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
			return func(ctx *fasthttp.RequestCtx) {
				ctx.Response.AppendBodyString(s)
				next(ctx)
			}
		}
	}
	next := func(ctx *fasthttp.RequestCtx) { ctx.Response.AppendBodyString("N") }

	testCases := []struct {
		name        string
		middlewares []Basic
		expected    string
	}{
		{name: "ShouldReturnNextWhenNoMiddlewares", middlewares: nil, expected: "N"},
		{name: "ShouldApplyMiddlewaresInOrder", middlewares: []Basic{mw("A"), mw("B"), mw("C")}, expected: "ABCN"},
		{name: "ShouldSkipNilMiddlewares", middlewares: []Basic{mw("A"), nil, mw("C")}, expected: "ACN"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := MultiWrap(next, tc.middlewares...)

			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			handler(&ctx)

			require.Equal(t, tc.expected, string(ctx.Response.Body()))
		})
	}
}

func TestRequestCtxRemoteIP(t *testing.T) {
	testCases := []struct {
		name     string
		remote   string
		xff      string
		expected string
	}{
		{name: "ShouldUseXForwardedForWhenValid", remote: "10.0.0.1", xff: "203.0.113.9", expected: "203.0.113.9"},
		{name: "ShouldParseFirstIPWhenMultiple", remote: "10.0.0.1", xff: "203.0.113.9, 198.51.100.2", expected: "203.0.113.9"},
		{name: "ShouldFallbackToRemoteIPWhenHeaderMissing", remote: "192.0.2.10", xff: "", expected: "192.0.2.10"},
		{name: "ShouldFallbackToRemoteIPWhenHeaderInvalid", remote: "198.51.100.77", xff: "invalid", expected: "198.51.100.77"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			if tc.xff != "" {
				req.Header.SetBytesKV(headerXForwardedFor, []byte(tc.xff))
			}

			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP(tc.remote), Port: 12345}, nil)
			ip := RequestCtxRemoteIP(&ctx)
			require.Equal(t, tc.expected, ip.String())
		})
	}
}

func TestRequestCtxRemoteIPHeaderLines(t *testing.T) {
	const remote = "192.0.2.1"

	testCases := []struct {
		name     string
		values   []string
		expected string
	}{
		{"ShouldReturnRemoteWithoutHeader", nil, remote},
		{"ShouldReturnLeftMost", []string{"1.1.1.1, 203.0.113.9"}, "1.1.1.1"},
		{"ShouldReturnFirstLineLeftMost", []string{"1.1.1.1", "203.0.113.9"}, "1.1.1.1"},
		{"ShouldSkipEmptyFirstLine", []string{"", "1.1.1.1, 203.0.113.9"}, "1.1.1.1"},
		{"ShouldReturnRemoteWhenLeftMostEntryEmpty", []string{" , ,1.1.1.1"}, remote},
		{"ShouldReturnRemoteWhenAllEmpty", []string{"", " , "}, remote},
		{"ShouldReturnRemoteWhenLeftMostInvalid", []string{"invalid, 1.1.1.1"}, remote},
		{"ShouldReturnRemoteWhenLeftMostInvalidAfterEmptyLine", []string{"", "invalid", "1.1.1.1"}, remote},
		{"ShouldHandleIPv6", []string{"2001:db8::1, 1.1.1.1"}, "2001:db8::1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			raw := "GET / HTTP/1.1\r\nHost: example.com\r\n"

			for _, value := range tc.values {
				raw += "X-Forwarded-For: " + value + "\r\n"
			}

			ctx := &fasthttp.RequestCtx{}

			ctx.Init(&fasthttp.Request{}, &net.TCPAddr{IP: net.ParseIP(remote), Port: 1234}, nil)

			require.NoError(t, ctx.Request.Read(bufio.NewReader(strings.NewReader(raw+"\r\n"))))

			assert.Equal(t, tc.expected, RequestCtxRemoteIP(ctx).String())
		})
	}
}
