package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRunInactivityScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestInactivityScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldRequireReauthenticationAfterInactivityPeriod", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

			s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")
			s.doVisit(s.Context(ctx), HomeBaseURL)
			s.verifyIsHome(t, s.Context(ctx))

			time.Sleep(6 * time.Second)

			s.doVisit(s.Context(ctx), targetURL)
			s.verifyIsFirstFactorPage(t, s.Context(ctx))
		})

		o.Spec("TestShouldRequireReauthenticationAfterCookieExpiration", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

			s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")

			for i := 0; i < 3; i++ {
				s.doVisit(s.Context(ctx), HomeBaseURL)
				s.verifyIsHome(t, s.Context(ctx))

				time.Sleep(2 * time.Second)

				s.doVisit(s.Context(ctx), targetURL)
				s.verifySecretAuthorized(t, s.Context(ctx))
			}

			time.Sleep(2 * time.Second)

			s.doVisit(s.Context(ctx), targetURL)
			s.verifyIsFirstFactorPage(t, s.Context(ctx))
		})

		o.Spec("TestShouldDisableCookieExpirationAndInactivity", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

			s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, true, secret, "")
			s.doVisit(s.Context(ctx), HomeBaseURL)
			s.verifyIsHome(t, s.Context(ctx))

			time.Sleep(10 * time.Second)

			s.doVisit(s.Context(ctx), targetURL)
			s.verifySecretAuthorized(t, s.Context(ctx))
		})
	})
}
