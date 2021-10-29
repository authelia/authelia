package suites

import (
	"context"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRunPasswordComplexityScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestPasswordComplexityScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldRejectPasswordReset", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doVisit(s.Context(ctx), GetLoginBaseURL())
			s.verifyIsFirstFactorPage(t, s.Context(ctx))

			// Attempt to reset the password to a.
			s.doResetPassword(t, s.Context(ctx), "harry", "a", "a", true)
			s.verifyNotificationDisplayed(t, s.Context(ctx), "Your supplied password does not meet the password policy requirements.")
		})
	})
}
