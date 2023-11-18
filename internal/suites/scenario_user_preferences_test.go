package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type UserPreferencesScenario struct {
	*RodSuite
}

func NewUserPreferencesScenario() *UserPreferencesScenario {
	return &UserPreferencesScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *UserPreferencesScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *UserPreferencesScenario) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *UserPreferencesScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *UserPreferencesScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *UserPreferencesScenario) TestShouldRememberLastUsed2FAMethod() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	// Authenticate.
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))

	// Then switch to push notification method.
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "push-notification-method")

	// Switch context to clean up state in portal.
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	// Then go back to portal.
	s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURL(BaseDomain))
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	// And check the latest method is still used.
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "push-notification-method")
	// Meaning the authentication is successful.
	s.verifyIsHome(s.T(), s.Context(ctx))

	// Logout the user and see what user 'harry' sees.
	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "harry", "password", false, BaseDomain, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "push-notification-method")

	s.doLogout(s.T(), s.Context(ctx))
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

	// Then log back as previous user and verify the push notification is still the default method.
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "push-notification-method")
	s.verifyIsHome(s.T(), s.Context(ctx))

	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, BaseDomain, "")

	// Eventually restore the default method.
	s.doChangeMethod(s.T(), s.Context(ctx), "one-time-password")
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "one-time-password-method")
}

func TestUserPreferencesScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewUserPreferencesScenario())
}
