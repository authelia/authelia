package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ZoraxySuite struct {
	*RodSuite
}

func NewZoraxySuite() *ZoraxySuite {
	return &ZoraxySuite{
		RodSuite: NewRodSuite(zoraxySuiteName),
	}
}

func (s *ZoraxySuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *ZoraxySuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *ZoraxySuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *ZoraxySuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *ZoraxySuite) TestChangePasswordScenario() {
	suite.Run(s.T(), NewChangePasswordScenario())
}

func TestZoraxySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewZoraxySuite())
}
