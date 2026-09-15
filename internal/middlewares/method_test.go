// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestNewMethodMaxLength(t *testing.T) {
	testCases := []struct {
		name         string
		length       int
		method       string
		expected     int
		expectedBody string
		expectedNext bool
	}{
		{
			"ShouldAllowGET",
			8,
			fasthttp.MethodGet,
			fasthttp.StatusOK,
			"",
			true,
		},
		{
			"ShouldAllowExactLength",
			8,
			"PROPFIND",
			fasthttp.StatusOK,
			"",
			true,
		},
		{
			"ShouldRejectOverLength",
			8,
			"PROPPATCHX",
			fasthttp.StatusMethodNotAllowed,
			"405 Method Not Allowed",
			false,
		},
		{
			"ShouldRejectOneOverLength",
			8,
			"PROPPATCH",
			fasthttp.StatusMethodNotAllowed,
			"405 Method Not Allowed",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}

			ctx.Request.Header.SetMethod(tc.method)

			called := false

			handler := NewMethodMaxLength(tc.length)(func(ctx *fasthttp.RequestCtx) {
				called = true
			})

			handler(ctx)

			assert.Equal(t, tc.expectedNext, called)
			assert.Equal(t, tc.expected, ctx.Response.StatusCode())
			assert.Equal(t, tc.expectedBody, string(ctx.Response.Body()))
		})
	}
}
