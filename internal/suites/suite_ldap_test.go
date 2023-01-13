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
		RodSuite: &RodSuite{
			Name: ldapSuiteName,
		},
	}
}

func (s *LDAPSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *LDAPSuite) Test2FAScenario() {
	suite.Run(s.T(), New2FAScenario())
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

func (s *LDAPSuite) SetupSuite() {
	s.LoadEnvironment()
}

func TestLDAPSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewLDAPSuite())
}
