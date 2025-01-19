package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestAuthBackendPostgresSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewAuthBackendPostgresSuite())
}

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

func (s *AuthBackendPostgresSuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *AuthBackendPostgresSuite) TestCLIScenario() {
	composeFiles := defaultComposeFiles
	composeFiles = append(composeFiles,
		"internal/suites/AuthenticationBackendPostgres/docker-compose.yml",
		"internal/suites/example/compose/postgres/docker-compose.yml",
	)

	dockerEnvironment := NewDockerEnvironment(composeFiles)
	suite.Run(s.T(), NewCLIScenario(s.Name, dockerEnvironment))
}
