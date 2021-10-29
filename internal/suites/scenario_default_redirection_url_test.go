package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRunDefaultRedirectionURLScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestDefaultRedirectionURLScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestUserIsRedirectedToDefaultURL", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)

			s.doVisit(s.Context(ctx), HomeBaseURL)
			s.verifyIsHome(t, s.Page)
			s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, targetURL)
			s.verifySecretAuthorized(t, s.Context(ctx))
			s.doLogout(t, s.Context(ctx))

			s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, "")
			s.verifyIsHome(t, s.Page)
		})
	})
}
