package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthBackendSqliteWebDriverSuite struct {
	*RodSuite
}

func TestAuthBackendSqliteSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewAuthBackendSqliteSuite())
}

func NewAuthBackendSqliteWebDriverSuite() *AuthBackendSqliteWebDriverSuite {
	return &AuthBackendSqliteWebDriverSuite{
		RodSuite: NewRodSuite(authBackendSqliteSuiteName),
	}
}

func (s *AuthBackendSqliteWebDriverSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	browser, err := NewRodSession(RodSessionWithCredentials(s))

	s.Require().NoError(err, "Failed to start Rod session")

	s.RodSession = browser
}

func (s *AuthBackendSqliteWebDriverSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	s.Require().NoError(err, "Failed to stop Rod session")
}

func (s *AuthBackendSqliteWebDriverSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *AuthBackendSqliteWebDriverSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

type AuthBackendSqliteSuite struct {
	*BaseSuite
}

func NewAuthBackendSqliteSuite() *AuthBackendSqliteSuite {
	return &AuthBackendSqliteSuite{
		BaseSuite: &BaseSuite{
			Name: authBackendSqliteSuiteName,
		},
	}
}

func (s *AuthBackendSqliteSuite) TestAuthBackendSqliteWebDriverScenario() {
	suite.Run(s.T(), NewAuthBackendSqliteWebDriverSuite())
}

func (s *AuthBackendSqliteSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *AuthBackendSqliteSuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *AuthBackendSqliteSuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *AuthBackendSqliteSuite) TestCLIScenario() {
	composeFiles := defaultComposeFiles
	composeFiles = append(composeFiles,
		"internal/suites/AuthenticationBackendSQLite/docker-compose.yml",
	)

	dockerEnvironment := NewDockerEnvironment(composeFiles)
	suite.Run(s.T(), NewCLIScenario(s.Name, dockerEnvironment))
}
