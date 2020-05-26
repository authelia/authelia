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

func TestPathPrefixSuite(t *testing.T) {
	suite.Run(t, NewPathPrefixSuite())
}
