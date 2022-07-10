package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestAllowedHosts(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		allowed  []string
		expected int
	}{
		{"ShouldRespondOKWhenNotConfigured", "authelia", nil, fasthttp.StatusOK},
		{"ShouldRespondOKWhenHostMatches", "authelia", []string{"authelia"}, fasthttp.StatusOK},
		{"ShouldRespondNotFoundWhenHostNotMatch", "auth.phishingdomain.com", []string{"authelia"}, fasthttp.StatusNotFound},
		{"ShouldRespondOKWhenHostMatchesMulti", "authelia", []string{"127.0.0.1", "authelia"}, fasthttp.StatusOK},
	}

	handler := func(ctx *fasthttp.RequestCtx) {}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			middleware := AllowedHosts(tc.allowed)

			ctx := &fasthttp.RequestCtx{}

			ctx.Request.SetHost(tc.have)

			middleware(handler)(ctx)

			assert.Equal(t, tc.expected, ctx.Response.StatusCode())
		})
	}
}
