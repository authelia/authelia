package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestAuthBackendMySQLSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewAuthBackendMySQLSuite())
}

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

func (s *AuthBackendMySQLSuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *AuthBackendMySQLSuite) TestCLIScenario() {
	composeFiles := defaultComposeFiles
	composeFiles = append(composeFiles,
		"internal/suites/AuthenticationBackendMySQL/docker-compose.yml",
		"internal/suites/example/compose/mysql/docker-compose.yml",
	)

	dockerEnvironment := NewDockerEnvironment(composeFiles)
	suite.Run(s.T(), NewCLIScenario(s.Name, dockerEnvironment))
}
