package suites

import (
	"testing"

	"github.com/poy/onpar"
)

func TestShortTimeoutsSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	s := setupTest(t, "", true)
	teardownTest(s)

	TestRunDefaultRedirectionURLScenario(t)
	TestRunInactivityScenario(t)
	TestRunRegulationScenario(t)
}
