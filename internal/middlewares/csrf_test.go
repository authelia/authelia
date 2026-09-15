// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestCrossSiteRequestForgery(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		headers  map[string]string
		expected int
	}{
		{
			"ShouldAllowSafeMethodCrossSite",
			fasthttp.MethodGet,
			map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "https://evil.example.org"},
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowHeadMethodCrossSite",
			fasthttp.MethodHead,
			map[string]string{"Sec-Fetch-Site": "cross-site"},
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowOptionsMethodCrossSite",
			fasthttp.MethodOptions,
			map[string]string{"Sec-Fetch-Site": "cross-site"},
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowUnsafeMethodSameOrigin",
			fasthttp.MethodPost,
			map[string]string{"Sec-Fetch-Site": "same-origin"},
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowUnsafeMethodUserInitiated",
			fasthttp.MethodPost,
			map[string]string{"Sec-Fetch-Site": "none"},
			fasthttp.StatusOK,
		},
		{
			"ShouldRejectUnsafeMethodSameSite",
			fasthttp.MethodPost,
			map[string]string{"Sec-Fetch-Site": "same-site"},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectUnsafeMethodCrossSite",
			fasthttp.MethodPost,
			map[string]string{"Sec-Fetch-Site": "cross-site"},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectPutMethodCrossSite",
			fasthttp.MethodPut,
			map[string]string{"Sec-Fetch-Site": "cross-site"},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectDeleteMethodCrossSite",
			fasthttp.MethodDelete,
			map[string]string{"Sec-Fetch-Site": "cross-site"},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectPatchMethodSameSite",
			fasthttp.MethodPatch,
			map[string]string{"Sec-Fetch-Site": "same-site"},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldPreferSecFetchSiteOverMatchingOrigin",
			fasthttp.MethodPost,
			map[string]string{"Sec-Fetch-Site": "same-site", "Origin": "http://auth.example.com"},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldAllowUnsafeMethodWithoutBrowserHeaders",
			fasthttp.MethodPost,
			nil,
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowUnsafeMethodMatchingOrigin",
			fasthttp.MethodPost,
			map[string]string{"Origin": "http://auth.example.com"},
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowUnsafeMethodMatchingOriginCaseInsensitive",
			fasthttp.MethodPost,
			map[string]string{"Origin": "HTTP://Auth.Example.com"},
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowUnsafeMethodMatchingForwardedOrigin",
			fasthttp.MethodPost,
			map[string]string{"Origin": "https://login.example.com", "X-Forwarded-Proto": "https", "X-Forwarded-Host": "login.example.com"},
			fasthttp.StatusOK,
		},
		{
			"ShouldRejectUnsafeMethodOriginSiblingSubdomain",
			fasthttp.MethodPost,
			map[string]string{"Origin": "http://app.example.com"},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectUnsafeMethodOriginSchemeMismatch",
			fasthttp.MethodPost,
			map[string]string{"Origin": "http://login.example.com", "X-Forwarded-Proto": "https", "X-Forwarded-Host": "login.example.com"},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectUnsafeMethodOriginPortMismatch",
			fasthttp.MethodPost,
			map[string]string{"Origin": "http://auth.example.com:8080"},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectUnsafeMethodNullOrigin",
			fasthttp.MethodPost,
			map[string]string{"Origin": "null"},
			fasthttp.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}

			ctx.Request.Header.SetMethod(tc.method)
			ctx.Request.Header.SetHost("auth.example.com")
			ctx.Request.SetRequestURI("/api/firstfactor")

			for key, value := range tc.headers {
				ctx.Request.Header.Set(key, value)
			}

			called := false

			CrossSiteRequestForgery(func(ctx *fasthttp.RequestCtx) {
				called = true

				ctx.SetStatusCode(fasthttp.StatusOK)
			})(ctx)

			assert.Equal(t, tc.expected, ctx.Response.StatusCode())
			assert.Equal(t, tc.expected == fasthttp.StatusOK, called)

			if tc.expected == fasthttp.StatusForbidden {
				assert.Equal(t, "403 Forbidden", string(ctx.Response.Body()))
				assert.Equal(t, "text/plain; charset=utf-8", string(ctx.Response.Header.ContentType()))
			}
		})
	}
}
