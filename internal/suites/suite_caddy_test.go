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

func (s *CaddySuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *CaddySuite) TestTwoFactorTOTPScenario() {
	suite.Run(s.T(), NewTwoFactorTOTPScenario())
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
