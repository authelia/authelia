package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type UserPreferencesScenario struct {
	*SeleniumSuite
}

func NewUserPreferencesScenario() *UserPreferencesScenario {
	return &UserPreferencesScenario{
		SeleniumSuite: new(SeleniumSuite),
	}
}

func (s *UserPreferencesScenario) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *UserPreferencesScenario) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *UserPreferencesScenario) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *UserPreferencesScenario) TestShouldRememberLastUsed2FAMethod() {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	// Authenticate
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyIsSecondFactorPage(ctx, s.T())

	// Then switch to push notification method
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.WaitElementLocatedByID(ctx, s.T(), "push-notification-method")

	// Switch context to clean up state in portal.
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())

	// Then go back to portal.
	s.doVisit(s.T(), LoginBaseURL)
	s.verifyIsSecondFactorPage(ctx, s.T())
	// And check the latest method is still used.
	s.WaitElementLocatedByID(ctx, s.T(), "push-notification-method")
	// Meaning the authentication is successful
	s.verifyIsHome(ctx, s.T())

	// Logout the user and see what user 'harry' sees.
	s.doLogout(ctx, s.T())
	s.doLoginOneFactor(ctx, s.T(), "harry", "password", false, "")
	s.verifyIsSecondFactorPage(ctx, s.T())
	s.WaitElementLocatedByID(ctx, s.T(), "one-time-password-method")

	s.doLogout(ctx, s.T())
	s.verifyIsFirstFactorPage(ctx, s.T())

	// Then log back as previous user and verify the push notification is still the default method
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyIsSecondFactorPage(ctx, s.T())
	s.WaitElementLocatedByID(ctx, s.T(), "push-notification-method")
	s.verifyIsHome(ctx, s.T())

	s.doLogout(ctx, s.T())
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")

	// Eventually restore the default method
	s.doChangeMethod(ctx, s.T(), "one-time-password")
	s.WaitElementLocatedByID(ctx, s.T(), "one-time-password-method")
}

func TestUserPreferencesScenario(t *testing.T) {
	suite.Run(t, NewUserPreferencesScenario())
}
