package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LDAPSuite struct {
	*RodSuite
}

func NewLDAPSuite() *LDAPSuite {
	return &LDAPSuite{
		RodSuite: NewRodSuite(ldapSuiteName),
	}
}

func (s *LDAPSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *LDAPSuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
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

func (s *LDAPSuite) TestChangePasswordScenario() {
	suite.Run(s.T(), NewChangePasswordScenario())
}

func (s *LDAPSuite) TestUserManagementAPIScenario() {
	suite.Run(s.T(), NewUserManagementAPIScenario())
}

func (s *LDAPSuite) TestUserManagementOpenLDAPScenario() {
	suite.Run(s.T(), NewUserManagementOpenLDAPScenario())
}

func TestLDAPSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewLDAPSuite())
}
