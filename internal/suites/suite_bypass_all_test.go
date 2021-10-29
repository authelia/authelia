package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestBypassAllSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

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

		s.doVisit(s.Context(ctx), fmt.Sprintf("%s/secret.html", AdminBaseURL))
		s.verifySecretAuthorized(t, s.Context(ctx))

		s.doVisit(s.Context(ctx), fmt.Sprintf("%s/secret.html", PublicBaseURL))
		s.verifySecretAuthorized(t, s.Context(ctx))
	})

	TestRunCustomHeadersScenario(t)
}
