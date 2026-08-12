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
	type tab struct {
		page *rod.Page
		err  error
	}

	created := make(chan tab, 1)

	go func() {
		browser, err := rs.WebDriver.Incognito()
		if err != nil {
			created <- tab{err: err}

			return
		}

		page, err := browser.Page(proto.TargetCreateTarget{URL: url})

		created <- tab{page: page, err: err}
	}()

	select {
	case result := <-created:
		require.NoError(t, result.err)
		require.NotNil(t, result.page)

		return result.page
	case <-time.After(time.Second * 10):
		require.FailNowf(t, "timeout creating tab", "the tab at '%s' was not created within 10 seconds", url)

		return nil
	}
}

func (rs *RodSession) doVisit(t *testing.T, page *rod.Page, url string) {
	assert.NoError(t, page.Navigate(url))
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
