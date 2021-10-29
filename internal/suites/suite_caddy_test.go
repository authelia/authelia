package suites

import (
	"testing"

	"github.com/poy/onpar"
)

func TestCaddySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	s := setupTest(t, "", true)
	teardownTest(s)

	o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
		s := setupTest(t, "", false)
		return t, s
	})

	o.AfterEach(func(t *testing.T, s RodSuite) {
		teardownTest(s)
	})

	TestRun1FAScenario(t)
	TestRun2FAScenario(t)
	TestRunCustomHeadersScenario(t)
	TestRunResetPasswordScenario(t)
}
