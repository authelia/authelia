package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRun1FAScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("Test1FAScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldAuthorizeSecretAfterOneFactor", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, targetURL)
			s.verifySecretAuthorized(t, s.Page)
		})

		o.Spec("TestShouldRedirectToSecondFactor", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, targetURL)
			s.verifyIsSecondFactorPage(t, s.Context(ctx))
		})

		o.Spec("TestShouldDenyAccessOnBadPassword", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
			s.doLoginOneFactor(t, s.Context(ctx), testUsername, badPassword, false, targetURL)
			s.verifyIsFirstFactorPage(t, s.Context(ctx))
			s.verifyNotificationDisplayed(t, s.Context(ctx), "Incorrect username or password.")
		})
	})
}
