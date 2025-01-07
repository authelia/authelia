package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthBackendPostgresSuite struct {
	*RodSuite
}

func NewAuthBackendPostgresSuite() *AuthBackendPostgresSuite {
	return &AuthBackendPostgresSuite{
		RodSuite: NewRodSuite(authBackendPostgresSuiteName),
	}
}

func (s *AuthBackendPostgresSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *AuthBackendPostgresSuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func TestAuthBackendPostgresSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewAuthBackendPostgresSuite())
}
