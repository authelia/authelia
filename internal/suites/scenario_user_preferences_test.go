package suites

import (
	"context"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRunUserPreferencesScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestUserPreferencesScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldRememberLastUsed2FAMethod", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			// Authenticate.
			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
			s.verifyIsSecondFactorPage(t, s.Context(ctx))

			// Then switch to push notification method.
			s.doChangeMethod(t, s.Context(ctx), "push-notification")
			s.WaitElementLocatedByID(t, s.Context(ctx), "push-notification-method")

			// Switch context to clean up state in portal.
			s.doVisit(s.Context(ctx), HomeBaseURL)
			s.verifyIsHome(t, s.Context(ctx))

			// Then go back to portal.
			s.doVisit(s.Context(ctx), GetLoginBaseURL())
			s.verifyIsSecondFactorPage(t, s.Context(ctx))
			// And check the latest method is still used.
			s.WaitElementLocatedByID(t, s.Context(ctx), "push-notification-method")
			// Meaning the authentication is successful.
			s.verifyIsHome(t, s.Context(ctx))

			// Logout the user and see what user 'harry' sees.
			s.doLogout(t, s.Context(ctx))
			s.doLoginOneFactor(t, s.Context(ctx), "harry", testPassword, false, "")
			s.verifyIsSecondFactorPage(t, s.Context(ctx))
			s.WaitElementLocatedByID(t, s.Context(ctx), "one-time-password-method")

			s.doLogout(t, s.Context(ctx))
			s.verifyIsFirstFactorPage(t, s.Context(ctx))

			// Then log back as previous user and verify the push notification is still the default method.
			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
			s.verifyIsSecondFactorPage(t, s.Context(ctx))
			s.WaitElementLocatedByID(t, s.Context(ctx), "push-notification-method")
			s.verifyIsHome(t, s.Context(ctx))

			s.doLogout(t, s.Context(ctx))
			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")

			// Eventually restore the default method.
			s.doChangeMethod(t, s.Context(ctx), "one-time-password")
			s.WaitElementLocatedByID(t, s.Context(ctx), "one-time-password-method")
		})
	})
}
