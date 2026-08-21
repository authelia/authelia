package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// A transport failure in the burst of requests the portal's entry module makes for its chunks leaves the
// module graph unresolved: the document reaches readyState complete and nothing renders or is fetched
// again. The lost response is what separates that page from one that is merely slow, and it reports a
// zero status against a zero body. Pages that are not the portal have no root and count as rendered.
const pageRenderState = `() => {
	const root = document.getElementById('root');

	if (root === null || root.innerHTML.length !== 0) {
		return 'rendered';
	}

	const lost = performance.getEntriesByType('resource').some((entry) => entry.responseStatus === 0 && entry.encodedBodySize === 0);

	return lost ? 'lost' : 'pending';
}`

func (s *RodSuite) doSetupTest(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), setupTestTimeout)
	defer cancel()

	s.Page = s.doCreateTab(s.T(), url)
	s.verifyIsHome(s.T(), s.Context(ctx))
}

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
	require.NoError(t, page.Navigate(url))
	require.NoError(t, page.WaitLoad())

	rs.doReloadIfRenderWasLost(t, page, url)
}

func (rs *RodSession) doReloadIfRenderWasLost(t *testing.T, page *rod.Page, url string) {
	if !rs.renderWasLost(page) {
		return
	}

	log.Warnf("The page at '%s' lost a response it needed to render, fetching it again once", url)

	require.NoError(t, page.Reload())
	require.NoError(t, page.WaitLoad())
}

func (rs *RodSession) renderWasLost(page *rod.Page) bool {
	deadline := time.Now().Add(pageRenderTimeout)

	for {
		value, err := page.Eval(pageRenderState)
		if err != nil {
			// The page carries the deadline of the test that owns it, and that test reports its expiry
			// against the element it was waiting for, which names the failure better than this can.
			return false
		}

		state := value.Value.Str()

		if state == "rendered" {
			return false
		}

		if time.Now().After(deadline) {
			return state == "lost"
		}

		time.Sleep(elementRetryInterval)
	}
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
