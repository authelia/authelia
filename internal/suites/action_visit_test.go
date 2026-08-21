package suites

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testVisitIndex = `<!doctype html>
<html><head><script type="module" src="./static/chunk.js"></script></head>
<body><div id="root"></div></body></html>`

const testVisitChunk = `document.getElementById('root').textContent = 'rendered';`

// visitFailure is how the server fails a chunk. Dropping the connection leaves the browser with no
// status to record, which is what a request cut mid-flight looks like, and is the only failure a visit
// fetches the page again for. Refusing the request leaves the page just as blank but with a status
// against it, which stands for every other reason a page does not render.
type visitFailure int

const (
	visitDropped visitFailure = iota
	visitRefused
)

func newVisitServer(t *testing.T, failure visitFailure, failFirstLoad bool) (server *httptest.Server, documents, requests *atomic.Int32) {
	documents, requests = &atomic.Int32{}, &atomic.Int32{}

	mux := http.NewServeMux()

	// Which document is asking decides whether the chunk fails, rather than a count of failures. How
	// many times the browser retries a request of its own accord is its business, and a count would let
	// a change to it serve the chunk within the first load, leaving the test passing without a reload.
	mux.HandleFunc("/static/chunk.js", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		switch {
		case !failFirstLoad || documents.Load() > 1:
			w.Header().Set("Content-Type", "text/javascript")

			_, err := w.Write([]byte(testVisitChunk))
			require.NoError(t, err)
		case failure == visitDropped:
			connection, _, err := w.(http.Hijacker).Hijack()
			require.NoError(t, err)
			require.NoError(t, connection.Close())
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	// This pattern is the mux's catch-all, so the favicon the browser asks for on its own lands here
	// too and only a request for the document itself counts as a load.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			documents.Add(1)
		}

		w.Header().Set("Content-Type", "text/html")

		_, err := w.Write([]byte(testVisitIndex))
		require.NoError(t, err)
	})

	server = httptest.NewServer(mux)

	t.Cleanup(server.Close)

	return server, documents, requests
}

func newVisitSession(t *testing.T) *RodSession {
	path, err := GetBrowserPath()
	require.NoError(t, err)

	l := launcher.New().Bin(path).Headless(true)

	t.Cleanup(l.Cleanup)

	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()

	t.Cleanup(func() { _ = browser.Close() })

	return &RodSession{WebDriver: browser}
}

func (rs *RodSession) mustVisitAndReadRoot(t *testing.T, url string) string {
	page, err := rs.WebDriver.Page(proto.TargetCreateTarget{URL: "about:blank"})
	require.NoError(t, err)

	rs.doVisit(t, page, url)

	value, err := page.Eval(`() => document.getElementById('root').textContent`)
	require.NoError(t, err)

	return value.Value.Str()
}

// TestVisitShouldFetchAgainAPageThatLostAResponse is the regression guard for the flake where a page
// whose module graph was cut mid-fetch stayed blank for the rest of the test: no element wait recovers
// it, because the document has loaded and nothing further is ever fetched.
func TestVisitShouldFetchAgainAPageThatLostAResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	server, documents, _ := newVisitServer(t, visitDropped, true)
	session := newVisitSession(t)

	assert.Equal(t, "rendered", session.mustVisitAndReadRoot(t, server.URL))
	assert.Equal(t, int32(2), documents.Load(), "the page is fetched again exactly once, after the browser gave up retrying")
}

// TestVisitShouldNotFetchAgainAPageThatRendered pins the second fetch to pages that need it, so that
// every other visit does not pay for it.
func TestVisitShouldNotFetchAgainAPageThatRendered(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	server, documents, requests := newVisitServer(t, visitDropped, false)
	session := newVisitSession(t)

	assert.Equal(t, "rendered", session.mustVisitAndReadRoot(t, server.URL))
	assert.Equal(t, int32(1), documents.Load(), "the page is fetched once")
	assert.Equal(t, int32(1), requests.Load(), "the chunk is fetched once")
}

// TestVisitShouldNotFetchAgainAPageThatWasAnswered keeps a blank page that was answered, rather than cut
// off, failing the way it always has. Fetching those again would hide real breakage and would spend a
// second page load out of the budget of the test that is already failing.
func TestVisitShouldNotFetchAgainAPageThatWasAnswered(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	server, documents, requests := newVisitServer(t, visitRefused, true)
	session := newVisitSession(t)

	assert.Empty(t, session.mustVisitAndReadRoot(t, server.URL))
	assert.Equal(t, int32(1), documents.Load(), "the page is fetched once")
	assert.Equal(t, int32(1), requests.Load(), "the chunk is fetched once")
}
