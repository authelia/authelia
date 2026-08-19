package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doOpenSettings(t *testing.T, page *rod.Page) {
	rs.ClickElementLocatedByID(t, page, "account-menu")
	rs.ClickElementLocatedByID(t, page, "account-menu-settings")
	require.NoError(t, page.WaitStable(time.Millisecond*100))
}

func (rs *RodSession) doOpenSettingsMenu(t *testing.T, page *rod.Page) {
	rs.doDismissTooltips(t, page)

	rs.ClickElementLocatedByID(t, page, "settings-menu")
}

func (rs *RodSession) doOpenSettingsMenuClickSecurity(t *testing.T, page *rod.Page) {
	rs.doOpenSettingsMenu(t, page)

	rs.ClickElementLocatedByID(t, page, "settings-menu-security")
}

func (rs *RodSession) doOpenSettingsMenuClickTwoFactor(t *testing.T, page *rod.Page) {
	rs.doOpenSettingsMenu(t, page)

	rs.ClickElementLocatedByID(t, page, "settings-menu-twofactor")
}

func (rs *RodSession) doOpenSettingsMenuClickClose(t *testing.T, page *rod.Page) {
	rs.doOpenSettingsMenu(t, page)

	rs.ClickElementLocatedByID(t, page, "settings-menu-close")
}
