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

func (s *DockerSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
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
