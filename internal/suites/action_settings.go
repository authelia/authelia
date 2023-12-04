package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doOpenSettings(t *testing.T, page *rod.Page) {
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "account-menu").Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "account-menu-settings").Click("left", 1))
	require.NoError(t, page.WaitStable(time.Millisecond*100))
}

func (rs *RodSession) doOpenSettingsMenu(t *testing.T, page *rod.Page) {
	require.NoError(t, page.WaitStable(time.Millisecond*100))

	rs.doHoverAllMuiTooltip(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "settings-menu").Click("left", 1))

	require.NoError(t, page.WaitStable(time.Millisecond*10))
}

func (rs *RodSession) doOpenSettingsMenuClickTwoFactor(t *testing.T, page *rod.Page) {
	rs.doOpenSettingsMenu(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "settings-menu-twofactor").Click("left", 1))
}

func (rs *RodSession) doOpenSettingsMenuClickClose(t *testing.T, page *rod.Page) {
	rs.doOpenSettingsMenu(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "settings-menu-close").Click("left", 1))
}
