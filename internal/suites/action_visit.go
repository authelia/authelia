package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/require"
)

func (s *RodSuite) doSetupTest(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), setupTestTimeout)
	defer cancel()

	s.Page = s.doCreateTab(s.T(), url)
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (rs *RodSession) doCreateTab(t *testing.T, url string) *rod.Page {
	incognito, err := rs.WebDriver.Incognito()
	require.NoError(t, err)

	rs.contexts = append(rs.contexts, incognito)

	p, err := incognito.Page(proto.TargetCreateTarget{URL: url})
	require.NoError(t, err)

	return p
}

func (rs *RodSession) doVisit(t *testing.T, page *rod.Page, url string) {
	require.NoError(t, page.Navigate(url))
	require.NoError(t, page.WaitLoad())
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
