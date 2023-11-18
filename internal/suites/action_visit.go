package suites

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doCreateTab(t *testing.T, url string) *rod.Page {
	p, err := rs.WebDriver.MustIncognito().Page(proto.TargetCreateTarget{URL: url})
	assert.NoError(t, err)

	return p
}

func (rs *RodSession) doVisit(t *testing.T, page *rod.Page, url string) {
	assert.NoError(t, page.Navigate(url))
	require.NoError(t, page.WaitStable(time.Millisecond*50))
}

func (rs *RodSession) doVisitAndVerifyOneFactorStep(t *testing.T, page *rod.Page, url string) {
	rs.doVisit(t, page, url)
	rs.verifyIsFirstFactorPage(t, page)
}

func (rs *RodSession) doVisitLoginPage(t *testing.T, page *rod.Page, baseDomain string, targetURL string) {
	suffix := ""
	if targetURL != "" {
		suffix = fmt.Sprintf("?rd=%s", targetURL)
	}

	rs.doVisitAndVerifyOneFactorStep(t, page, fmt.Sprintf("%s/%s", GetLoginBaseURL(baseDomain), suffix))
}
