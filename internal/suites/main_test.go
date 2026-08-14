//go:build !externalsuites
// +build !externalsuites

package suites

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()

	closeSharedBrowsers()

	os.Exit(code)
}
