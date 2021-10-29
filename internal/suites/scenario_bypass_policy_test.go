package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestRunBypassPolicyScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestBypassPolicyScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldAccessPublicResource", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doVisit(s.Context(ctx), AdminBaseURL)
			s.verifyIsFirstFactorPage(t, s.Context(ctx))

			s.doVisit(s.Context(ctx), fmt.Sprintf("%s/secret.html", PublicBaseURL))
			s.verifySecretAuthorized(t, s.Context(ctx))
		})
	})
}
