package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PathPrefixSuite struct {
	*RodSuite
}

func NewPathPrefixSuite() *PathPrefixSuite {
	return &PathPrefixSuite{RodSuite: new(RodSuite)}
}

func (s *PathPrefixSuite) Test11CheckEnv() {
	s.Assert().Equal("/auth", GetPathPrefix())
}

func (s *PathPrefixSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *PathPrefixSuite) Test2FAScenario() {
	suite.Run(s.T(), New2FAScenario())
}

func (s *PathPrefixSuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *PathPrefixSuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func TestPathPrefixSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewPathPrefixSuite())
}
