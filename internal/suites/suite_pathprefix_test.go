package suites

import (
	"testing"

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

func (s *PathPrefixSuite) TestTwoFactorTOTPScenario() {
	suite.Run(s.T(), NewTwoFactorTOTPScenario())
}

func (s *PathPrefixSuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *PathPrefixSuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
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
