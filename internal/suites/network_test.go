// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

//go:build !externalsuites
// +build !externalsuites

package suites

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testNetworkDocument = `<!doctype html>
<html><body><script>
fetch('/ok').then((response) => response.text()).then(() => fetch('/drop', {method: 'PUT'})).catch(() => {}).then(() => {
	document.body.insertAdjacentHTML('beforeend', '<div id="reported"></div>');
});
</script></body></html>`

func TestNetworkFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		_, err := w.Write([]byte(testNetworkDocument))
		require.NoError(t, err)
	})

	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/drop", func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := w.(http.Hijacker).Hijack()
		require.NoError(t, err)
		require.NoError(t, conn.Close())
	})

	server := httptest.NewServer(mux)

	t.Cleanup(server.Close)

	session := newVisitSession(t)

	page := session.doCreateTab(t, server.URL)

	session.WaitElementLocatedByID(t, page, "reported")

	var failures []networkFailure

	require.Eventually(t, func() bool {
		data, ok := networkFailures(page)
		if !ok {
			return false
		}

		failures = nil

		if err := json.Unmarshal(data, &failures); err != nil {
			return false
		}

		for _, failure := range failures {
			if strings.HasSuffix(failure.URL, "/drop") {
				return true
			}
		}

		return false
	}, 5*time.Second, 50*time.Millisecond, "a tab created for a test records its failed requests")

	t.Run("ShouldRecordAFailedRequestWithTheReasonChromeGave", func(t *testing.T) {
		require.Len(t, failures, 1, "only the request that failed is recorded")

		assert.Equal(t, "PUT", failures[0].Method)
		assert.True(t, strings.HasSuffix(failures[0].URL, "/drop"), failures[0].URL)
		assert.Equal(t, "net::ERR_EMPTY_RESPONSE", failures[0].Error)
	})

	t.Run("ShouldWriteTheRecordWithTheFailureDiagnostics", func(t *testing.T) {
		t.Setenv("SUITE", "NetworkUnit")
		t.Setenv("CI", "")

		path, _ := screenshotPaths("TestNetworkFailures.network.json")

		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))

		t.Cleanup(func() {
			_ = os.RemoveAll(filepath.Dir(path))
		})

		session.collectDiagnostics(page, "TestNetworkFailures")

		data, err := os.ReadFile(path)
		require.NoError(t, err)

		var written []networkFailure

		require.NoError(t, json.Unmarshal(data, &written))
		assert.Equal(t, failures, written)
	})

	t.Run("ShouldStopRecordingOnceTheTabIsClosed", func(t *testing.T) {
		require.NoError(t, page.Close())

		assert.Eventually(t, func() bool {
			_, ok := networkFailures(page)

			return !ok
		}, 5*time.Second, 50*time.Millisecond, "the subscription ends with its tab rather than with the browser")
	})
}
