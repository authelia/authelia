package middlewares

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestRecoverErr(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		expected any
	}{
		{
			"ShouldHandleNil",
			nil,
			nil,
		},
		{
			"ShouldHandleString",
			"a string",
			fmt.Errorf("recovered panic: a string"),
		},
		{
			"ShouldHandleWrapped",
			fmt.Errorf("a string"),
			fmt.Errorf("recovered panic: %w", fmt.Errorf("a string")),
		},
		{
			"ShouldHandleInt",
			5,
			fmt.Errorf("recovered panic with unknown type: 5"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, recoverErr(tc.have))
		})
	}
}

func TestRecoverPanic(t *testing.T) {
	type testCase struct {
		name                        string
		next                        fasthttp.RequestHandler
		expectedStatus              int
		expectedBody                string
		expectedContentTypeContains string
	}

	testCases := []testCase{
		{
			name: "ShouldRecoverFromStringPanic",
			next: func(ctx *fasthttp.RequestCtx) {
				panic("boom")
			},
			expectedStatus:              fasthttp.StatusInternalServerError,
			expectedBody:                "500 Internal Server Error",
			expectedContentTypeContains: "text/plain",
		},
		{
			name: "ShouldRecoverFromErrorPanic",
			next: func(ctx *fasthttp.RequestCtx) {
				panic(errors.New("fail"))
			},
			expectedStatus:              fasthttp.StatusInternalServerError,
			expectedBody:                "500 Internal Server Error",
			expectedContentTypeContains: "text/plain",
		},
		{
			name: "ShouldRecoverFromUnknownTypePanic",
			next: func(ctx *fasthttp.RequestCtx) {
				panic(struct{}{})
			},
			expectedStatus:              fasthttp.StatusInternalServerError,
			expectedBody:                "500 Internal Server Error",
			expectedContentTypeContains: "text/plain",
		},
		{
			name: "ShouldPassThroughWithoutPanic",
			next: func(ctx *fasthttp.RequestCtx) {
				ctx.SetStatusCode(fasthttp.StatusCreated)
				ctx.SetContentType("application/json")
				ctx.SetBodyString("ok")
			},
			expectedStatus:              fasthttp.StatusCreated,
			expectedBody:                "ok",
			expectedContentTypeContains: "application/json",
		},
		{
			name: "ShouldResetResponseAfterPanic",
			next: func(ctx *fasthttp.RequestCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetBodyString("partial")
				panic("oops")
			},
			expectedStatus:              fasthttp.StatusInternalServerError,
			expectedBody:                "500 Internal Server Error",
			expectedContentTypeContains: "text/plain",
		},
	}

	newCtx := func() *fasthttp.RequestCtx {
		var (
			ctx fasthttp.RequestCtx
			req fasthttp.Request
		)

		req.Header.SetMethod("GET")
		req.SetRequestURI("http://example.com/")
		ctx.Init(&req, nil, nil)

		return &ctx
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newCtx()
			h := RecoverPanic(tc.next)
			h(ctx)

			require.Equal(t, tc.expectedStatus, ctx.Response.StatusCode())

			if tc.expectedBody != "" {
				require.Equal(t, tc.expectedBody, string(ctx.Response.Body()))
			}

			if tc.expectedContentTypeContains != "" {
				assert.Contains(t, string(ctx.Response.Header.ContentType()), tc.expectedContentTypeContains)
			}
		})
	}
}
