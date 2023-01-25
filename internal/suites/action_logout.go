package suites

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doLogout(t *testing.T, page *rod.Page) {
	rs.doVisit(t, page, fmt.Sprintf("%s%s", GetLoginBaseURL(BaseDomain), "/logout"))
	rs.verifyIsFirstFactorPage(t, page)
}

func (rs *RodSession) doLogoutWithRedirect(t *testing.T, page *rod.Page, targetURL string, firstFactor bool) {
	rs.doVisit(t, page, fmt.Sprintf("%s%s%s", GetLoginBaseURL(BaseDomain), "/logout?rd=", url.QueryEscape(targetURL)))

	if firstFactor {
		rs.verifyIsFirstFactorPage(t, page)

		return
	}

	page.MustElementR("h1", "Public resource")

	rs.verifyURLIs(t, page, targetURL)
}
