package suites

import (
	"context"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRunResetPasswordScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestResetPasswordScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldResetPassword", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doVisit(s.Context(ctx), GetLoginBaseURL())
			s.verifyIsFirstFactorPage(t, s.Context(ctx))

			// Reset the password to abc.
			s.doResetPassword(t, s.Context(ctx), "harry", "abc", "abc", false)

			// Try to login with the old password.
			s.doLoginOneFactor(t, s.Context(ctx), "harry", testPassword, false, "")
			s.verifyNotificationDisplayed(t, s.Context(ctx), "Incorrect username or password.")

			// Try to login with the new password.
			s.doLoginOneFactor(t, s.Context(ctx), "harry", "abc", false, "")

			// Logout.
			s.doLogout(t, s.Context(ctx))

			// Reset the original password.
			s.doResetPassword(t, s.Context(ctx), "harry", testPassword, testPassword, false)
		})

		o.Spec("TestShouldMakeAttackerThinkPasswordResetIsInitiated", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doVisit(s.Context(ctx), GetLoginBaseURL())
			s.verifyIsFirstFactorPage(t, s.Context(ctx))

			// Try to initiate a password reset of a nonexistent user.
			s.doInitiatePasswordReset(t, s.Context(ctx), "i_dont_exist")

			// Check that the notification make the attacker thinks the process is initiated.
			s.verifyMailNotificationDisplayed(t, s.Context(ctx))
		})

		o.Spec("TestShouldLetUserNoticeThereIsAPasswordMismatch", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doVisit(s.Context(ctx), GetLoginBaseURL())
			s.verifyIsFirstFactorPage(t, s.Context(ctx))

			s.doInitiatePasswordReset(t, s.Context(ctx), testUsername)
			s.verifyMailNotificationDisplayed(t, s.Context(ctx))

			s.doCompletePasswordReset(t, s.Context(ctx), testPassword, "another_password")
			s.verifyNotificationDisplayed(t, s.Context(ctx), "Passwords do not match.")
		})
	})
}
