package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RegulationScenario struct {
	*SeleniumSuite
}

func NewRegulationScenario() *RegulationScenario {
	return &RegulationScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *RegulationScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *RegulationScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *RegulationScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *RegulationScenario) TestShouldBanUserAfterTooManyAttempt() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.doVisitLoginPage(ctx, s.T(), "")
	s.doFillLoginPageAndClick(ctx, s.T(), "john", "bad-password", false)
	s.verifyNotificationDisplayed(ctx, s.T(), "Incorrect username or password.")

	for i := 0; i < 3; i++ {
		s.WaitElementLocatedByID(ctx, s.T(), "password-textfield").SendKeys("bad-password")
		s.WaitElementLocatedByID(ctx, s.T(), "sign-in-button").Click()
		time.Sleep(1 * time.Second)
	}

	// Enter the correct password and test the regulation lock out
	s.WaitElementLocatedByID(ctx, s.T(), "password-textfield").SendKeys("password")
	s.WaitElementLocatedByID(ctx, s.T(), "sign-in-button").Click()
	s.verifyNotificationDisplayed(ctx, s.T(), "Incorrect username or password.")

	time.Sleep(1 * time.Second)
	s.verifyIsFirstFactorPage(ctx, s.T())

	time.Sleep(9 * time.Second)

	// Enter the correct password and test a successful login
	s.WaitElementLocatedByID(ctx, s.T(), "password-textfield").SendKeys("password")
	s.WaitElementLocatedByID(ctx, s.T(), "sign-in-button").Click()
	s.verifyIsSecondFactorPage(ctx, s.T())
}

func TestBlacklistingScenario(t *testing.T) {
	suite.Run(t, NewRegulationScenario())
}
