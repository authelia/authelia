package suites

import (
	"context"
	"log"
	"time"

	"github.com/authelia/authelia/v4/internal/utils"
)

type AvailableMethodsScenario struct {
	*RodSuite

	methods []string
}

func NewAvailableMethodsScenario(methods []string) *AvailableMethodsScenario {
	return &AvailableMethodsScenario{
		RodSuite: NewRodSuite(""),
		methods:  methods,
	}
}

func (s *AvailableMethodsScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *AvailableMethodsScenario) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *AvailableMethodsScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *AvailableMethodsScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *AvailableMethodsScenario) TestShouldCheckAvailableMethods() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")

	methodsButton := s.WaitElementLocatedByID(s.T(), s.Context(ctx), "methods-button")
	err := methodsButton.Click("left", 1)
	s.Assert().NoError(err)

	methodsDialog := s.WaitElementLocatedByID(s.T(), s.Context(ctx), "methods-dialog")
	options, err := methodsDialog.Elements(".method-option")
	s.Assert().NoError(err)
	s.Assert().Len(options, len(s.methods))

	optionsList := make([]string, 0, len(options))

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
