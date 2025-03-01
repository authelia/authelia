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
			Name: twoFactorSuiteName,
		},
	}
}

func (s *TwoFactorSuite) TestTwoFactorOneTimePasswordScenario() {
	suite.Run(s.T(), NewTwoFactorOneTimePasswordScenario())
}

func (s *TwoFactorSuite) TestTwoFactorWebAuthnScenario() {
	suite.Run(s.T(), NewTwoFactorWebAuthnScenario())
}

func (s *TwoFactorSuite) TestPasskeyScenario() {
	suite.Run(s.T(), NewPasskeyScenario())
}

func TestTwoFactorSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTwoFactorSuite())
}
