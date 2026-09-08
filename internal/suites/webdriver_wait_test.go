//go:build !externalsuites
// +build !externalsuites

package suites

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWaitDocument = `<!doctype html>
<html><body><div id="present">here</div>
<iframe srcdoc="<html><body><strong id='inner'>framed</strong></body></html>"></iframe></body></html>`

func newWaitServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		_, err := w.Write([]byte(testWaitDocument))
		require.NoError(t, err)
	})

	server := httptest.NewServer(mux)

	t.Cleanup(server.Close)

	return server
}

func TestWaitElementLocatedBySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	server := newWaitServer(t)
	session := newVisitSession(t)

	newPage := func(t *testing.T) *rod.Page {
		page, err := session.WebDriver.Page(proto.TargetCreateTarget{URL: server.URL})
		require.NoError(t, err)

		return page
	}

	t.Run("ShouldReturnAnElementThatOutlivesTheWait", func(t *testing.T) {
		page := newPage(t)

		element := session.WaitElementLocatedByID(t, page, "present")
		require.NotNil(t, element)

		// The wait bounds itself with a context it cancels on the way out, so an element still bound to it
		// would be unusable by the time the caller received it.
		text, err := element.Text()
		require.NoError(t, err)
		assert.Equal(t, "here", text)
	})

	// The element carries the page it was found on and hands it to everything derived from it, so an
	// element rebound to the caller's context while still holding the page the wait bounded descends into
	// a frame that is already cancelled. This is what the templates suite hit.
	t.Run("ShouldReturnAnElementWhosePageOutlivesTheWait", func(t *testing.T) {
		page := newPage(t)

		element := session.WaitElementLocatedBySelector(t, page, "iframe")
		require.NotNil(t, element)

		frame, err := element.Frame()
		require.NoError(t, err, "the frame is descended into on the caller's context, not the wait's")

		inner, err := frame.Element("#inner")
		require.NoError(t, err)

		text, err := inner.Text()
		require.NoError(t, err)
		assert.Equal(t, "framed", text)
	})

	t.Run("ShouldGiveUpOnTheBudgetRatherThanThePageContext", func(t *testing.T) {
		page := newPage(t)

		// A page context far longer than the budget, so whichever of the two ends the wait is unambiguous.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

		defer cancel()

		previous := elementLocateBudget
		elementLocateBudget = 2 * time.Second

		defer func() { elementLocateBudget = previous }()

		reported := &testing.T{}
		started := time.Now()

		// Runs on its own goroutine because the failure is reported with FailNow, which ends the goroutine
		// it happens on rather than returning.
		done := make(chan struct{})

		go func() {
			defer close(done)

			session.WaitElementLocatedByID(reported, page.Context(ctx), "absent")
		}()

		<-done

		elapsed := time.Since(started)

		// The ceiling is the budget the wait was given plus the state it gathers to report with the
		// failure, so that raising either constant does not leave this asserting against a stale number.
		ceiling := elementLocateBudget + elementStateTimeout + time.Second*5

		assert.True(t, reported.Failed(), "the wait reports a failure when the element never appears")
		assert.Less(t, elapsed, ceiling, "the wait ends on its own budget rather than the page context")
		assert.GreaterOrEqual(t, elapsed, elementLocateBudget, "the wait spends its budget before giving up")
	})

	t.Run("ShouldRaiseTheBudgetsWhileTheDatastoreIsDisrupted", func(t *testing.T) {
		require.Equal(t, elementLocateTimeout, elementLocateBudget)

		doWithDisruptedDatastore(func() {
			assert.Equal(t, elementLocateTimeoutDisrupted, elementLocateBudget)
			assert.Equal(t, elementActionTimeoutDisrupted, elementActionBudget)
		})

		assert.Equal(t, elementLocateTimeout, elementLocateBudget, "the budget is restored for whatever runs next")
		assert.Equal(t, elementActionTimeout, elementActionBudget, "the budget is restored for whatever runs next")
	})
}
