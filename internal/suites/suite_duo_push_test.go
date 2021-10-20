package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DuoPushWebDriverSuite struct {
	*RodSuite
}

func NewDuoPushWebDriverSuite() *DuoPushWebDriverSuite {
	return &DuoPushWebDriverSuite{RodSuite: new(RodSuite)}
}

func (s *DuoPushWebDriverSuite) SetupSuite() {
	browser, err := StartRod()

	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *DuoPushWebDriverSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *DuoPushWebDriverSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *DuoPushWebDriverSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		s.collectCoverage(s.Page)
		s.MustClose()
	}()

	s.doLogout(s.T(), s.Context(ctx))
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doChangeMethod(s.T(), s.Context(ctx), "one-time-password")
	s.WaitElementLocatedByCSSSelector(s.T(), s.Context(ctx), "one-time-password-method")
}

func (s *DuoPushWebDriverSuite) TestShouldSucceedAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	ConfigureDuo(s.T(), Allow)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyIsHome(s.T(), s.Context(ctx))
}

func (s *DuoPushWebDriverSuite) TestShouldFailAuthentication() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	ConfigureDuo(s.T(), Deny)

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.WaitElementLocatedByClassName(s.T(), s.Context(ctx), "failure-icon")
}

type DuoPushDefaultRedirectionSuite struct {
	*RodSuite
}

func NewDuoPushDefaultRedirectionSuite() *DuoPushDefaultRedirectionSuite {
	return &DuoPushDefaultRedirectionSuite{RodSuite: new(RodSuite)}
}

func (s *DuoPushDefaultRedirectionSuite) SetupSuite() {
	browser, err := StartRod()

	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *DuoPushDefaultRedirectionSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *DuoPushDefaultRedirectionSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *DuoPushDefaultRedirectionSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *DuoPushDefaultRedirectionSuite) TestUserIsRedirectedToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, "")
	s.doChangeMethod(s.T(), s.Context(ctx), "push-notification")
	s.verifyIsHome(s.T(), s.Page)
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
		"TIME-BASED ONE-TIME PASSWORD",
		"PUSH NOTIFICATION",
	}))
}

func (s *DuoPushSuite) TestUserPreferencesScenario() {
	suite.Run(s.T(), NewUserPreferencesScenario())
}

func TestDuoPushSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewDuoPushSuite())
}
