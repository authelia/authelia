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

func TestDockerSuite(t *testing.T) {
	suite.Run(t, NewOneFactorScenario())
	suite.Run(t, NewTwoFactorScenario())
}
