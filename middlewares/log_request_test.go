package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestShouldCallNextFunction(t *testing.T) {
	var val = false
	f := func(ctx *fasthttp.RequestCtx) { val = true }

	context := &fasthttp.RequestCtx{}
	LogRequestMiddleware(f)(context)

	assert.Equal(t, true, val)
}
