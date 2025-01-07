package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthBackendMySQLSuite struct {
	*RodSuite
}

func NewAuthBackendMySQLSuite() *AuthBackendMySQLSuite {
	return &AuthBackendMySQLSuite{
		RodSuite: NewRodSuite(authBackendMysqlSuiteName),
	}
}

func (s *AuthBackendMySQLSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *AuthBackendMySQLSuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func AuthBackendTestMySQLSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewAuthBackendMySQLSuite())
}
