package suites

import (
	"testing"

	"github.com/poy/onpar"
)

func TestPathPrefixSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	s := setupTest(t, "", true)
	teardownTest(s)

	TestRun1FAScenario(t)
	TestRun2FAScenario(t)
	TestRunCustomHeadersScenario(t)
	TestRunResetPasswordScenario(t)
}
