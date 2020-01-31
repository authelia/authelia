package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ShortTimeoutsSuite struct {
	*SeleniumSuite
}

func NewShortTimeoutsSuite() *ShortTimeoutsSuite {
	return &ShortTimeoutsSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *ShortTimeoutsSuite) TestDefaultRedirectionURLScenario() {
	suite.Run(s.T(), NewDefaultRedirectionURLScenario())
}

func (s *ShortTimeoutsSuite) TestInactivityScenario() {
	suite.Run(s.T(), NewInactivityScenario())
}

func (s *ShortTimeoutsSuite) TestRegulationScenario() {
	suite.Run(s.T(), NewRegulationScenario())
}

func TestShortTimeoutsSuite(t *testing.T) {
	suite.Run(t, NewShortTimeoutsSuite())
}
