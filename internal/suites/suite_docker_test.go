package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DockerSuite struct {
	*SeleniumSuite
}

func NewDockerSuite() *DockerSuite {
	return &DockerSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *DockerSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *DockerSuite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
}

func TestDockerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewDockerSuite())
}
