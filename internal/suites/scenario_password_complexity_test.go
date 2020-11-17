package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PasswordComplexityScenario struct {
	*SeleniumSuite
}

func NewPasswordComplexityScenario() *PasswordComplexityScenario {
	return &PasswordComplexityScenario{SeleniumSuite: new(SeleniumSuite)}
}

func (s *PasswordComplexityScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *PasswordComplexityScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *PasswordComplexityScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *PasswordComplexityScenario) TestShouldRejectPasswordReset() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s.doVisit(s.T(), GetLoginBaseURL())
	s.verifyIsFirstFactorPage(ctx, s.T())

	// Attempt to reset the password to a
	s.doResetPassword(ctx, s.T(), "john", "a", "a", true)
	s.verifyNotificationDisplayed(ctx, s.T(), "Your supplied password does not meet the password policy requirements.")
}

func TestRunPasswordComplexityScenario(t *testing.T) {
	suite.Run(t, NewPasswordComplexityScenario())
}
