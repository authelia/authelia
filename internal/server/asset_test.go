package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestSetCacheControl(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected []byte
	}{
		{"ShouldNotSetCacheControl", "index.html", nil},
		{"ShouldSetCacheControlForReactJS", "index.abc123.js", headerValueCacheControlReact},
		{"ShouldSetCacheControlForReactCSS", "index.abc123.css", headerValueCacheControlReact},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}

			ctx.Request.SetRequestURI(tc.have)

			setCacheControl(ctx)

			assert.Equal(t, tc.expected, ctx.Response.Header.PeekBytes(headerCacheControl))
		})
	}
}
