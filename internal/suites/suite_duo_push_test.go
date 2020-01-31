package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DuoPushWebDriverSuite struct {
	*SeleniumSuite
}

func NewDuoPushWebDriverSuite() *DuoPushWebDriverSuite {
	return &DuoPushWebDriverSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *DuoPushWebDriverSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *DuoPushWebDriverSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *DuoPushWebDriverSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
}

func (s *DuoPushWebDriverSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.doChangeMethod(ctx, s.T(), "one-time-password")
	s.WaitElementLocatedByID(ctx, s.T(), "one-time-password-method")
}

func (s *DuoPushWebDriverSuite) TestShouldSucceedAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.WaitElementLocatedByClassName(ctx, s.T(), "success-icon")
}

func (s *DuoPushWebDriverSuite) TestShouldFailAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ConfigureDuo(s.T(), Deny)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.WaitElementLocatedByClassName(ctx, s.T(), "failure-icon")
}

type DuoPushSuite struct {
	suite.Suite
}

func NewDuoPushSuite() *DuoPushSuite {
	return &DuoPushSuite{}
}

func (s *DuoPushSuite) TestDuoPushWebDriverSuite() {
	suite.Run(s.T(), NewDuoPushWebDriverSuite())
}

func (s *DuoPushSuite) TestAvailableMethodsScenario() {
	suite.Run(s.T(), NewAvailableMethodsScenario([]string{
		"ONE-TIME PASSWORD",
		"PUSH NOTIFICATION",
	}))
}

func (s *DuoPushSuite) TestUserPreferencesScenario() {
	suite.Run(s.T(), NewUserPreferencesScenario())
}

func (s *DuoPushSuite) TestDefaultRedirectionURLScenario() {
	suite.Run(s.T(), NewDefaultRedirectionURLScenario())
}

func TestDuoPushSuite(t *testing.T) {
	suite.Run(t, NewDuoPushSuite())
}
