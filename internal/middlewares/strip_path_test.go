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
			"ShouldHandleAuthPathNoPath",
			"/auth",
			"/",
			"",
			"",
			"/",
		},
		{
			"ShouldHandleAuthPathNoPathQuery",
			"/auth",
			"/?rd=abc",
			"",
			"",
			"/?rd=abc",
		},
		{
			"ShouldHandleAuthPathNoPathQueryNoSlash",
			"/auth",
			"?rd=abc",
			"",
			"",
			"?rd=abc",
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
			"ShouldHandleAuthSubPathQuery",
			"/auth",
			"/auth?rd=123",
			"/auth",
			"/auth?rd=123",
			"?rd=123",
		},
		{
			"ShouldHandleAuthSubPathQueryWithTrailingSlash",
			"/auth",
			"/auth/?rd=123",
			"/auth",
			"/auth/?rd=123",
			"/?rd=123",
		},
		{
			"ShouldHandleShortSubPath",
			"/a",
			"/api/example",
			"",
			"",
			"/api/example",
		},
		{
			"ShouldHandleShortSubPathQuery",
			"/a",
			"/api/example?rd=123",
			"",
			"",
			"/api/example?rd=123",
		},
		{
			"ShouldHandleShortSubPathQueryWithTrailingSlash",
			"/a",
			"/api/example/?rd=123",
			"",
			"",
			"/api/example/?rd=123",
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

			if tc.expected == "" {
				assert.Nil(t, baseURL)
			} else {
				assert.Equal(t, tc.expected, baseURL)
			}

			if tc.expectedRaw == "" {
				assert.Nil(t, rawURI)
			} else {
				assert.Equal(t, tc.expectedRaw, rawURI)
			}

			assert.Equal(t, tc.expectedURI, string(ctx.RequestURI()))
		})
	}
}
