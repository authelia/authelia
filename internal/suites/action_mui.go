package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doHoverAllMuiTooltip(t *testing.T, page *rod.Page) {
	pos := page.Mouse.Position()

	elements, err := page.Elements(".MuiTooltip-tooltip")

	require.NoError(t, err)

	for _, element := range elements {
		require.NoError(t, element.Hover())
		require.NoError(t, page.WaitStable(time.Millisecond*10))
	}

	require.NoError(t, page.Mouse.MoveTo(pos))
}
