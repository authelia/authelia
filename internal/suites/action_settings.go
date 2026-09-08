package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doOpenSettings(t *testing.T, page *rod.Page) {
	rs.ClickElementLocatedByID(t, page, "account-menu")
	rs.ClickElementLocatedByID(t, page, "account-menu-settings")
}

func (rs *RodSession) doOpenSettingsMenu(t *testing.T, page *rod.Page) {
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
