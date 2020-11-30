package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LDAPSuite struct {
	*SeleniumSuite
}

func NewLDAPSuite() *LDAPSuite {
	return &LDAPSuite{SeleniumSuite: new(SeleniumSuite)}
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
	suite.Run(t, NewLDAPSuite())
}
