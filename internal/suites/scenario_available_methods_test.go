package suites

import (
	"context"
	"log"
	"time"

	"github.com/tebeka/selenium"

	"github.com/authelia/authelia/internal/utils"
)

type AvailableMethodsScenario struct {
	*SeleniumSuite

	methods []string
}

func NewAvailableMethodsScenario(methods []string) *AvailableMethodsScenario {
	return &AvailableMethodsScenario{
		SeleniumSuite: new(SeleniumSuite),
		methods:       methods,
	}
}

func (s *AvailableMethodsScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.SeleniumSuite.WebDriverSession = wds
}

func (s *AvailableMethodsScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *AvailableMethodsScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *AvailableMethodsScenario) TestShouldCheckAvailableMethods() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")

	methodsButton := s.WaitElementLocatedByID(ctx, s.T(), "methods-button")
	err := methodsButton.Click()
	s.Assert().NoError(err)

	methodsDialog := s.WaitElementLocatedByID(ctx, s.T(), "methods-dialog")
	options, err := methodsDialog.FindElements(selenium.ByClassName, "method-option")
	s.Assert().NoError(err)
	s.Assert().Len(options, len(s.methods))

	optionsList := make([]string, 0)

	for _, o := range options {
		txt, err := o.Text()
		s.Assert().NoError(err)

		optionsList = append(optionsList, txt)
	}

	s.Assert().Len(optionsList, len(s.methods))

	for _, m := range s.methods {
		s.Assert().True(utils.IsStringInSlice(m, optionsList))
	}
}
