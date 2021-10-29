package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRunSigninEmailScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestSigninEmailScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		// This scenario is used to test sign in using the user email address.
		o.Spec("TestShouldSignInWithUserEmail", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
			s.doLoginOneFactor(t, s.Context(ctx), "john.doe@authelia.com", testPassword, false, targetURL)
			s.verifySecretAuthorized(t, s.Context(ctx))
		})
	})
}
