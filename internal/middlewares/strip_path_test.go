package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestStripPath(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		uri         string
		expected    string
		expectedRaw string
		expectedURI string
	}{
		{
			"ShouldHandleEmpty",
			"",
			"",
			"",
			"",
			"",
		},
		{
			"ShouldHandleSingleSlash",
			"/",
			"",
			"",
			"",
			"",
		},
		{
			"ShouldHandleAuthPath",
			"/auth",
			"/auth",
			"/auth",
			"/auth",
			"/",
		},
		{
			"ShouldHandleAuthSubPath",
			"/auth",
			"/auth/example",
			"/auth",
			"/auth/example",
			"/example",
		},
		{
			"ShouldHandleAuthPathNoLeadingSlash",
			"auth",
			"/auth",
			"/auth",
			"/auth",
			"/",
		},
		{
			"ShouldHandleAuthSubPathNoLeadingSlash",
			"auth",
			"/auth/example",
			"/auth",
			"/auth/example",
			"/example",
		},
	}

	nilHandler := func(ctx *fasthttp.RequestCtx) {}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			header := fasthttp.RequestHeader{}

			header.SetRequestURI(tc.uri)

			ctx := &fasthttp.RequestCtx{
				Request: fasthttp.Request{
					Header:                         header, //nolint:govet
					UseHostHeader:                  false,
					DisableRedirectPathNormalizing: false,
				},
			}

			var next fasthttp.RequestHandler

			if tc.path != "" && tc.path != "/" {
				next = nilHandler
			}

			handler := StripPath(tc.path)(next)

			if tc.path == "" || tc.path == "/" {
				assert.Nil(t, handler)

				return
			}

			handler(ctx)

			baseURL := ctx.UserValue(UserValueKeyBaseURL)
			rawURI := ctx.UserValue(UserValueKeyRawURI)

			assert.Equal(t, tc.expected, baseURL)
			assert.Equal(t, tc.expectedRaw, rawURI)
			assert.Equal(t, tc.expectedURI, string(ctx.RequestURI()))
		})
	}
}
