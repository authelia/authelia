//go:build !externalsuites
// +build !externalsuites

package suites

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A suite whose setup fails before the tab is opened still has its teardown run, because testify defers
// it, and the page every teardown reaches for was never assigned. Each of these collected against that
// page unguarded, so the panic replaced the failure the setup had already reported and took the run's
// remaining results with it.
func TestTeardownShouldNotPanicWhenTheSetupLeftNoPage(t *testing.T) {
	t.Run("ShouldNotCollectCoverage", func(t *testing.T) {
		session := &RodSession{}

		require.NotPanics(t, func() { session.collectCoverage(nil) })
	})

	t.Run("ShouldNotCollectPage", func(t *testing.T) {
		session := &RodSession{}

		var err error

		require.NotPanics(t, func() { err = session.collectPage(nil, "TestTeardown") })
		assert.ErrorIs(t, err, errPageNotCreated)
	})

	t.Run("ShouldNotClosePage", func(t *testing.T) {
		suite := &RodSuite{}

		require.Nil(t, suite.Page)
		require.NotPanics(t, suite.MustClose)
	})
}
