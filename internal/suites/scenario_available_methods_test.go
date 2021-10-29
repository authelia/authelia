package suites

import (
	"context"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/poy/onpar"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestRunAvailableMethodsScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestAvailableMethodsScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldCheckAvailableMethods", func(t *testing.T, s RodSuite) {
			is := is.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")

			methodsButton := s.WaitElementLocatedByID(t, s.Context(ctx), "methods-button")
			err := methodsButton.Click("left")
			is.NoErr(err)

			methodsDialog := s.WaitElementLocatedByID(t, s.Context(ctx), "methods-dialog")
			options, err := methodsDialog.Elements(".method-option")
			is.NoErr(err)
			is.True(len(options) == len(methods))

			optionsList := make([]string, 0)

			for _, o := range options {
				txt, err := o.Text()
				is.NoErr(err)

				optionsList = append(optionsList, txt)
			}

			is.True(len(optionsList) == len(methods))

			for _, m := range methods {
				is.True(utils.IsStringInSlice(m, optionsList))
			}
		})
	})
}
