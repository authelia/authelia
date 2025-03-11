package embed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldPanicNilCtx(t *testing.T) {
	assert.Panics(t, func() {
		_ = ProvidersStartupCheck(nil, false)
	})

	ctx := &ctxEmbed{}

	assert.Panics(t, func() {
		_ = ProvidersStartupCheck(ctx, false)
	})
}
