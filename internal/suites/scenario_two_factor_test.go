package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRun2FAScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("Test2FAScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldAuthorizeSecretAfterTwoFactor", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			// Login with 1FA & 2FA.
			targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
			s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, targetURL)

			// And check if the user is redirected to the secret.
			s.verifySecretAuthorized(t, s.Context(ctx))

			// Leave the secret.
			s.doVisit(s.Context(ctx), HomeBaseURL)
			s.verifyIsHome(t, s.Context(ctx))

			// And try to reload it again to check the session is kept.
			s.doVisit(s.Context(ctx), targetURL)
			s.verifySecretAuthorized(t, s.Context(ctx))
		})

		o.Spec("TestShouldFailTwoFactor", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			wrongPasscode := "123456"

			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
			s.verifyIsSecondFactorPage(t, s.Context(ctx))
			s.doEnterOTP(t, s.Context(ctx), wrongPasscode)
			s.verifyNotificationDisplayed(t, s.Context(ctx), "The one-time password might be wrong")
		})
	})
}
