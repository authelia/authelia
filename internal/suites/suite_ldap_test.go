package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LDAPSuite struct {
	*RodSuite
}

func NewLDAPSuite() *LDAPSuite {
	return &LDAPSuite{RodSuite: new(RodSuite)}
}

func (s *LDAPSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *LDAPSuite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
}

func (s *LDAPSuite) TestResetPassword() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *LDAPSuite) TestPasswordComplexity() {
	suite.Run(s.T(), NewPasswordComplexityScenario())
}

func (s *LDAPSuite) TestSigninEmailScenario() {
	suite.Run(s.T(), NewSigninEmailScenario())
}

func TestLDAPSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewLDAPSuite())
}
