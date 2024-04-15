package suites

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PathPrefixSuite struct {
	*RodSuite
}

func NewPathPrefixSuite() *PathPrefixSuite {
	return &PathPrefixSuite{
		RodSuite: NewRodSuite(pathPrefixSuiteName),
	}
}

func (s *PathPrefixSuite) TestCheckEnv() {
	s.Assert().Equal("/auth", GetPathPrefix())
}

func (s *PathPrefixSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *PathPrefixSuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *PathPrefixSuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *PathPrefixSuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *PathPrefixSuite) TestShouldRenderFrontendWithTrailingSlash() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectCoverage(s.Page)
		s.collectScreenshot(ctx.Err(), s.Page)
		s.MustClose()
		err := s.RodSession.Stop()
		s.Require().NoError(err)
	}()

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	s.Require().NoError(err)
	s.RodSession = browser

	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)

	s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURL(BaseDomain)+"/")
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
}

func (s *PathPrefixSuite) TestShouldRenderFrontendWithoutTrailingSlash() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectCoverage(s.Page)
		s.collectScreenshot(ctx.Err(), s.Page)
		s.MustClose()
		err := s.RodSession.Stop()
		s.Require().NoError(err)
	}()

	browser, err := NewRodSession(RodSessionWithCredentials(s))
	s.Require().NoError(err)
	s.RodSession = browser

	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)

	s.doVisit(s.T(), s.Context(ctx), GetLoginBaseURL(BaseDomain))
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
}

func (s *PathPrefixSuite) SetupSuite() {
	s.T().Setenv("PathPrefix", "/auth")
}

func TestPathPrefixSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewPathPrefixSuite())
}
