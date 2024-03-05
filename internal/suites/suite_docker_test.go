package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DockerSuite struct {
	*RodSuite
}

func NewDockerSuite() *DockerSuite {
	return &DockerSuite{
		RodSuite: NewRodSuite(dockerSuiteName),
	}
}

func (s *DockerSuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *DockerSuite) TestTwoFactorTOTPScenario() {
	suite.Run(s.T(), NewTwoFactorTOTPScenario())
}

func TestDockerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewDockerSuite())
}
