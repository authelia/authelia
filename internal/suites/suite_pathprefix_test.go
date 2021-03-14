package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PathPrefixSuite struct {
	*SeleniumSuite
}

func NewPathPrefixSuite() *PathPrefixSuite {
	return &PathPrefixSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *PathPrefixSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *PathPrefixSuite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
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
