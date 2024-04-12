package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type EnvoySuite struct {
	*RodSuite
}

func NewEnvoySuite() *EnvoySuite {
	return &EnvoySuite{
		RodSuite: NewRodSuite(envoySuiteName),
	}
}

func (s *EnvoySuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *EnvoySuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *EnvoySuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *EnvoySuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func TestEnvoySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewEnvoySuite())
}
