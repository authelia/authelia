package suites

import (
	"fmt"
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doCreateTab(url string) *rod.Page {
	p := rs.WebDriver.MustIncognito().MustPage(url)
	return p
}

func (rs *RodSession) doVisit(page *rod.Page, url string) {
	page.MustNavigate(url)
}

func (rs *RodSession) doVisitAndVerifyOneFactorStep(t *testing.T, page *rod.Page, url string) {
	rs.doVisit(page, url)
	rs.verifyIsFirstFactorPage(t, page)
}

func (rs *RodSession) doVisitLoginPage(t *testing.T, page *rod.Page, targetURL string) {
	suffix := ""
	if targetURL != "" {
		suffix = fmt.Sprintf("?rd=%s", targetURL)
	}

	rs.doVisitAndVerifyOneFactorStep(t, page, fmt.Sprintf("%s/%s", GetLoginBaseURL(), suffix))
}
