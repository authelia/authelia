//go:build !externalsuites
// +build !externalsuites

package suites

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()

	startArtifactWatchdog()

	code := m.Run()

	discardWatchdogArtifacts()

	closeSharedBrowsers()

	os.Exit(code)
}
