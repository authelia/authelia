package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CaddySuite struct {
	*RodSuite
}

func NewCaddySuite() *CaddySuite {
	return &CaddySuite{
		RodSuite: NewRodSuite(caddySuiteName),
	}
}

func (s *CaddySuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *CaddySuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *CaddySuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *CaddySuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func TestCaddySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewCaddySuite())
}
