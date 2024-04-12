package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TwoFactorSuite struct {
	*BaseSuite
}

func NewTwoFactorSuite() *TwoFactorSuite {
	return &TwoFactorSuite{
		BaseSuite: &BaseSuite{
			Name: standaloneSuiteName,
		},
	}
}

func (s *TwoFactorSuite) TestTwoFactorOneTimePasswordRegistrationScenario() {
	suite.Run(s.T(), NewTwoFactorOneTimePasswordScenario())
}

func (s *TwoFactorSuite) TestTwoFactorWebAuthnRegistrationScenario() {
	suite.Run(s.T(), NewTwoFactorWebAuthnScenario())
}

func TestTwoFactorSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTwoFactorSuite())
}
