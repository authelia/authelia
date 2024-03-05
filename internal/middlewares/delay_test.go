package middlewares

import (
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestArbitraryDelay(t *testing.T) {
	handler := func(ctx *fasthttp.RequestCtx) {
		time.Sleep(time.Millisecond * 30)
	}

	ctx := &fasthttp.RequestCtx{}

	ArbitraryDelay(time.Millisecond * 20)(handler)(ctx)
	ArbitraryDelay(time.Millisecond * 40)(handler)(ctx)
}
