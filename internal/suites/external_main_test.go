//go:build externalsuites
// +build externalsuites

package suites

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
)

// globalDevServer is shared across external suites so the TestMain signal handler can stop the
// currently-running dev server on Ctrl+C even while a suite goroutine is blocked in a CDP wait.
// Each SetupSuite assigns it via the StartDevServer onSpawn callback; each TearDownSuite clears it.
var globalDevServer *DevServer

func TestMain(m *testing.M) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		if globalDevServer != nil {
			_ = globalDevServer.Stop()
		}

		os.Exit(130)
	}()

	os.Exit(m.Run())
}

func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}
