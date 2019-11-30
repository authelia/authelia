package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DuoPushSuite struct {
	*SeleniumSuite
}

func NewDuoPushSuite() *DuoPushSuite {
	return &DuoPushSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *DuoPushSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *DuoPushSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *DuoPushSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
}

func (s *DuoPushSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.doChangeMethod(ctx, s.T(), "one-time-password")
	s.WaitElementLocatedByID(ctx, s.T(), "one-time-password-method")
}

func (s *DuoPushSuite) TestShouldSucceedAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.WaitElementLocatedByClassName(ctx, s.T(), "success-icon")
}

func (s *DuoPushSuite) TestShouldFailAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ConfigureDuo(s.T(), Deny)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.WaitElementLocatedByClassName(ctx, s.T(), "failure-icon")
}

func TestDuoPushSuite(t *testing.T) {
	suite.Run(t, NewDuoPushSuite())
	suite.Run(t, NewAvailableMethodsScenario([]string{
		"ONE-TIME PASSWORD",
		"PUSH NOTIFICATION",
	}))
	suite.Run(t, NewUserPreferencesScenario())
}
