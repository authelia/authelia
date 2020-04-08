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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
}

func (s *DuoPushWebDriverSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyIsSecondFactorPage(ctx, s.T())
	s.doChangeMethod(ctx, s.T(), "one-time-password")
	s.WaitElementLocatedByID(ctx, s.T(), "one-time-password-method")
}

func (s *DuoPushWebDriverSuite) TestShouldSucceedAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.verifyIsHome(ctx, s.T())
}

func (s *DuoPushWebDriverSuite) TestShouldFailAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ConfigureDuo(s.T(), Deny)

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.WaitElementLocatedByClassName(ctx, s.T(), "failure-icon")
}

type DuoPushDefaultRedirectionSuite struct {
	*SeleniumSuite
}

func NewDuoPushDefaultRedirectionSuite() *DuoPushDefaultRedirectionSuite {
	return &DuoPushDefaultRedirectionSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *DuoPushDefaultRedirectionSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *DuoPushDefaultRedirectionSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *DuoPushDefaultRedirectionSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
}

func (s *DuoPushDefaultRedirectionSuite) TestUserIsRedirectedToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.doChangeMethod(ctx, s.T(), "push-notification")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
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

func (s *DuoPushSuite) TestDuoPushRedirectionURLSuite() {
	suite.Run(s.T(), NewDuoPushDefaultRedirectionSuite())
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

func TestDuoPushSuite(t *testing.T) {
	suite.Run(t, NewDuoPushSuite())
}
