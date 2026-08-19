package suites

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doDismissTooltips(t *testing.T, page *rod.Page) {
	require.NoError(t, page.Mouse.MoveTo(proto.Point{X: 0, Y: 0}))

	require.NoError(t, page.Wait(rod.Eval(`() => document.querySelector('[data-slot="tooltip-content"]') === null`)))
}
