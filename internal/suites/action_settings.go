package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doSettingsOpen(t *testing.T, page *rod.Page) {
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "account-menu").Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "account-menu-settings").Click("left", 1))
	require.NoError(t, page.WaitStable(time.Millisecond*100))
}

func (rs *RodSession) doSettingsMenuOpen(t *testing.T, page *rod.Page) {
	require.NoError(t, page.WaitStable(time.Millisecond*500))

	rs.doHoverAllMuiTooltip(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "settings-menu").Click("left", 1))

	require.NoError(t, page.WaitStable(time.Millisecond*10))
}

func (rs *RodSession) doSettingsMenuTwoFactor(t *testing.T, page *rod.Page) {
	rs.doSettingsMenuOpen(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "settings-menu-twofactor").Click("left", 1))
}

func (rs *RodSession) doSettingsMenuClose(t *testing.T, page *rod.Page) {
	rs.doSettingsMenuOpen(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "settings-menu-close").Click("left", 1))
}
