package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OneFactorOnlySuite struct {
	suite.Suite
}

type OneFactorOnlyWebSuite struct {
	*SeleniumSuite
}

func NewOneFactorOnlyWebSuite() *OneFactorOnlyWebSuite {
	return &OneFactorOnlyWebSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *OneFactorOnlyWebSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *OneFactorOnlyWebSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OneFactorOnlyWebSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
}

// No target url is provided, then the user should be redirect to the default url.
func (s *OneFactorOnlyWebSuite) TestShouldRedirectUserToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
}

// Unsafe URL is provided, then the user should be redirect to the default url.
func (s *OneFactorOnlyWebSuite) TestShouldRedirectUserToDefaultURLWhenURLIsUnsafe() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "http://unsafe.local")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
}

// When use logged in and visit the portal again, she gets redirect to the authenticated view.
func (s *OneFactorOnlyWebSuite) TestShouldDisplayAuthenticatedView() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "http://unsafe.local")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
	s.doVisit(s.T(), LoginBaseURL)
	s.verifyIsAuthenticatedPage(ctx, s.T())
}

func (s *OneFactorOnlySuite) TestWeb() {
	suite.Run(s.T(), NewOneFactorOnlyWebSuite())
}

func TestOneFactorOnlySuite(t *testing.T) {
	suite.Run(t, new(OneFactorOnlySuite))
}
