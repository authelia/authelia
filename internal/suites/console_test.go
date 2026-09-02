//go:build !externalsuites
// +build !externalsuites

package suites

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testConsoleDocument = `<!doctype html>
<html><body><script>
console.warn('a warning');
console.error('an error');
Promise.reject(new Error('a rejection'));
setTimeout(() => { throw new Error('an exception') }, 0);
</script><div id="reported"></div></body></html>`

type consoleRecord struct {
	Installed bool `json:"installed"`
	Entries   []struct {
		Kind string `json:"kind"`
		Text string `json:"text"`
	} `json:"entries"`
}

func TestConsoleCollector(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		_, err := w.Write([]byte(testConsoleDocument))
		require.NoError(t, err)
	})

	server := httptest.NewServer(mux)

	t.Cleanup(server.Close)

	session := newVisitSession(t)

	read := func(t *testing.T, page interface {
		Eval(string, ...any) (*proto.RuntimeRemoteObject, error)
	}) consoleRecord {
		value, err := page.Eval(diagnosticsConsole)
		require.NoError(t, err)

		var record consoleRecord

		require.NoError(t, json.Unmarshal([]byte(value.Value.Str()), &record))

		return record
	}

	t.Run("ShouldReportThatItWasNotInstalled", func(t *testing.T) {
		page, err := session.WebDriver.Page(proto.TargetCreateTarget{URL: server.URL})
		require.NoError(t, err)

		session.WaitElementLocatedByID(t, page, "reported")

		record := read(t, page)

		assert.False(t, record.Installed, "an absent recording is reported as absent rather than as an empty one")
		assert.Empty(t, record.Entries)
	})

	// The tab is opened at about:blank and navigated, rather than created at its target, so that the
	// recording covers the first document too. A portal that fails on its very first load leaves no
	// later navigation to carry the collector.
	t.Run("ShouldRecordTheFirstDocumentOfATab", func(t *testing.T) {
		session := newVisitSession(t)

		page := session.doCreateTab(t, server.URL)

		session.WaitElementLocatedByID(t, page, "reported")

		record := read(t, page)

		require.True(t, record.Installed, "the first document a tab loads carries the recording")

		kinds := map[string]string{}

		for _, entry := range record.Entries {
			kinds[entry.Kind] = entry.Text
		}

		assert.Contains(t, kinds, "error")
	})

	t.Run("ShouldRecordWhatThePageReported", func(t *testing.T) {
		page, err := session.WebDriver.Page(proto.TargetCreateTarget{URL: "about:blank"})
		require.NoError(t, err)

		_, err = page.EvalOnNewDocument(consoleCollector)
		require.NoError(t, err)

		// Installed ahead of the navigation rather than on the document in front of it, which is how the
		// suites install it: the load a test fails on is a later one than the tab was opened at.
		require.NoError(t, page.Navigate(server.URL))

		session.WaitElementLocatedByID(t, page, "reported")

		record := read(t, page)

		require.True(t, record.Installed)

		kinds := map[string]string{}

		for _, entry := range record.Entries {
			kinds[entry.Kind] = entry.Text
		}

		assert.Contains(t, kinds, "warn")
		assert.Contains(t, kinds, "error")
		assert.Contains(t, kinds["error"], "an error")
	})
}
