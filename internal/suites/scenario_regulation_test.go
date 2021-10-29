package suites

import (
	"context"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/poy/onpar"
)

func TestRunRegulationScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestRegulationScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldBanUserAfterTooManyAttempt", func(t *testing.T, s RodSuite) {
			is := is.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doVisitLoginPage(t, s.Context(ctx), "")
			s.doFillLoginPageAndClick(t, s.Context(ctx), "james", badPassword, false)
			s.verifyNotificationDisplayed(t, s.Context(ctx), "Incorrect username or password.")

			for i := 0; i < 3; i++ {
				err := s.WaitElementLocatedByID(t, s.Context(ctx), "password-textfield").Input(badPassword)
				is.NoErr(err)
				err = s.WaitElementLocatedByID(t, s.Context(ctx), "sign-in-button").Click("left")
				is.NoErr(err)
			}

			// Enter the correct password and test the regulation lock out.
			err := s.WaitElementLocatedByID(t, s.Context(ctx), "password-textfield").Input(testPassword)
			is.NoErr(err)
			err = s.WaitElementLocatedByID(t, s.Context(ctx), "sign-in-button").Click("left")
			is.NoErr(err)
			s.verifyNotificationDisplayed(t, s.Context(ctx), "Incorrect username or password.")

			s.verifyIsFirstFactorPage(t, s.Context(ctx))
			time.Sleep(10 * time.Second)

			// Enter the correct password and test a successful login.
			err = s.WaitElementLocatedByID(t, s.Context(ctx), "password-textfield").Input(testPassword)
			is.NoErr(err)
			err = s.WaitElementLocatedByID(t, s.Context(ctx), "sign-in-button").Click("left")
			is.NoErr(err)
			s.verifyIsSecondFactorPage(t, s.Context(ctx))
		})
	})
}
