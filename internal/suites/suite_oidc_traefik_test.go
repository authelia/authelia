package suites

import (
	"testing"

	"github.com/poy/onpar"
)

func TestOIDCTraefikSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	s := setupTest(t, "", true)
	teardownTest(s)

	TestRunOIDCScenario(t)
}
