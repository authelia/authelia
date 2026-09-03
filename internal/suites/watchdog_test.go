//go:build !externalsuites
// +build !externalsuites

package suites

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The watchdog collects speculatively, while the tests may still be about to finish, so a run that does
// finish has to take back what it wrote. Without this a suite that merely ran close to its budget uploads
// a set of timeout artifacts and is annotated as though it had failed.
//
// Run under both path layouts because they are different directories: under CI the artifacts are written
// where the pipeline collects them from, and a discard that only cleaned the local one would leave
// exactly the files this exists to remove.
func TestDiscardWatchdogArtifacts(t *testing.T) {
	for _, tc := range []struct{ name, ci string }{
		{"Local", ""},
		{"CI", "true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SUITE", "WatchdogUnit")
			t.Setenv("CI", tc.ci)

			base := "TIMEOUT-WATCHDOGUNIT"

			written := make([]string, 0, 3)

			for _, suffix := range []string{"-page-0.png", "-page-0.console.json", ".containers.log"} {
				path, _ := screenshotPaths(base + suffix)

				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
				require.NoError(t, os.WriteFile(path, []byte("x"), 0600))

				written = append(written, path)
			}

			t.Cleanup(func() {
				for _, path := range written {
					_ = os.Remove(path)
				}
			})

			watchdogMutex.Lock()
			watchdogCollected = base
			watchdogMutex.Unlock()

			discardWatchdogArtifacts()

			for _, path := range written {
				assert.NoFileExists(t, path, "a run that finished discards what the watchdog collected")
			}
		})
	}

	t.Run("ShouldDoNothingWhenNothingWasCollected", func(t *testing.T) {
		require.NotPanics(t, discardWatchdogArtifacts)
	})

	t.Run("ShouldWaitForACollectionThatHasNotReachedTheLock", func(t *testing.T) {
		t.Setenv("SUITE", "WatchdogUnit")

		base := "TIMEOUT-WATCHDOGUNIT"
		path, _ := screenshotPaths(base + "-page-0.png")

		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))

		t.Cleanup(func() { _ = os.Remove(path) })

		started := make(chan struct{})

		done := make(chan struct{})

		watchdogDone = done
		watchdogTimer = time.AfterFunc(time.Millisecond, func() {
			defer close(done)

			close(started)

			// Stands for the interval between the callback being scheduled and it taking the lock, which
			// is the window the completion signal exists to cover.
			time.Sleep(time.Millisecond * 200)

			watchdogMutex.Lock()
			defer watchdogMutex.Unlock()

			require.NoError(t, os.WriteFile(path, []byte("x"), 0600))

			watchdogCollected = base
		})

		<-started

		discardWatchdogArtifacts()

		// Waited on so the assertion cannot pass merely by running before the collection wrote anything:
		// a discard that returned early leaves the file behind, and by here it exists.
		<-done

		assert.NoFileExists(t, path, "the discard waits for the collection rather than passing it")

		watchdogTimer, watchdogDone = nil, nil
	})
}

// The external suites ask for a browser of their own rather than a shared one, which is a different path
// out of NewRodSession and used to leave that browser out of what the watchdog could reach: a run that
// expired reported no pages and wrote nothing at all, which is the state the suites it covers were in.
func TestCollectWatchdogArtifactsCapturesAnUnsharedBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	t.Setenv("SUITE", "WatchdogUnit")
	t.Setenv("CI", "")

	session, err := NewRodSession(RodSessionWithoutDevtools())
	require.NoError(t, err)

	t.Cleanup(func() { _ = session.Stop() })

	page, err := session.WebDriver.Page(proto.TargetCreateTarget{
		URL: "data:text/html," + url.PathEscape(`<html><body><div id="captured">captured</div></body></html>`),
	})
	require.NoError(t, err)

	session.WaitElementLocatedByID(t, page, "captured")

	base := "TIMEOUT-WATCHDOGUNIT"

	t.Cleanup(func() {
		pattern, _ := screenshotPaths(base + "*")

		matches, _ := filepath.Glob(pattern)

		for _, match := range matches {
			_ = os.Remove(match)
		}

		watchdogMutex.Lock()
		watchdogCollected = ""
		watchdogMutex.Unlock()
	})

	collectWatchdogArtifacts()

	shot, _ := screenshotPaths(base + "-page-0.png")

	assert.FileExists(t, shot, "a browser that was not shared is captured when the run runs out of time")
}
