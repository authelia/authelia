package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRunRedirectionURLScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestRedirectionURLScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldVerifyCustomURLParametersArePropagatedAfterRedirection", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			targetURL := fmt.Sprintf("%s/secret.html?myparam=test", SingleFactorBaseURL)
			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, targetURL)
			s.verifySecretAuthorized(t, s.Context(ctx))
			s.verifyURLIs(t, s.Context(ctx), targetURL)
		})
	})
}
