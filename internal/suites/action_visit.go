package suites

import (
	"context"
	"fmt"
	"strings"
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
//
// A document that has not finished loading has no root either, and the responses this looks at are the
// ones it is still waiting on, so it has to be told apart from a page that never has a root. Reported
// separately from a pending render because only the latter is the page taking too long to draw: the
// caller gives loading whatever time the page context allows, the same as waiting on the load event did,
// and holds the render budget for the question this is actually asking.
const pageRenderState = `() => {
	if (document.readyState !== 'complete') {
		return 'loading';
	}

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

	page := s.Context(ctx)

	s.doReloadIfRenderWasLost(s.T(), page, url)

	s.verifyIsHome(s.T(), page)
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

		page, err := browser.Page(proto.TargetCreateTarget{})
		if err != nil {
			created <- tab{err: err}

			return
		}

		// Installed before the tab reaches its target, because a document that has already loaded cannot
		// be given the recording afterwards, and the first page of a test is the one these failures land
		// on: a portal that never renders renders nothing to navigate away from.
		if _, err = page.EvalOnNewDocument(consoleCollector); err != nil {
			log.Debugf("Error installing the console collector: %v", err)
		}

		err = rs.doNavigate(page, url)

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

func (rs *RodSession) doNavigate(page *rod.Page, url string) error {
	err := page.Navigate(url)

	// Chrome reports this when its certificate verifier is reconfigured while a request is already in
	// flight, which it only does while it is still starting: both occurrences in CI were the first
	// navigation a suite made, within half a second of the suite beginning. Sent again rather than
	// failing a test against a browser that had not finished coming up.
	if err != nil && strings.Contains(err.Error(), "net::ERR_CERT_VERIFIER_CHANGED") {
		return page.Navigate(url)
	}

	return err
}

func (rs *RodSession) doVisit(t *testing.T, page *rod.Page, url string) {
	require.NoError(t, rs.doNavigate(page, url))

	rs.doReloadIfRenderWasLost(t, page, url)
}

func (rs *RodSession) doReloadIfRenderWasLost(t *testing.T, page *rod.Page, url string) {
	if !rs.renderWasLost(page) {
		return
	}

	log.Warnf("The page at '%s' lost a response it needed to render, fetching it again once", url)

	require.NoError(t, page.Reload())
}

func (rs *RodSession) renderWasLost(page *rod.Page) bool {
	deadline := time.Now().Add(pageRenderTimeout)

	for {
		value, err := page.Eval(pageRenderState)
		if err != nil {
			// The page carries the deadline of the test that owns it, and that test reports its expiry
			// against the element it was waiting for, which names the failure better than this can.
			if page.GetContext().Err() != nil {
				return false
			}

			// Anything else is the execution context going away underneath the probe, which is what a
			// navigation committing looks like from here. The next document is the one being asked
			// about, so wait for it rather than reporting on the one that has just gone.
			if time.Now().After(deadline) {
				return false
			}

			time.Sleep(elementRetryInterval)

			continue
		}

		state := value.Value.Str()

		if state == "rendered" {
			return false
		}

		// The budget is for the render, not for the load in front of it, so it only starts once the
		// responses this reports on have all been settled one way or the other. A document that never
		// finishes is left to the page context, which the failing element reports against.
		if state == "loading" {
			deadline = time.Now().Add(pageRenderTimeout)
		} else if time.Now().After(deadline) {
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
