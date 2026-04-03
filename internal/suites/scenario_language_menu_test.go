package suites

import (
	"context"
	"log"
	"time"
)

type LanguageMenuScenario struct {
	*RodSuite
}

func NewLanguageMenuScenario() *LanguageMenuScenario {
	return &LanguageMenuScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *LanguageMenuScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *LanguageMenuScenario) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *LanguageMenuScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *LanguageMenuScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *LanguageMenuScenario) TestShouldChangePreferredLanguage() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doVisitLoginPage(s.T(), s.Context(ctx), BaseDomain, "")

	menu := s.WaitElementLocatedByID(s.T(), s.Context(ctx), "language-button")
	s.Assert().NoError(menu.Click("left", 1))

	text, err := menu.Text()
	s.Assert().NoError(err)
	s.Assert().Equal("English", text)

	afrikaans := s.WaitElementLocatedByID(s.T(), s.Context(ctx), "language-af-ZA")
	s.Assert().NoError(afrikaans.Click("left", 1))

	s.Assert().NoError(s.WaitStable(time.Millisecond * 20))

	text, err = menu.Text()
	s.Assert().NoError(err)
	s.Assert().Equal("Afrikaans", text)

	button := s.WaitElementLocatedByID(s.T(), s.Context(ctx), "sign-in-button")

	text, err = button.Text()
	s.Assert().NoError(err)
	s.Assert().NotEqual("SIGN IN", text)

	s.Assert().NoError(menu.Click("left", 1))

	english := s.WaitElementLocatedByID(s.T(), s.Context(ctx), "language-en")
	s.Assert().NoError(english.Click("left", 1))

	s.Assert().NoError(s.WaitStable(time.Millisecond * 20))

	text, err = button.Text()
	s.Assert().NoError(err)
	s.Assert().Equal("SIGN IN", text)

	text, err = menu.Text()
	s.Assert().NoError(err)
	s.Assert().Equal("English", text)
}
