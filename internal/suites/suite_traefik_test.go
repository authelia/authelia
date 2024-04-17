package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TraefikSuite struct {
	*RodSuite
}

func NewTraefikSuite() *TraefikSuite {
	return &TraefikSuite{
		RodSuite: NewRodSuite(traefikSuiteName),
	}
}

func (s *TraefikSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *TraefikSuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *TraefikSuite) TestRedirectionURLScenario() {
	suite.Run(s.T(), NewRedirectionURLScenario())
}

func (s *TraefikSuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func TestTraefikSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTraefikSuite())
}
