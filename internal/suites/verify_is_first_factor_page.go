package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) verifyIsFirstFactorPage(t *testing.T, page *rod.Page) {
	require.NoError(t, page.WaitStable(time.Millisecond*50))

	rs.WaitElementLocatedByID(t, page, "first-factor-stage")
}
