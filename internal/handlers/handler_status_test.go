package handlers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestStatus(t *testing.T) {
	testCases := []struct {
		name string
		code int
	}{
		{"ShouldHandleOK", fasthttp.StatusOK},
		{"ShouldHandleNotFound", fasthttp.StatusNotFound},
		{"ShouldHandleMethodNotAllowed", fasthttp.StatusMethodNotAllowed},
		{"ShouldHandleBadGateway", fasthttp.StatusBadGateway},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}

			Status(tc.code)(ctx)

			assert.Equal(t, tc.code, ctx.Response.StatusCode())
			assert.Equal(t, fmt.Sprintf("%d %s", tc.code, fasthttp.StatusMessage(tc.code)), string(ctx.Response.Body()))
			assert.Equal(t, "text/plain; charset=utf-8", string(ctx.Response.Header.ContentType()))
		})
	}
}

func TestSetStatusCodeResponse(t *testing.T) {
	testCases := []struct {
		name string
		code int
	}{
		{"ShouldHandleOK", fasthttp.StatusOK},
		{"ShouldHandleForbidden", fasthttp.StatusForbidden},
		{"ShouldHandleUnauthorized", fasthttp.StatusUnauthorized},
		{"ShouldHandleInternalServerError", fasthttp.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}

			ctx.Response.Header.Set(fasthttp.HeaderLocation, "https://example.com")
			ctx.SetBodyString("this should be discarded")

			SetStatusCodeResponse(ctx, tc.code)

			assert.Equal(t, tc.code, ctx.Response.StatusCode())
			assert.Equal(t, fmt.Sprintf("%d %s", tc.code, fasthttp.StatusMessage(tc.code)), string(ctx.Response.Body()))
			assert.Equal(t, "text/plain; charset=utf-8", string(ctx.Response.Header.ContentType()))
			assert.Equal(t, "", string(ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
		})
	}
}
