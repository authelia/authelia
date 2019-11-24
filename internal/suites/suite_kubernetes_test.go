package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type KubernetesSuite struct {
	*SeleniumSuite
}

func NewKubernetesSuite() *KubernetesSuite {
	return &KubernetesSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestKubernetesSuite(t *testing.T) {
	suite.Run(t, NewOneFactorScenario())
	suite.Run(t, NewTwoFactorScenario())
}
