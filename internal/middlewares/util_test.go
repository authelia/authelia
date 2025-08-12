package middlewares

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestSetContentType(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func(*fasthttp.RequestCtx)
		expected string
	}{
		{
			name:     "ShouldSetContentTypeApplicationJSON",
			fn:       SetContentTypeApplicationJSON,
			expected: "application/json; charset=utf-8",
		},
		{
			name:     "ShouldSetContentTypeTextPlain",
			fn:       SetContentTypeTextPlain,
			expected: "text/plain; charset=utf-8",
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			tc := testCases[i]

			var ctx fasthttp.RequestCtx

			tc.fn(&ctx)

			require.Equal(t, tc.expected, string(ctx.Response.Header.ContentType()))
		})
	}
}

func TestNewAuthenticationProvider(t *testing.T) {
	testCases := []struct {
		name   string
		config schema.Configuration
	}{
		{
			name:   "ShouldReturnNilProviderWhenNoBackendConfigured",
			config: schema.Configuration{},
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			tc := testCases[i]
			provider := NewAuthenticationProvider(&tc.config, nil)
			require.Nil(t, provider)
		})
	}
}
