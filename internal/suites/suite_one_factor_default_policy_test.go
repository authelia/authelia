package suites

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OneFactorDefaultPolicySuite struct {
	suite.Suite
}

type OneFactorDefaultPolicyWebSuite struct {
	*SeleniumSuite
}

func NewOneFactorDefaultPolicyWebSuite() *OneFactorDefaultPolicyWebSuite {
	return &OneFactorDefaultPolicyWebSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *OneFactorDefaultPolicyWebSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *OneFactorDefaultPolicyWebSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OneFactorDefaultPolicyWebSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
}

// No target url is provided, then the user should be redirect to the default url.
func (s *OneFactorDefaultPolicyWebSuite) TestShouldRedirectUserToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
}

// Unsafe URL is provided, then the user should be redirect to the default url.
func (s *OneFactorDefaultPolicyWebSuite) TestShouldRedirectUserToDefaultURLWhenURLIsUnsafe() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "http://unsafe.local")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
}

// When use logged in and visit the portal again, she gets redirect to the authenticated view.
func (s *OneFactorDefaultPolicyWebSuite) TestShouldDisplayAuthenticatedView() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "http://unsafe.local")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
	s.doVisit(s.T(), LoginBaseURL)
	s.verifyIsAuthenticatedPage(ctx, s.T())
}

func (s *OneFactorDefaultPolicySuite) TestWeb() {
	suite.Run(s.T(), NewOneFactorDefaultPolicyWebSuite())
}

func TestOneFactorDefaultPolicySuite(t *testing.T) {
	suite.Run(t, new(OneFactorDefaultPolicySuite))
}
