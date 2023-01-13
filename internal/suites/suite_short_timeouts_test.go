package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ShortTimeoutsSuite struct {
	*RodSuite
}

func NewShortTimeoutsSuite() *ShortTimeoutsSuite {
	return &ShortTimeoutsSuite{
		RodSuite: &RodSuite{
			Name: shortTimeoutsSuiteName,
		},
	}
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

func (s *ShortTimeoutsSuite) SetupSuite() {
	s.LoadEnvironment()
}

func TestShortTimeoutsSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewShortTimeoutsSuite())
}
